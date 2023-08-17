package serializer

import (
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"github.com/UnplugCharger/grpc_tutorial/sample"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriteProtobufToFile(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/test.bin"
	//jsonFile := "../test/test.json"

	laptop1 := sample.NewLaptop()
	err := WriteProtobufToBinFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = ReadProtobufFromBinFile(binaryFile, laptop2)
	require.NoError(t, err)

	require.True(t, proto.Equal(laptop1, laptop2))
}

func TestReadProtobufFromBinFile(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/test.bin"
	//jsonFile := "../test/test.json"

	laptop1 := sample.NewLaptop()
	err := WriteProtobufToBinFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}
	err = ReadProtobufFromBinFile(binaryFile, laptop2)
	require.NoError(t, err)

	require.True(t, proto.Equal(laptop1, laptop2))
}
