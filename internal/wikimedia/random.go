package wikimedia

import (
	"math/rand"
	"time"
)

func init() {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(seed)
}

var rnd *rand.Rand

func getInt(max int) int {
	return rnd.Intn(max)
}
