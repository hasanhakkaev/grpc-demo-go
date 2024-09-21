package sample

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func randomInt(min, max int) int {
	return min + rand.Int()%(max-min+1)
}

func randomID() uint32 { return uuid.New().ID() }
