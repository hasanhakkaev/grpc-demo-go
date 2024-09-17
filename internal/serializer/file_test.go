package serializer_test

import (
	"github.com/hasanhakkaev/yqapp-demo/internal/sample"
	"github.com/hasanhakkaev/yqapp-demo/internal/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../../tmp/task.bin"
	jsonFile := "../../tmp/task.json"

	task1 := sample.NewTask()

	err := serializer.WriteProtobufToBinaryFile(task1, binaryFile)
	require.NoError(t, err)

	err = serializer.WriteProtobufToJSONFile(task1, jsonFile)
	require.NoError(t, err)

	task2 := sample.NewTask()

	err = serializer.ReadProtobufFromBinaryFile(binaryFile, task2)
	require.NoError(t, err)

	require.True(t, proto.Equal(task1, task2))

}
