package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/sample"
	"github.com/UnplugCharger/grpc_tutorial/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()
	store := NewInMemoryLaptopStore()
	address := startTestLaptopServer(t, store, nil, nil)
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

	address := startTestLaptopServer(t, store, nil, nil)
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

func TestClientUploadImage(t *testing.T) {
	t.Parallel()
	testImageFolder := "../tmp"

	laptopStore := NewInMemoryLaptopStore()
	imageStore := NewDiskImageStore(testImageFolder)

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	address := startTestLaptopServer(t, laptopStore, imageStore, nil)
	laptopClient := newTestLaptopClient(t, address)

	imagePath := testImageFolder + "/laptop.jpg"
	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer func(file *os.File) {
		err := file.Close()
		require.NoError(t, err)
	}(file)

	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)

	imageType := filepath.Ext(imagePath)
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_ImageInfo{
			ImageInfo: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: imageType,
			},
		},
	}

	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		size += n
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		require.NoError(t, err)

	}

	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, res.GetImageId())
	require.EqualValues(t, size, res.GetSize(), "size of image uploaded is not the same as original image")

	savedImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetImageId(), imageType)
	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))

}

func TestClientRateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := NewInMemoryLaptopStore()
	ratingStore := NewInMemoryRatingStore()

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	address := startTestLaptopServer(t, laptopStore, nil, ratingStore)
	laptopClient := newTestLaptopClient(t, address)

	stream, err := laptopClient.RateLaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	averages := []float64{8, 7.75, 8.5}
	expectedCount := []int32{1, 2, 3}

	n := len(scores)
	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}
		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; idx < n; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			return
		}

		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLaptopId())
		require.Equal(t, uint32(expectedCount[idx]), res.GetRatedCount())
		require.Equal(t, averages[idx], res.GetAverageScore())
	}

}

func startTestLaptopServer(t *testing.T, laptopStore LaptopStore, imageSore ImageStore, ratingStore RatingStore) string {
	laptopServer := NewLaptopServer(laptopStore, imageSore, ratingStore)
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
