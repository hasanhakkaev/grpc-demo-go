package domain

import (
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
)

//var _ serializer.API[*v1.Task] = (*Task)(nil)

type State string

const (
	StateRECEIVED   State = "RECEIVED"
	StatePROCESSING State = "PROCESSING"
	StateDONE       State = "DONE"
)

type Task struct {
	ID             uint32
	Type           uint32
	Value          uint32
	State          State
	CreationTime   float64
	LastUpdateTime float64
}

// ToTaskCreateParams converts this v1.Task to a database.CreateTaskParams.
func (t *Task) ToTaskCreateParams() *database.CreateTaskParams {

	return &database.CreateTaskParams{
		Type:           t.Type,
		Value:          t.Value,
		State:          database.State(t.State),
		CreationTime:   t.CreationTime,
		LastUpdateTime: t.LastUpdateTime,
	}
}

func (t *Task) ToTaskUpdateParams() *database.UpdateTaskStateParams {
	return &database.UpdateTaskStateParams{
		State:          database.State(t.State),
		LastUpdateTime: t.LastUpdateTime,
		ID:             int32(t.ID),
	}
}

func FromProtoToDomain(pbTask *v1.Task) *Task {
	return &Task{
		Type:  pbTask.GetType(),
		Value: pbTask.GetValue(),
		State: State(pbTask.GetState()),
	}
}

func FromDBToDomain(dbTask *database.Task) *Task {
	return &Task{
		ID:             uint32(dbTask.ID),
		Type:           dbTask.Type,
		Value:          dbTask.Value,
		State:          State(dbTask.State),
		CreationTime:   dbTask.CreationTime,
		LastUpdateTime: dbTask.LastUpdateTime,
	}
}

func FromDomainToProto(task *Task) *v1.Task {
	return &v1.Task{
		Id:    task.ID,
		Type:  task.Type,
		Value: task.Value,
		State: MapDomainStateToGrpc(task.State),
	}
}

func RandomTask() *Task {
	return &Task{
		ID:    0,
		Type:  uint32(RandomInt(0, 9)),
		Value: uint32(RandomInt(0, 99)),
		State: StateRECEIVED,
	}
}
