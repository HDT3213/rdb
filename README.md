![license](https://img.shields.io/github/license/HDT3213/rdb)
![download](https://img.shields.io/github/downloads/hdt3213/rdb/total)
[![Go Reference](https://pkg.go.dev/badge/github.com/hdt3213/rdb.svg)](https://pkg.go.dev/github.com/hdt3213/rdb)
<br>
[![Build Status](https://github.com/hdt3213/rdb/actions/workflows/main.yml/badge.svg)](https://github.com/HDT3213/rdb/actions?query=branch%3Amaster)
[![Coverage Status](https://coveralls.io/repos/github/HDT3213/rdb/badge.svg?branch=master)](https://coveralls.io/github/HDT3213/rdb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/HDT3213/rdb)](https://goreportcard.com/report/github.com/HDT3213/rdb)
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

Support RDB version: 1 <= version <= 12(Redis 7.2)

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
  -c command, including: json/memory/aof/bigkey/prefix/flamegraph
  -o output file path, if there is no `-o` option, output to stdout
  -n number of result, using in command: bigkey/prefix
  -port listen port for flame graph web service
  -sep separator for flamegraph, rdb will separate key by it, default value is ":". 
                supporting multi separators: -sep sep1 -sep sep2 
  -regex using regex expression filter keys
  -expire filter keys by its expiration time
		1. '>1751731200' '>now' get keys with expiration time greater than given time
		2. '<1751731200' '<now' get keys with expiration time less than given time
		3. '1751731200~1751817600' '1751731200~now' get keys with expiration time in range
		4. 'noexpire' get keys without expiration time
		5. 'anyexpire' get all keys with expiration time
  -no-expired reserve expired keys
  -concurrent The number of concurrent json converters. (CpuNum -1 by default, reserve a core for decoder)

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
5. get number and size by prefix
  rdb -c prefix [-n 10] [-max-depth 3] [-o prefix-report.csv] dump.rdb
6. draw flamegraph
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

You can use `-concurrent` to change the number of concurrent convertes. The default value is 4.

```
rdb -c json -o intset_16.json -concurrent 8 cases/intset_16.rdb
```

<details>
<summary>Json Fromat Detail</summary>
  
## string 

```json
{
    "db": 0,
    "key": "string",
    "size": 10, // estimated memory size
    "type": "string",
	"expiration":"2022-02-18T06:15:29.18+08:00",
    "value": "aaaaaaa"
}
```

## list

```json
{
    "db": 0,
    "key": "list",
    "expiration": "2022-02-18T06:15:29.18+08:00",
    "size": 66,
    "type": "list",
    "values": [
        "7fbn7xhcnu",
        "lmproj6c2e",
        "e5lom29act",
        "yy3ux925do"
    ]
}
```

## set

```json
{
    "db": 0,
    "key": "set",
    "expiration": "2022-02-18T06:15:29.18+08:00",
    "size": 39,
    "type": "set",
    "members": [
        "2hzm5rnmkmwb3zqd",
        "tdje6bk22c6ddlrw"
    ]
}
```

## hash

```json
{
    "db": 0,
    "key": "hash",
    "size": 64,
    "type": "hash",
	"expiration": "2022-02-18T06:15:29.18+08:00",
    "hash": {
        "ca32mbn2k3tp41iu": "ca32mbn2k3tp41iu",
        "mddbhxnzsbklyp8c": "mddbhxnzsbklyp8c"
    }
}
```

## zset

```json
{
    "db": 0,
    "key": "zset",
    "expiration": "2022-02-18T06:15:29.18+08:00",
    "size": 57,
    "type": "zset",
    "entries": [
        {
            "member": "zn4ejjo4ths63irg",
            "score": 1
        },
        {
            "member": "1ik4jifkg6olxf5n",
            "score": 2
        }
    ]
}
```

## stream

```json
{
    "db": 0,
    "key": "mystream",
    "size": 1776,
    "type": "stream",
    "encoding": "",
    "version": 3, // Version 2 means is RDB_TYPE_STREAM_LISTPACKS_2, 3 means is RDB_TYPE_STREAM_LISTPACKS_3
	// StreamEntry is a node in the underlying radix tree of redis stream, of type listpacks, which contains several messages. There is no need to care about which entry the message belongs to when using it.
    "entries": [ 
        {
            "firstMsgId": "1704557973866-0", // ID of the master entry at listpack head 
            "fields": [ // master fields, used for compressing size
                "name",
                "surname"
            ],
            "msgs": [ // messages in entry
                {
                    "id": "1704557973866-0",
                    "fields": {
                        "name": "Sara",
                        "surname": "OConnor"
                    },
                    "deleted": false
                }
            ]
        }
    ],
    "groups": [ // consumer groups
        {
            "name": "consumer-group-name",
            "lastId": "1704557973866-0",
            "pending": [ // pending messages
                {
                    "id": "1704557973866-0",
                    "deliveryTime": 1704557998397,
                    "deliveryCount": 1
                }
            ],
            "consumers": [ // consumers in the group
                {
                    "name": "consumer-name",
                    "seenTime": 1704557998397,
                    "pending": [
                        "1704557973866-0"
                    ],
                    "activeTime": 1704557998397
                }
            ],
            "entriesRead": 1
        }
    ],
    "len": 1, // current number of messages inside this stream
    "lastId": "1704557973866-0",
    "firstId": "1704557973866-0",
    "maxDeletedId": "0-0",
    "addedEntriesCount": 1
}
```

</details>

# Generate Memory Report

RDB uses rdb encoded size to estimate redis memory usage.

```bash
rdb -c memory -o <output_path> <source_path>
```

Example:

```bash
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

# Analyze By Prefix

If you can distinguish modules based on the prefix of the key, for example, the key of user data is `User:<uid>`, the key of Post is `Post:<postid>`, the user statistics is `Stat:User:???`, and the statistics of Post is `Stat:Post:???`.Then we can get the status of each module through prefix analysis:

```csv
database,prefix,size,size_readable,key_count
0,Post:,1170456184,1.1G,701821
0,Stat:,405483812,386.7M,3759832
0,Stat:Post:,291081520,277.6M,2775043
0,User:,241572272,230.4M,265810
0,Topic:,171146778,163.2M,694498
0,Topic:Post:,163635096,156.1M,693758
0,Stat:Post:View,133201208,127M,1387516
0,Stat:User:,114395916,109.1M,984724
0,Stat:Post:Comment:,80178504,76.5M,693758
0,Stat:Post:Like:,77701688,74.1M,693768
```

Format:

```bash
rdb -c prefix [-n <top-n>] [-max-depth <max-depth>] -o <output_path> <source_path>
```

- The prefix analysis results are arranged in descending order of memory space. The `-n` option can specify the number of outputs. All are output by default.

- `-max-depth` can limit the maximum depth of the prefix tree. In the above example, the depth of `Stat:` is 1, and the depth of `Stat:User:` and `Stat:Post:` is 2.

Example:

```bash
rdb -c prefix -n 10 -max-depth 2 -o prefix.csv cases/memory.rdb
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

# Expiration Filter

The `-expire` parameter can be configured to filter based on the expiration time.

Keys with expiration times between 2025-07-06 00:00:00 and 2025-07-07 00:00:00:
```bash
# toTimestamp(2025-07-06 00:00:00) == 1751731200
# toTimestamp(2025-07-07 00:00:00) == 1751817600
rdb -c json -o dump.json -expire 1751731200~1751817600 cases/expiration.rdb
```

```bash
# Keys with expiration times earlier than 2025-07-07 00:00:00
rdb -c json -o dump.json -expire 0~1751817600 cases/expiration.rdb
```

The magic variable `inf` represents infinity:
```bash
rdb -c json -o dump.json -expire 1751731200~inf cases/expiration.rdb
```

The magic variable `now` represents the current time:

```bash
# Keys with expiration times earlier than now
rdb -c json -o dump.json -expire 0~now cases/expiration.rdb
```

```bash
# Keys with expiration times later than now
rdb -c json -o dump.json -expire now~inf cases/expiration.rdb
```

All keys with expiration times set:
```bash
rdb -c json -o dump.json -expire anyexpire cases/expiration.rdb
```

All keys without expiration times set:
```bash
rdb -c json -o dump.json -expire noexpire cases/expiration.rdb
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
		case parser.StreamType:
			stream := o.(*parser.StreamObject)
			println(stream.Entries, stream.Groups)
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

Tested on MacBook Air（M2，2022年）, using  a 1.3 GB RDB file encoded with v9 format from Redis 5.0 in production environment.

|usage|elapsed|speed|
|:-:|:-:|:-:|
|ToJson|25s|53.24MB/s|
|Memory|10s|133.12MB/s|
|AOF|25s|53.24MB/s|
|Top10|6s|221.87MB/s|
|Prefix|25s|53.24MB/s|


