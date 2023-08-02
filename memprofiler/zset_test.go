package memprofiler

import "testing"

func TestRandLevel(t *testing.T) {
	size := 100000
	sum := 0
	for i := 0; i < size; i++ {
		sum += zsetRandomLevel()
	}
	t.Log(float64(sum) / float64(size))
}
