package service

import (
	"context"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/sample"
	"github.com/UnplugCharger/grpc_tutorial/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()
	store := NewInMemoryLaptopStore()
	address := startTestLaptopServer(t, store, nil)
	laptopClient := newTestLaptopClient(t, address)

	laptop := sample.NewLaptop()
	expectedId := laptop.Id

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedId, res.Id)

	// Check if the laptop is saved to the store.
	other, err := store.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	requireSameSample(t, laptop, other)
}
func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPrice:    2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinMemory:   &pb.Memory{Value: 8, Unit: pb.Memory_GB},
	}
	store := NewInMemoryLaptopStore()
	expectedIds := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.Price = 1000
			laptop.Cpu.NumberOfCores = 2
		case 1:
			laptop.Price = 2000
			laptop.Cpu.NumberOfCores = 2
		case 2:
			laptop.Price = 3000
			laptop.Cpu.NumberOfCores = 4
		case 3:
			laptop.Price = 4000
			laptop.Cpu.NumberOfCores = 4
		case 4:
			laptop.Price = 5000
			laptop.Cpu.NumberOfCores = 8
		case 5:
			laptop.Price = 6000
			laptop.Cpu.NumberOfCores = 8
		}
		err := store.Save(laptop)
		require.NoError(t, err)
	}

	address := startTestLaptopServer(t, store, nil)
	laptopClient := newTestLaptopClient(t, address)

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err != nil {
			break
		}
		found++
		require.True(t, expectedIds[res.GetLaptop().GetId()])
	}
	require.Equal(t, len(expectedIds), found)

}

func startTestLaptopServer(t *testing.T, laptopStore LaptopStore, imageSore ImageStore) string {
	laptopServer := NewLaptopServer(laptopStore, imageSore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String()

}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)

}

func requireSameSample(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
