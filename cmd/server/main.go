package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"time"
)

const (
	secretKey = "Mahin-ya"
	duration  = 20 * time.Minute
)

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, fmt.Errorf("cannot load server cert: %w", err)
	}

	permClientCA, err := os.ReadFile("cert/ca-cert.pem")

	if err != nil {
		return nil, fmt.Errorf("cannot load client cert: %w", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(permClientCA) {
		return nil, fmt.Errorf("cannot add client cert to cert pool: %w", err)
	}

	cnf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(cnf), nil
}

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/grpc_tutorial.LaptopService/"
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func seedUser(userStore service.UserStore) error {
	err := createUser(userStore, "admin", "secret", "admin")
	if err != nil {
		return err
	}
	err = createUser(userStore, "user1", "secret", "user")
	if err != nil {
		return err
	}
	err = createUser(userStore, "user2", "secret", "user")

	return err

}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}
	err = userStore.Save(user)
	if err != nil {
		return err
	}
	return nil

}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("starting server on port %d", *port)
	imageStore := service.NewDiskImageStore("img")
	userStore := service.NewInMemoryUserStore()
	err := seedUser(userStore)
	if err != nil {
		log.Fatal("can not seed user: ", err)
	}
	jwtManager := service.NewJwtManager(secretKey, duration)

	authServer := service.NewAuthServer(userStore, jwtManager)
	ratingStore := service.NewInMemoryRatingStore()
	laptopStore := service.NewInMemoryLaptopStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("can not load TLS credentials: ", err)
	}
	interceptor := service.AuthInterceptor{
		JwtManager:      jwtManager,
		AccessibleRoles: accessibleRoles(),
	}
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.Unary(),
		),
		grpc.StreamInterceptor(interceptor.Stream()),
		grpc.Creds(tlsCredentials),
	)
	pb.RegisterAuthServiceServer(server, authServer)
	pb.RegisterLaptopServiceServer(server, laptopServer)
	reflection.Register(server)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal("can not start server port assignment error : ", err)
	}

	err = server.Serve(listener)
	if err != nil {
		log.Fatal("can not start server: ", err)
	}
}
