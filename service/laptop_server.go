package service

import (
	"context"
	"errors"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type LaptopServer struct {
	Store LaptopStore
	pb.UnimplementedLaptopServiceServer
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{store, pb.UnimplementedLaptopServiceServer{}}

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
	err := server.Store.Save(laptop)
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

	err := server.Store.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
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
