package serializer

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// ProtobufToJSON converts a protobuf message to JSON string
func ProtobufToJSON(message proto.Message) (string, error) {

	marshaller := jsonpb.Marshaler{
		Indent:       " ",
		EnumsAsInts:  false,
		EmitDefaults: true,
	}

	return marshaller.MarshalToString(message)

}
