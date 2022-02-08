package main

import (
	"os"
	"testing"
)

// just make sure it can parse command line args correctly
func TestCmd(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		err := os.RemoveAll("tmp")
		if err != nil {
			t.Logf("remove tmp directory failed: %v", err)
		}
	}()
	// test command line parser only
	os.Args = []string{"", "-c", "json", "-o", "tmp/cmd.json", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/cmd.json"); f == nil {
		t.Error("command json failed")
	}
}
