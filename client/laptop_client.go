package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"time"
)

type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(cc *grpc.ClientConn) LaptopClient {
	return LaptopClient{service: pb.NewLaptopServiceClient(cc)}
}

func (client *LaptopClient) CreateLaptop(laptop *pb.Laptop) {
	req := &pb.CreateLaptopRequest{

		Laptop: laptop,
	}

	// set time out for the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Printf("laptop already exists")

		} else {
			log.Fatal("can not create laptop: ", err)
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}

func (client *LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Print("searching for laptop..", filter)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}

	stream, err := client.service.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("can not search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			log.Fatal("can not receive response: ", err)
		}

		log.Print("received: ", res.GetLaptop())
	}
}

func (client *LaptopClient) UploadImage(laptopId string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("can not open image file: ", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal("can not close file: ", err)
		}
	}(file)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("can not upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_ImageInfo{
			ImageInfo: &pb.ImageInfo{
				LaptopId:  laptopId,
				ImageType: ".jpg",
			},
		},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("can not send image info to server: ", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("can not read chunk to buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("can not send chunk to server: ", err)
		}

	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("can not receive response from server: ", err)
	}
	log.Printf("image uploaded with id: %s, size: %d", res.GetImageId(), res.GetSize())

}

func (client *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("can not rate laptop: %v", err)
	}

	waitResponse := make(chan error)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more response")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("can not receive stream response: %v", err)
				return
			}
			log.Printf("received response: %v", res)
		}
	}()

	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}
		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("can not send request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Printf("send request: %v", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("can not close send: %v", err)
	}

	err = <-waitResponse
	return err

}
