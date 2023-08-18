package main

import (
	"context"
	"flag"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	address := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dialing server on address  %s", *address)

	dial, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("can not dial server: ", err)
	}

	client := pb.NewLaptopServiceClient(dial)

	for i := 0; i < 10; i++ {
		createRandomLaptop(err, client)

	}

	filter := &pb.Filter{
		MaxPrice:    3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinMemory:   &pb.Memory{Value: 8, Unit: pb.Memory_GB},
	}

	searchLaptop(client, filter)
}

func createRandomLaptop(err error, client pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	laptop.Id = ""
	req := &pb.CreateLaptopRequest{

		Laptop: laptop,
	}

	// set time out for the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.CreateLaptop(ctx, req)
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

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("searching for laptop..", filter)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}

	stream, err := laptopClient.SearchLaptop(ctx, req)
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
