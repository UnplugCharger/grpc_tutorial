package client

import (
	"context"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"time"
)

type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

func NewAuthClient(cc *grpc.ClientConn, username string, password string) *AuthClient {
	return &AuthClient{
		service:  pb.NewAuthServiceClient(cc),
		username: username,
		password: password,
	}
}

func (client *AuthClient) LogIn() (string, error) {
	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create login request
	req := &pb.LogInRequest{
		Username: client.username,
		Password: client.password,
	}

	// send request to server
	res, err := client.service.LogIn(ctx, req)
	if err != nil {
		return "", status.Errorf(status.Code(err), "cannot login: %v", err)
	}

	return res.Token, nil

}
