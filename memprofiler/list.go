package memprofiler

import (
	"github.com/hdt3213/rdb/model"
	"math"
	"strconv"
)

func sizeOfListObject(obj *model.ListObject) int {
	switch obj.GetEncoding() {
	case model.QuickListEncoding:
		detail := obj.Extra.(*model.QuicklistDetail)
		return sizeOfQuickList(detail)
	case model.ListEncoding:
		return sizeOfList(obj.Values)
	case model.ZipListEncoding:
		return sizeOfZiplist(obj.Values)
	}
	return 0
}

func zipListIntEntryOverhead(v int) int {
	header := 1
	size := 0
	if v < 12 {
		size = 0
	} else if v < int(math.Pow(2, 8)) {
		size = 1
	} else if v < int(math.Pow(2, 16)) {
		size = 2
	} else if v < int(math.Pow(2, 24)) {
		size = 3
	} else if v < int(math.Pow(2, 32)) {
		size = 4
	} else {
		size = 8
	}
	prevLen := 1
	if size < 254 {
		prevLen = 5
	}
	return prevLen + header + size
}

func zipListStrEntryOverhead(v string) int {
	size := len(v)
	header := 1
	if size <= 63 {
		header = 1
	} else if size <= 1 {
		header = 2
	} else {
		header = 5
	}
	prevLen := 1
	if size < 254 {
		prevLen = 5
	}
	return prevLen + header + size
}

func sizeOfZiplist(values [][]byte) int {
	// See https://github.com/antirez/redis/blob/unstable/src/ziplist.c
	// <zlbytes><zltail><zllen><entry><entry><zlend>
	size := 4 + 4 + 2 + 1
	for _, value := range values {
		str := unsafeBytes2Str(value)
		i, err := strconv.Atoi(str)
		if err != nil { // not int
			size += zipListStrEntryOverhead(str)
		} else { // is int
			size += zipListIntEntryOverhead(i)
		}
	}
	return size
}

func sizeOfQuickList(detail *model.QuicklistDetail) int {
	size := 2*sizeOfPointer() + sizeOfLong() + 2*4
	itemHeaderSize := 4*sizeOfPointer() + sizeOfLong() + 2*4
	size += len(detail.ZiplistStruct) * itemHeaderSize
	for _, ziplist := range detail.ZiplistStruct {
		size += sizeOfZiplist(ziplist)
	}
	return size
}

func sizeOfList(values [][]byte) int {
	// See https://github.com/antirez/redis/blob/unstable/src/adlist.h
	// A list has 5 pointers + an unsigned long
	size := 5*sizeOfPointer() + sizeOfLong()
	// A node has 3 pointers
	entryHeadSize := 3 * sizeOfPointer()
	size += len(values) * entryHeadSize
	// fixme: since redis 4.0
	for _, v := range values {
		s := unsafeBytes2Str(v)
		size += sizeOfString(s)
	}
	return size
}
