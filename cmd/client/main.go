package main

import (
	"flag"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/client"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strings"
	"time"
)

const (
	username      = "admin"
	password      = "secret"
	tokenDuration = time.Second * 10
)

func authMethods() map[string]bool {
	const laptopServicePath = "/grpc_tutorial.LaptopService/"
	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func main() {
	address := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dialing server on address  %s", *address)

	dial, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("can not dial server: ", err)
	}

	authClient := client.NewAuthClient(dial, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), tokenDuration)
	if err != nil {
		log.Fatal("can not create auth interceptor: ", err)
	}

	dial2, err := grpc.Dial(*address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("can not dial server: ", err)
	}

	laptopClient := client.NewLaptopClient(dial2)
	testSearchLaptop(laptopClient)
	testCreateLaptop(laptopClient)
	testUploadImage(laptopClient)
	testRateLaptop(laptopClient)

}

func testCreateLaptop(client client.LaptopClient) {
	laptop := sample.NewLaptop()
	client.CreateLaptop(laptop)
}

func testSearchLaptop(client client.LaptopClient) {
	for i := 0; i < 3; i++ {
		client.CreateLaptop(sample.NewLaptop())

	}

	filter := &pb.Filter{
		MaxPrice:    3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinMemory:   &pb.Memory{Value: 8, Unit: pb.Memory_GB},
	}

	client.SearchLaptop(filter)
}

func testRateLaptop(client client.LaptopClient) {
	n := 5

	laptopIDs := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		client.CreateLaptop(laptop)
		laptopIDs[i] = laptop.GetId()
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := client.RateLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal("can not rate laptop: ", err)
		}

	}
}

func testUploadImage(client client.LaptopClient) {
	laptop := sample.NewLaptop()
	client.CreateLaptop(laptop)
	client.UploadImage(laptop.GetId(), "tmp/laptop.jpg")

}
