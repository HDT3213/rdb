package helper

import (
	"sort"
)

type Sized interface {
	GetSize() int
}

type topList struct {
	list     []Sized
	capacity int
}

func (tl *topList) add(x Sized) {
	index := sort.Search(len(tl.list), func(i int) bool {
		return tl.list[i].GetSize() <= x.GetSize()
	})
	tl.list = append(tl.list, x)
	copy(tl.list[index+1:], tl.list[index:])
	tl.list[index] = x
	if len(tl.list) > tl.capacity {
		tl.list = tl.list[:tl.capacity]
	}
}

func newToplist(cap int) *topList {
	return &topList{
		capacity: cap,
	}
}
