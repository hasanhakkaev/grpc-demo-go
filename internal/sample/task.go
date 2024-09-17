package sample

import (
	pb "github.com/hasanhakkaev/yqapp-demo/proto/gen"
	"time"
)

func NewTask() *pb.Task {
	task := &pb.Task{
		Type:           uint32(randomInt(0, 9)),
		Value:          uint32(randomInt(0, 99)),
		State:          pb.TaskState_RECEIVED,
		CreationTime:   float32(time.Now().Unix()),
		LastUpdateTime: 0.0,
	}
	return task
}
