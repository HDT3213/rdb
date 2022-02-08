package main

import (
	"flag"
	"fmt"
	"github.com/hdt3213/rdb/helper"
)

const help = `
This is a tool to parse Redis' RDB files
Options:
  -c command, including: json/memory/aof
  -o output file path

Examples:
1. convert rdb to json
  rdb -c json -o dump.json dump.rdb
2. generate memory report
  rdb -c memory -o memory.csv dump.rdb
3. convert to aof file
  rdb -c aof -o dump.aof dump.rdb
`

func main() {
	var cmd string
	var output string
	flag.StringVar(&cmd, "c", "", "command for rdb: json")
	flag.StringVar(&output, "o", "", "output file path")
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
	if output == "" {
		println("output file path is required")
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
	default:
		println("unknown command")
		return
	}
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
