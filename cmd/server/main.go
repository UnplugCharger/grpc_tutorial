package main

import (
	"flag"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("starting server on port %d", *port)

	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
	server := grpc.NewServer()
	pb.RegisterLaptopServiceServer(server, laptopServer)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal("can not start server port assignment error : ", err)
	}

	err = server.Serve(listener)
	if err != nil {
		log.Fatal("can not start server: ", err)
	}
}
