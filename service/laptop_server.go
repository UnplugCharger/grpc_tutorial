package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

const maxImageSize = 1 << 20

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	LaptopStore LaptopStore
	ImageStore  ImageStore
	RatingStore RatingStore
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{LaptopStore: laptopStore, ImageStore: imageStore, RatingStore: ratingStore}

}

func (server *LaptopServer) CreateLaptop(ctx context.Context, in *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := in.GetLaptop()
	log.Printf("received a create-laptop request  with id : %s", laptop.Id)
	// If the ID is empty, generate a new UUID.
	if len(laptop.Id) == 0 {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v ", err)
		}
		laptop.Id = id.String()
		log.Printf("generated a new laptop id: %s", laptop.Id)
	} else {
		// If the ID is not empty, check if it's a valid UUID.
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "provided laptop UUID is not valid: %v", err)
		}
	}

	//// some heavy processing
	//time.Sleep(6 * time.Second)

	if errors.Is(ctx.Err(), context.Canceled) {
		log.Print("request cancelled")
		return nil, status.Error(codes.Canceled, "request is cancelled")
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		log.Print("deadline exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	// save the laptop to the store
	err := server.LaptopStore.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "failed to save laptop to the store : %v", err)
	}

	log.Printf("laptop  with id: %s saved to the store ", laptop.Id)

	resp := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return resp, nil

}

func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("received a search-laptop request with filter: %v", filter)

	err := server.LaptopStore.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
		res := &pb.SearchLaptopResponse{
			Laptop: laptop,
		}

		err := stream.Send(res)
		if err != nil {
			return err
		}

		log.Printf("sent laptop with id: %s", laptop.Id)
		return nil
	})
	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	log.Print("received an upload-image request")

	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot receive image info: %v", err)
	}

	laptopID := req.GetImageInfo().GetLaptopId()
	imageType := req.GetImageInfo().GetImageType()
	log.Printf("received an upload-image request for laptop %s with image type %s", laptopID, imageType)

	laptop, err := server.LaptopStore.Find(laptopID)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot find laptop: %v", err)
	}

	if laptop == nil {
		return status.Errorf(codes.InvalidArgument, "laptop %s doesn't exist", laptopID)
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		log.Print("waiting to receive more data")
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data to receive ")
			break
		}

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Print("client cancelled streaming")
				return status.Errorf(codes.Canceled, "client cancelled streaming")
			}
			log.Printf("cannot receive chunk data: %v", err)
			return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		imageSize += size
		log.Printf("received a chunk with size: %d", size)
		if imageSize > maxImageSize {
			return status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize)
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}

	}

	imageID, err := server.ImageStore.Save(laptopID, imageType, imageData.Bytes())
	if err != nil {
		return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
	}

	res := &pb.UploadImageResponse{
		ImageId: imageID,
		Size:    uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}

	log.Printf("image with id: %s, size: %d saved to the store", imageID, imageSize)
	return nil
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}

		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream: %v", err)
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("received a rate-laptop request: id = %s, score = %.2f", laptopID, score)

		found, err := server.LaptopStore.Find(laptopID)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot find laptop: %v", err)
		}

		if found == nil {
			return status.Errorf(codes.NotFound, "laptop %s doesn't exist", laptopID)
		}

		rating, err := server.RatingStore.Add(laptopID, score)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot add rating to the store: %v", err)
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot send response: %v", err)
		}
	}
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Error(codes.Canceled, "request is cancelled")
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}
