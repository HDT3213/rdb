package main

import (
	"flag"
	"fmt"
	"github.com/hdt3213/rdb/helper"
	"os"
)

const help = `
This is a tool to parse Redis' RDB files
Options:
  -c command, including: json/memory/aof
  -o output file path
  -n number of result 

Examples:
1. convert rdb to json
  rdb -c json -o dump.json dump.rdb
2. generate memory report
  rdb -c memory -o memory.csv dump.rdb
3. convert to aof file
  rdb -c aof -o dump.aof dump.rdb
4. get largest keys
  rdb -c bigkey -o dump.aof dump.rdb
`

func main() {
	var cmd string
	var output string
	var n int
	flag.StringVar(&cmd, "c", "", "command for rdb: json")
	flag.StringVar(&output, "o", "", "output file path")
	flag.IntVar(&n, "n", 0, "")
	flag.Parse()
	src := flag.Arg(0)

	if cmd == "" {
		println(help)
		return
	}
	if src == "" {
		println("src file is required")
		return
	}

	var err error
	switch cmd {
	case "json":
		err = helper.ToJsons(src, output)
	case "memory":
		err = helper.MemoryProfile(src, output)
	case "aof":
		err = helper.ToAOF(src, output)
	case "bigkey":
		err = helper.FindBiggestKeys(src, n, os.Stdout)
	default:
		println("unknown command")
		return
	}
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
