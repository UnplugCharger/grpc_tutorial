package serializer

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"os"
)

func WriteProtobufToBinFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("could not encode message: %v", err)
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("could not write file: %v", err)
	}
	return nil
}

func ReadProtobufFromBinFile(filename string, message proto.Message) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("could not decode message: %v", err)
	}
	return nil
}

func WriteProtobufToJsonFile(message proto.Message, filename string) error {
	return nil
}
