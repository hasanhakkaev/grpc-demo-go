package domain

import (
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"math/rand"
	"time"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func RandomInt(min, max int) int {
	return min + rand.Int()%(max-min+1)
}

func MapGrpcStateToDomain(grpcState v1.TaskState) State {
	switch grpcState {
	case v1.TaskState_RECEIVED:
		return StateRECEIVED
	case v1.TaskState_PROCESSING:
		return StatePROCESSING
	case v1.TaskState_DONE:
		return StateDONE
	default:
		return ""
	}
}

func MapDomainStateToGrpc(domainState State) v1.TaskState {
	switch domainState {
	case StateRECEIVED:
		return v1.TaskState_RECEIVED
	case StatePROCESSING:
		return v1.TaskState_PROCESSING
	case StateDONE:
		return v1.TaskState_DONE
	default:
		return v1.TaskState_UNKNOWN
	}
}
