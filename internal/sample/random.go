package sample

import (
	"math/rand"
	"time"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func randomInt(min, max int) int {
	return min + rand.Int()%(max-min+1)
}
