package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hdt3213/rdb/helper"
)

const help = `
This is a tool to parse Redis' RDB files
Options:
  -c command, including: json/memory/aof/bigkey/prefix/flamegraph
  -o output file path
  -n number of result, using in command: bigkey/prefix
  -port listen port for flame graph web service
  -sep separator for flamegraph, rdb will separate key by it, default value is ":". 
		supporting multi separators: -sep sep1 -sep sep2 
  -regex using regex expression filter keys
  -expire filter keys by its expiration time
		1. '1751731200~1751817600' get keys with expiration time in range [1751731200, 1751817600]
		2. '1751731200~now' 'now~1751731200' magic variable 'now' represents the current timestamp
		3. '1751731200~inf' 'now~inf' magic variable 'inf' represents the Infinity
		4. 'noexpire' get keys without expiration time
		5. 'anyexpire' get all keys with expiration time
  -no-expired filter expired keys(deprecated, please use 'expire' option)
  -concurrent The number of concurrent json converters. 4 by default.

Examples:
parameters between '[' and ']' is optional
1. convert rdb to json
  rdb -c json -o dump.json dump.rdb
2. generate memory report
  rdb -c memory -o memory.csv dump.rdb
3. convert to aof file
  rdb -c aof -o dump.aof dump.rdb
4. get largest keys
  rdb -c bigkey [-o dump.aof] [-n 10] dump.rdb
5. get number and memory size by prefix
  rdb -c prefix [-n 10] [-max-depth 3] [-o prefix-report.csv] dump.rdb
6. draw flamegraph
  rdb -c flamegraph [-port 16379] [-sep :] dump.rdb
`

type separators []string

func (s *separators) String() string {
	return strings.Join(*s, " ")
}

func (s *separators) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var cmd string
	var output string
	var n int
	var port int
	var seps separators
	var regexExpr string
	var noExpired bool
	var expirationExpr string
	var maxDepth int
	var concurrent int
	var keySize int
	var err error
	flagSet.StringVar(&cmd, "c", "", "command for rdb: json")
	flagSet.StringVar(&output, "o", "", "output file path")
	flagSet.IntVar(&n, "n", 0, "")
	flagSet.IntVar(&maxDepth, "max-depth", 0, "max depth of prefix tree")
	flagSet.IntVar(&port, "port", 0, "listen port for web")
	flagSet.IntVar(&concurrent, "concurrent", 0, "concurrent number for json converter")
	flagSet.Var(&seps, "sep", "separator for flame graph")
	flagSet.StringVar(&regexExpr, "regex", "", "regex expression")
	flagSet.StringVar(&expirationExpr, "expire", "", "expiration filter expression")
	flagSet.BoolVar(&noExpired, "no-expired", false, "filter expired keys(deprecated, please use expire)")
	flagSet.IntVar(&keySize, "key-size", 0, "filter keys by its size")
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

	var options []interface{}
	if regexExpr != "" {
		options = append(options, helper.WithRegexOption(regexExpr))
	}
	if noExpired {
		options = append(options, helper.WithNoExpiredOption())
	}
	if concurrent != 0 {
		options = append(options, helper.WithConcurrent(concurrent))
	}
	if expirationExpr != "" {
		options = append(options, helper.WithExpirationOption(expirationExpr))
	}
	if keySize >= 0 {
		options = append(options, helper.WithKeySizeFilterOption(keySize))
	}

	var outputFile *os.File
	if output == "" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(output)
		if err != nil {
			fmt.Printf("open output faild: %v", err)
		}
		defer func() {
			_ = outputFile.Close()
		}()
	}

	switch cmd {
	case "json":
		err = helper.ToJsons(src, output, options...)
	case "memory":
		err = helper.MemoryProfile(src, output, options...)
	case "aof":
		err = helper.ToAOF(src, output, options)
	case "bigkey":
		err = helper.FindBiggestKeys(src, n, outputFile, options...)
	case "prefix":
		err = helper.PrefixAnalyse(src, n, maxDepth, outputFile, options...)
	case "flamegraph":
		_, err = helper.FlameGraph(src, port, seps, options...)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
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
