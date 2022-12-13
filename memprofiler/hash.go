package memprofiler

import "github.com/hdt3213/rdb/model"

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
	// The length of the table is the next nextPower of 2
	// When the hashtable is rehashing, another instance of **table is created
	// Due to the possibility of rehashing during loading, we calculate the worse
	// case in which both tables are allocated, and so multiply
	// the size of **table by 1.5
	return 4 + 7*sizeOfLong() + 4*sizeOfPointer() + nextPower(size)*sizeOfPointer()*3/2
}

func sizeOfHashObject(obj *model.HashObject) int {
	if obj.GetEncoding() == model.ZipListEncoding {
		extra := obj.Extra.(*model.ZiplistDetail)
		return extra.RawStringSize
	}
	size := hashtableOverhead(len(obj.Hash))
	for k, v := range obj.Hash {
		size += hashTableEntryOverhead()
		size += sizeOfString(k)
		size += sizeOfString(unsafeBytes2Str(v))
	}
	return size
}

func sizeOfSetObject(obj *model.SetObject) int {
	if obj.GetEncoding() == model.IntSetEncoding {
		extra := obj.Extra.(*model.IntsetDetail)
		return extra.RawStringSize
	}
	size := hashtableOverhead(len(obj.Members))
	for _, v := range obj.Members {
		size += hashTableEntryOverhead() + sizeOfString(unsafeBytes2Str(v))
	}
	return size
}
