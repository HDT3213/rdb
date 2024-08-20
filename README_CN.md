![license](https://img.shields.io/github/license/HDT3213/rdb)
![download](https://img.shields.io/github/downloads/hdt3213/rdb/total)
[![Go Reference](https://pkg.go.dev/badge/github.com/hdt3213/rdb.svg)](https://pkg.go.dev/github.com/hdt3213/rdb)
<br>
[![Build Status](https://github.com/hdt3213/rdb/actions/workflows/main.yml/badge.svg)](https://github.com/HDT3213/rdb/actions?query=branch%3Amaster)
[![Coverage Status](https://coveralls.io/repos/github/HDT3213/rdb/badge.svg?branch=master)](https://coveralls.io/github/HDT3213/rdb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/HDT3213/rdb)](https://goreportcard.com/report/github.com/HDT3213/rdb)
<br>
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)

[English](https://github.com/HDT3213/rdb/blob/master/README.md)

这是一个可以用于二次开发和内存分析的 RDB 文件分析工具，它具备下列能力：
- 为 RDB 文件生成内存用量报告
- 将 RDB 文件中键值对数据转换为 JSON 格式
- 将 RDB 文件转换为 AOF 文件（即 Redis 序列化协议）
- 寻找 RDB 文件中大键值对
- 根据 RDB 文件绘制内存火焰图，用来分析哪类键值对占用了最多内存
- 通过 API 遍历 RDB 文件内容，自定义用途
- 生成 RDB 文件

支持 RDB 文件版本： 1 <= version <= 12(Redis 7.2)

您可以在这里阅读 RDB 文件格式的详尽介绍：[Golang 实现 Redis(11): RDB 文件格式](https://www.cnblogs.com/Finley/p/16251360.html)

# 安装

如果您的电脑上安装 go 语言运行环境，可以使用 go get 安装本工具:

```bash
go install github.com/hdt3213/rdb@latest
```

或者您可以在 [releases](https://github.com/HDT3213/rdb/releases) 页面下载可执行文件，然后将它放入 PATH 变量中的目录内。

在终端中输入 rdb 命令即可获得本工具的使用手册：

```
$ rdb
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

# 转换为 JSON 格式

用法：

```
rdb -c json -o <output_path> <source_path>
```

示例：

```
rdb -c json -o intset_16.json cases/intset_16.rdb
```

本仓库的 [cases](https://github.com/HDT3213/rdb/tree/master/cases) 目录中准备了一些示例 RDB 文件，可供您进行测试。

转换出的 JSON 结果示例：

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

<details>
<summary>Json 格式</summary>
  
## string 

```json
{
    "db": 0,
    "key": "string",
    "size": 10, // 估计的内存占用量
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
    "version": 3, // Version 2 表示 RDB_TYPE_STREAM_LISTPACKS_2, 3 表示RDB_TYPE_STREAM_LISTPACKS_3
    "entries": [ // StreamEntry 是 redis stream 底层 radix tree 中的一个节点，类型为 listpacks, 其中包含了若干条消息。在使用时无需关心消息属于哪个 entry。
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

# 生成内存用量报告

本工具使用 RDB 编码后的大小来估算键值对占用的内存大小。

用法：

```
rdb -c memory -o <output_path> <source_path>
```

示例：

```
rdb -c memory -o mem.csv cases/memory.rdb
```

内存报告示例：

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

# 前缀分析

如果您可以根据 key 的前缀区分模块，比如用户数据的 key 是 `User:<uid>`， Post 的模式是 `Post:<postid>`, 用户统计信息是 `Stat:User:???`, Post 的统计信息是 `Stat:User:???`。 那么我们可以通过前缀分析来得到各模块的情况：

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

命令格式：

```bash
rdb -c prefix [-n <top-n>] [-max-depth <max-depth>] -o <output_path> <source_path>
```

- 前缀分析结果按照内存空间从大到小排列，`-n` 选项可以指定输出的数量。默认全部输出。

- `-max-depth` 可以限制前缀树的的最大深度。比如示例中 `Stat:` 的深度是1，`Stat:User:` 和 `Stat:Post:` 的深度是 2。

Example:

```bash
rdb -c prefix -n 10 -max-depth 2 -o prefix.csv cases/memory.rdb
```


# 火焰图

在很多时候并不是少量的大键值对占据了大部分内存，而是数量巨大的小键值对消耗了很多内存。

很多企业要求使用 Redis key 采用类似于 `user:basic.info:{userid}` 的命名规范，所以我们可以使用分隔符将 key 拆分并将拥有相同前缀的 key 聚合在一起。

最后我们将聚合的结果以火焰图的方式呈现可以直观地看出哪类键值对消耗内存过多，进而优化缓存和逐出策略节约内存开销。

![截屏2022-10-30 12.06.00.png](https://s2.loli.net/2022/11/08/HW9ZxGfeEzArUhM.png)

在上图示例中，`Comment:*` 模式的键值对消耗了 8.463% 内存.

用法：

```
rdb -c flamegraph [-port <port>] [-sep <separator1>] [-sep <separator2>] <source_path>
```

示例:

```
rdb -c flamegraph -port 16379 -sep : dump.rdb
```

# 寻找最大的键值对

本工具可以用来寻找 RDB 文件中最大的 N 个键值对。用法：

```
rdb -c bigkey -n <result_number> <source_path>
```

示例：

```
rdb -c bigkey -n 5 cases/memory.rdb
```

结果示例：

```csv
database,key,type,size,size_readable,element_count
0,large,string,2056,2K,0
0,list,list,66,66B,4
0,hash,hash,64,64B,2
0,zset,zset,57,57B,2
0,set,set,39,39B,2
```

# 转换为 AOF 文件

用法：

```
rdb -c aof -o <output_path> <source_path>
```

示例：

```
rdb -c aof -o mem.aof cases/memory.rdb
```

输出的 AOF 文件示例：

```
*3
$3
SET
$1
s
$7
aaaaaaa
```

# 正则过滤器

本工具支持使用正则表达式过滤自己关心的键值对：

示例:

```bash
rdb -c json -o regex.json -regex '^l.*' cases/memory.rdb
```

# 自定义用途

除了命令行工具之外，您可以在自己的项目中引入 hdt3213/rdb/parser 包，自行决定如何处理 RDB 中的数据。

示例：

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

# 生成 RDB 文件

除了解析之外，本项目也可以用于生成 RDB 文件：

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

在 MacBook Pro (16-inch, 2019) 2.6 GHz 六核 Intel Core i7 笔记本上，使用从生产环境的 Redis 5.0 上获得 1.3 GB 大小使用 v9 编码的 RDB 文件进行测试：

|usage|elapsed|speed|
|:-:|:-:|:-:|
|ToJson|74.12s|17.96MB/s|
|Memory|18.585s|71.62MB/s|
|AOF|104.77s|12.76MB/s|
|Top10|14.8s|89.95MB/s|
|FlameGraph|21.83s|60.98MB/s|