package raft

import (
	"log"
	"math/rand"
	"time"
)

// Debugging
const Debug = 0

var start bool = true

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if format[len(format)-1] != '\n' {
		format += "\n"
	}

	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

// RangeInt returns an int in range [low, high)
func RangeInt(low int, high int) int {
	if start {
		start = false
		rand.Seed(time.Now().UnixNano())
	}

	return low + rand.Intn(high-low)
}

func Min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}
