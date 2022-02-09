
This project is a parser for Redis' RDB files. 

It provides utilities to:

- Generate memory report for rdb file
- Convert RDB files to JSON
- Convert RDB files to Redis Serialization Protocol (or AOF file)
- Customize data usage

Thanks sripathikrishnan for his [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools)

# Install

If you have installed `go` on your compute, just simply use:
```
go get github.com/hdt3213/rdb
```

Or, you can download executable binary file from releases(https://github.com/HDT3213/rdb/releases) and put its path to PATH environment.

use `rdb` command in terminal, you can see it's manual

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

The examples for json result:
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