package sample

import (
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"time"
)

func NewTask() *v1.Task {
	task := &v1.Task{
		Id:             randomID(),
		Type:           uint32(randomInt(0, 9)),
		Value:          uint32(randomInt(0, 99)),
		State:          v1.TaskState_RECEIVED,
		CreationTime:   float32(time.Now().Unix()),
		LastUpdateTime: 0.0,
	}
	return task
}
