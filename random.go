package alimns

import (
	"math/rand"
)

func randomIntInRange(min, max int) int {
	return rand.Intn(max-min) + min
}
