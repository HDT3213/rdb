![license](https://img.shields.io/github/license/HDT3213/rdb)
[![Build Status](https://travis-ci.com/HDT3213/rdb.svg?branch=master)](https://app.travis-ci.com/github/HDT3213/rdb)
[![Coverage Status](https://coveralls.io/repos/github/HDT3213/rdb/badge.svg?branch=master)](https://coveralls.io/github/HDT3213/rdb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/HDT3213/rdb)](https://goreportcard.com/report/github.com/HDT3213/rdb)
[![Go Reference](https://pkg.go.dev/badge/github.com/hdt3213/rdb.svg)](https://pkg.go.dev/github.com/hdt3213/rdb)
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

支持 RDB 文件版本： 1 <= version <= 9

您可以在这里阅读 RDB 文件格式的详尽介绍：[Golang 实现 Redis(11): RDB 文件格式](https://www.cnblogs.com/Finley/p/16251360.html)

# 安装

如果您的电脑上安装 go 语言运行环境，可以使用 go get 安装本工具:

```bash
go get github.com/hdt3213/rdb
```

或者您可以在 [releases](https://github.com/HDT3213/rdb/releases) 页面下载可执行文件，然后将它放入 PATH 变量中的目录内。

在终端中输入 rdb 命令即可获得本工具的使用手册：

```
$ rdb
This is a tool to parse Redis' RDB files
Options:
  -c command, including: json/memory/aof
  -o output file path
  -n number of result, using in 
  -port listen port for flame graph web service
  -sep separator for flamegraph, rdb will separate key by it, default value is ":". 
                supporting multi separators: -sep sep1 -sep sep2 
  -regex using regex expression filter keys

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

# 火焰图

在很多时候并不是少量的大键值对占据了大部分内存，而是数量巨大的小键值对消耗了很多内存。目前市面上尚无分析工具可以有效处理这个问题。

很多企业要求使用 Redis key 采用类似于 `user:basic.info:{userid}` 的命名规范，所以我们可以使用分隔符将 key 拆分并将拥有相同前缀的 key 聚合在一起。

最后我们将聚合的结果以火焰图的方式呈现可以直观地看出哪类键值对消耗内存过多，进而优化缓存和逐出策略节约内存开销。

![](https://s2.loli.net/2022/03/27/eNGvVIdAuWp8EhT.png)

在上图示例中，`Comment:*` 模式的键值对消耗了 8.463% 内存.

用法：

```
rdb -c flamegraph [-port <port>] [-sep <separator1>] [-sep <separator2>] <source_path>
```

示例:

```
rdb -c flamegraph -port 16379 -sep : dump.rdb
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
|ToJson|144.11s|9.23MB/s|
|Memory|18.585s|71.62MB/s|
|AOF|104.77s|12.76MB/s|
|Top10|14.8s|89.95MB/s|
|FlameGraph|49.38s|26.96MB/s|