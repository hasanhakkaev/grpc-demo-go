package domain

import (
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/serializer"
)

var _ serializer.API[*v1.Task] = (*Task)(nil)

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

func (t *Task) API() *v1.Task {
	return &v1.Task{
		Type:  t.Type,
		Value: t.Value,
		State: MapDomainStateToGrpc(t.State),
	}
}

func (t *Task) FromAPI(in *v1.Task) {
	t.Type = in.GetType()
	t.Value = in.GetValue()
	t.State = MapGrpcStateToDomain(in.GetState())

}
