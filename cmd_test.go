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
	os.Args = []string{"", "-c", "memory", "-o", "tmp/memory_size.csv", "-size", "100~1024KB", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/memory_size.csv"); f == nil {
		t.Error("command memory with size failed")
	}

	os.Args = []string{"", "-c", "json", "-o", "tmp/cmd.json", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/cmd.json"); f == nil {
		t.Error("command json failed")
	}
	os.Args = []string{"", "-c", "memory", "-o", "tmp/memory.csv", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/memory.csv"); f == nil {
		t.Error("command memory failed")
	}
	os.Args = []string{"", "-c", "aof", "-o", "tmp/memory.aof", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/memory.aof"); f == nil {
		t.Error("command memory failed")
	}
	os.Args = []string{"", "-c", "bigkey", "-o", "tmp/bigkey.csv", "-n", "10", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/bigkey.csv"); f == nil {
		t.Error("command bigkey failed")
	}
	os.Args = []string{"", "-c", "bigkey", "-n", "10", "cases/memory.rdb"}
	main()

	os.Args = []string{"", "-c", "memory", "-o", "tmp/memory_regex.csv", "-regex", "^l.*", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/memory_regex.csv"); f == nil {
		t.Error("command memory failed")
	}

	os.Args = []string{"", "-c", "memory", "-o", "tmp/memory_regex.csv", "-regex", "^l.*", "-no-expired", "cases/memory.rdb"}
	main()
	if f, _ := os.Stat("tmp/memory_regex.csv"); f == nil {
		t.Error("command memory failed")
	}
	os.Args = []string{"", "-c", "prefix", "-o", "tmp/tree.csv", "cases/tree.rdb"}
	main()
	if f, _ := os.Stat("tmp/tree.csv"); f == nil {
		t.Error("command prefix failed")
	}

	// test error command line
	os.Args = []string{"", "-c", "json", "-o", "tmp/output", "/none/a"}
	main()
	os.Args = []string{"", "-c", "aof", "-o", "tmp/output", "/none/a"}
	main()
	os.Args = []string{"", "-c", "memory", "-o", "tmp/output", "/none/a"}
	main()
	os.Args = []string{"", "-c", "bigkey", "-o", "tmp/output", "/none/a"}
	main()

	os.Args = []string{"", "-c", "bigkey", "-o", "/none/a", "-n", "10", "cases/memory.rdb"}
	main()
	os.Args = []string{"", "-c", "aof", "-o", "/none/a", "cases/memory.rdb"}
	main()
	os.Args = []string{"", "-c", "memory", "-o", "/none/a", "cases/memory.rdb"}
	main()
	os.Args = []string{"", "-c", "json", "-o", "/none/a", "cases/memory.rdb"}
	main()

	os.Args = []string{"", "-c", "bigkey", "-o", "tmp/bigkey.csv", "cases/memory.rdb"}
	main()
	os.Args = []string{"", "-c", "none", "-o", "tmp/memory.aof", "cases/memory.rdb"}
	main()
	os.Args = []string{""}
	main()
	os.Args = []string{"", "-c", "aof"}
	main()
}
