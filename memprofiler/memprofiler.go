package memprofiler

import (
	"github.com/hdt3213/rdb/model"
	"math/rand"
)

// used to evaluate memory usage

func hashTableEntryOverhead() int {
	// See  https://github.com/antirez/redis/blob/unstable/src/dict.h
	// Each dictEntry has 2 pointers + int64
	return 2*sizeOfPointer() + 8
}

func hashtableOverhead(size int) int {
	// See  https://github.com/antirez/redis/blob/unstable/src/dict.h
	// See the structures dict and dictht
	// 2 * (3 unsigned longs + 1 pointer) + int + long + 2 pointers
	//
	// Additionally, see **table in dictht
	// The length of the table is the next power of 2
	// When the hashtable is rehashing, another instance of **table is created
	// Due to the possibility of rehashing during loading, we calculate the worse
	// case in which both tables are allocated, and so multiply
	// the size of **table by 1.5
	return 4 + 7*sizeOfLong() + 4*sizeOfPointer() + power(size)*sizeOfPointer()*3/2
}

func skipListOverhead(size int) int {
	return 2*sizeOfPointer() + hashtableOverhead(size) + (2*sizeOfPointer() + 16)
}

func skipListEntryOverhead() int {
	return hashTableEntryOverhead() + 2*sizeOfPointer() + 8 + (sizeOfPointer()+8)*zsetRandomLevel()
}

func zsetRandomLevel() int {
	const maxLevel = 32
	const p = 25
	i := 1
	for ; i <= maxLevel; i++ {
		r := rand.Intn(100)
		if r >= p {
			break
		}
	}
	return i
}

// RedisMeta stores redis version and architecture
type RedisMeta struct {
	Version string
	Bits    int // 32/64
}

// SizeOfObject evaluates memory usage of obj
func SizeOfObject(obj model.RedisObject) int {
	// todo: memory profile by redis version and architecture
	size := topLevelObjectOverhead(obj.GetKey(), obj.GetExpiration() != nil)
	switch o := obj.(type) {
	case *model.StringObject:
		size += sizeOfString(unsafeBytes2Str(o.Value))
	case *model.ListObject:
		size += sizeOfListObject(o)
	case *model.SetObject:
	case *model.HashObject:
	case *model.ZSetObject:
	}
	return size
}
