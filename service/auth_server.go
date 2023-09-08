package service

import (
	"context"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	userStore  UserStore
	jwtManager *JwtManager
	pb.UnimplementedAuthServiceServer
}

func NewAuthServer(userStore UserStore, jwtManager *JwtManager) *AuthServer {
	return &AuthServer{userStore: userStore, jwtManager: jwtManager}
}

func (server *AuthServer) LogIn(ctx context.Context, in *pb.LogInRequest) (*pb.LogInResponse, error) {
	user, err := server.userStore.Find(in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.CheckPassword(in.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token: %v", err)
	}

	return &pb.LogInResponse{Token: token}, nil
}
