package serializer

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtobufToJSON converts protocol buffer message to JSON string
func ProtobufToJSON(message proto.Message) (string, error) {
	mOptions := protojson.MarshalOptions{
		EmitDefaultValues: true,
		Indent:            "  ",
		UseEnumNumbers:    false,
	}

	return mOptions.Format(message), nil
}
