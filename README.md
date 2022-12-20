![license](https://img.shields.io/github/license/HDT3213/rdb)
[![Build Status](https://travis-ci.com/HDT3213/rdb.svg?branch=master)](https://app.travis-ci.com/github/HDT3213/rdb)
[![Coverage Status](https://coveralls.io/repos/github/HDT3213/rdb/badge.svg?branch=master)](https://coveralls.io/github/HDT3213/rdb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/HDT3213/rdb)](https://goreportcard.com/report/github.com/HDT3213/rdb)
[![Go Reference](https://pkg.go.dev/badge/github.com/hdt3213/rdb.svg)](https://pkg.go.dev/github.com/hdt3213/rdb)
<br>
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)

[中文版](https://github.com/HDT3213/rdb/blob/master/README_CN.md)

This is a golang implemented Redis RDB parser for secondary development and memory analysis.

It provides abilities to:

- Generate memory report for rdb file
- Convert RDB files to JSON
- Convert RDB files to Redis Serialization Protocol (or AOF file)
- Find the biggest N keys in RDB files
- Draw FlameGraph to analysis which kind of keys occupied most memory
- Customize data usage
- Generate RDB file

Support RDB version: 1 <= version <= 10(Redis 7.0)

If you read Chinese, you could find a thorough introduction to the RDB file format here: [Golang 实现 Redis(11): RDB 文件格式](https://www.cnblogs.com/Finley/p/16251360.html)

Thanks sripathikrishnan for his [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools)

# Install

If you have installed `go` on your compute, just simply use:

```
go install github.com/hdt3213/rdb@latest
```

### Package Managers

If you're a [Homebrew](https://brew.sh/) user, you can install [rdb](https://formulae.brew.sh/formula/rdb) via:

```sh
$ brew install rdb
```

Or, you can download executable binary file from [releases](https://github.com/HDT3213/rdb/releases) and put its path to
PATH environment.

use `rdb` command in terminal, you can see it's manual

```
This is a tool to parse Redis' RDB files
Options:
  -c command, including: json/memory/aof/bigkey/flamegraph
  -o output file path
  -n number of result, using in 
  -port listen port for flame graph web service
  -sep separator for flamegraph, rdb will separate key by it, default value is ":". 
                supporting multi separators: -sep sep1 -sep sep2 
  -regex using regex expression filter keys
  -no-expired filter expired keys

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
5. draw flamegraph
  rdb -c flamegraph [-port 16379] [-sep :] dump.rdb
```

# Convert to Json

Usage:

```
rdb -c json -o <output_path> <source_path>
```

example:

```
rdb -c json -o intset_16.json cases/intset_16.rdb
```

You can get some rdb examples in [cases](https://github.com/HDT3213/rdb/tree/master/cases)

The examples for json result:

```json
[
    {"db":0,"key":"hash","size":64,"type":"hash","hash":{"ca32mbn2k3tp41iu":"ca32mbn2k3tp41iu","mddbhxnzsbklyp8c":"mddbhxnzsbklyp8c"}},
    {"db":0,"key":"string","size":10,"type":"string","value":"aaaaaaa"},
    {"db":0,"key":"expiration","expiration":"2022-02-18T06:15:29.18+08:00","size":8,"type":"string","value":"zxcvb"},
    {"db":0,"key":"list","expiration":"2022-02-18T06:15:29.18+08:00","size":66,"type":"list","values":["7fbn7xhcnu","lmproj6c2e","e5lom29act","yy3ux925do"]},
    {"db":0,"key":"zset","expiration":"2022-02-18T06:15:29.18+08:00","size":57,"type":"zset","entries":[{"member":"zn4ejjo4ths63irg","score":1},{"member":"1ik4jifkg6olxf5n","score":2}]},
    {"db":0,"key":"set","expiration":"2022-02-18T06:15:29.18+08:00","size":39,"type":"set","members":["2hzm5rnmkmwb3zqd","tdje6bk22c6ddlrw"]}
]
```

# Generate Memory Report

RDB uses rdb encoded size to estimate redis memory usage.

```
rdb -c memory -o <output_path> <source_path>
```

Example:

```
rdb -c memory -o mem.csv cases/memory.rdb
```

The examples for csv result:

```csv
database,key,type,size,size_readable,element_count
0,hash,hash,64,64B,2
0,s,string,10,10B,0
0,e,string,8,8B,0
0,list,list,66,66B,4
0,zset,zset,57,57B,2
0,large,string,2056,2K,0
0,set,set,39,39B,2
```

# Flame Graph

In many cases there is not a few very large key but lots of small keys that occupied most memory.

RDB tool could separate keys by the given delimeters, then aggregate keys with same prefix.

Finally RDB tool presents the result as flame graph, with which you could find out which kind of keys consumed most
memory.

![截屏2022-10-30 12.06.00.png](https://s2.loli.net/2022/11/08/HW9ZxGfeEzArUhM.png)

In this example, the keys of pattern `Comment:*` use 8.463% memory.

Usage:

```
rdb -c flamegraph [-port <port>] [-sep <separator1>] [-sep <separator2>] <source_path>
```

Example:

```
rdb -c flamegraph -port 16379 -sep : dump.rdb
```

# Find The Biggest Keys

RDB can find biggest N keys in file

```
rdb -c bigkey -n <result_number> <source_path>
```

Example:

```
rdb -c bigkey -n 5 cases/memory.rdb
```

The examples for csv result:

```csv
database,key,type,size,size_readable,element_count
0,large,string,2056,2K,0
0,list,list,66,66B,4
0,hash,hash,64,64B,2
0,zset,zset,57,57B,2
0,set,set,39,39B,2
```

# Convert to AOF

Usage:

```
rdb -c aof -o <output_path> <source_path>
```

Example:

```
rdb -c aof -o mem.aof cases/memory.rdb
```

The examples for aof result:

```
*3
$3
SET
$1
s
$7
aaaaaaa
```

# Regex Filter

RDB tool supports using regex expression to filter keys.

Example:
```rdb
rdb -c json -o regex.json -regex '^l.*' cases/memory.rdb
```

# Customize data usage

```go
package main

import (
	"github.com/hdt3213/rdb/parser"
	"os"
)

func main() {
	rdbFile, err := os.Open("dump.rdb")
	if err != nil {
		panic("open dump.rdb failed")
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	decoder := parser.NewDecoder(rdbFile)
	err = decoder.Parse(func(o parser.RedisObject) bool {
		switch o.GetType() {
		case parser.StringType:
			str := o.(*parser.StringObject)
			println(str.Key, str.Value)
		case parser.ListType:
			list := o.(*parser.ListObject)
			println(list.Key, list.Values)
		case parser.HashType:
			hash := o.(*parser.HashObject)
			println(hash.Key, hash.Hash)
		case parser.ZSetType:
			zset := o.(*parser.ZSetObject)
			println(zset.Key, zset.Entries)
		}
		// return true to continue, return false to stop the iteration
		return true
	})
	if err != nil {
		panic(err)
	}
}
```

# Generate RDB file

This library can generate RDB file: 

```go
package main

import (
	"github.com/hdt3213/rdb/encoder"
	"github.com/hdt3213/rdb/model"
	"os"
	"time"
)

func main() {
	rdbFile, err := os.Create("dump.rdb")
	if err != nil {
		panic(err)
	}
	defer rdbFile.Close()
	enc := encoder.NewEncoder(rdbFile)
	err = enc.WriteHeader()
	if err != nil {
		panic(err)
	}
	auxMap := map[string]string{
		"redis-ver":    "4.0.6",
		"redis-bits":   "64",
		"aof-preamble": "0",
	}
	for k, v := range auxMap {
		err = enc.WriteAux(k, v)
		if err != nil {
			panic(err)
		}
	}

	err = enc.WriteDBHeader(0, 5, 1)
	if err != nil {
		panic(err)
	}
	expirationMs := uint64(time.Now().Add(time.Hour*8).Unix() * 1000)
	err = enc.WriteStringObject("hello", []byte("world"), encoder.WithTTL(expirationMs))
	if err != nil {
		panic(err)
	}
	err = enc.WriteListObject("list", [][]byte{
		[]byte("123"),
		[]byte("abc"),
		[]byte("la la la"),
	})
	if err != nil {
		panic(err)
	}
	err = enc.WriteSetObject("set", [][]byte{
		[]byte("123"),
		[]byte("abc"),
		[]byte("la la la"),
	})
	if err != nil {
		panic(err)
	}
	err = enc.WriteHashMapObject("list", map[string][]byte{
		"1":  []byte("123"),
		"a":  []byte("abc"),
		"la": []byte("la la la"),
	})
	if err != nil {
		panic(err)
	}
	err = enc.WriteZSetObject("list", []*model.ZSetEntry{
		{
			Score: 1.234,
			Member: "a",
		},
		{
			Score: 2.71828,
			Member: "b",
		},
	})
	if err != nil {
		panic(err)
	}
	err = enc.WriteEnd()
	if err != nil {
		panic(err)
	}
}
```

# Benchmark

Tested on MacBook Pro (16-inch, 2019) 2.6 GHz 6cores Intel Core i7, using  a 1.3 GB RDB file encoded with v9 format from Redis 5.0 in production environment.

|usage|elapsed|speed|
|:-:|:-:|:-:|
|ToJson|144.11s|9.23MB/s|
|Memory|18.585s|71.62MB/s|
|AOF|104.77s|12.76MB/s|
|Top10|14.8s|89.95MB/s|
|FlameGraph|49.38s|26.96MB/s|
