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
  -port listen port for flame graph web service
  -sep separator for flamegraph, rdb will separate key by it, default value is ":"

Examples:
1. convert rdb to json
  rdb -c json -o dump.json dump.rdb
2. generate memory report
  rdb -c memory -o memory.csv dump.rdb
3. convert to aof file
  rdb -c aof -o dump.aof dump.rdb
4. get largest keys
  rdb -c bigkey -o dump.aof dump.rdb
5. draw flamegraph
  rdb -c flamegraph -port 16379 -sep : dump.rdb
`

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var cmd string
	var output string
	var n int
	var port int
	var separator string
	flagSet.StringVar(&cmd, "c", "", "command for rdb: json")
	flagSet.StringVar(&output, "o", "", "output file path")
	flagSet.IntVar(&n, "n", 0, "")
	flagSet.IntVar(&port, "port", 0, "listen port for web")
	flagSet.StringVar(&separator, "sep", "", "")
	_ = flagSet.Parse(os.Args[1:]) // ExitOnError
	src := flagSet.Arg(0)

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
		if output == "" {
			err = helper.FindBiggestKeys(src, n, os.Stdout)
		} else {
			outputFile, err := os.Create(output)
			if err != nil {
				fmt.Printf("open output faild: %v", err)
			}
			defer func() {
				_ = outputFile.Close()
			}()
			err = helper.FindBiggestKeys(src, n, outputFile)
		}
	case "flamegraph":
		_, err = helper.FlameGraph(src, port, separator)
		<-make(chan struct{})
	default:
		println("unknown command")
		return
	}
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
