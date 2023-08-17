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
}

func NewLaptopService() *LaptopServer {
	return &LaptopServer{}
}

func (server *LaptopServer) CreateLaptop(ctx context.Context, in *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := in.GetLaptop()
	log.Printf("received a create-laptop request  with id : %s", laptop.Id)
	if len(laptop.Id) > 0 {
		// check if its a valid uuid
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "provided laptop uuid is not valid: %v", err)
		} else {
			id, err := uuid.NewRandom()
			if err != nil {
				return nil, status.Errorf(codes.Internal, "can not generate a new laptop ID: %v ", err)
			}
			laptop.Id = id.String()

		}

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
