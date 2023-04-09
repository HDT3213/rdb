package memprofiler

import (
	"github.com/hdt3213/rdb/model"
)

// used to evaluate memory usage

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
		size += sizeOfSetObject(o)
	case *model.HashObject:
		size += sizeOfHashObject(o)
	case *model.ZSetObject:
		size += sizeOfZSetObject(o)
	case *model.StreamObject:
		size += sizeOfStreamObject(o)
	}
	return size
}
