package helper

import (
	"container/heap"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
	"os"
	"strconv"
)

type redisHeap struct {
	list     []model.RedisObject
	capacity int
	minSize  int // size of min object
	minIndex int // index of min object
}

func (h redisHeap) Len() int {
	return len(h.list)
}

// Max Heap
func (h *redisHeap) Less(i, j int) bool {
	return h.list[i].GetSize() > h.list[j].GetSize()
}

func (h *redisHeap) Swap(i, j int) {
	h.list[i], h.list[j] = h.list[j], h.list[i]
}

func (h *redisHeap) Push(x interface{}) {
	h.list = append(h.list, x.(model.RedisObject))
}

func (h *redisHeap) Pop() interface{} {
	item := h.list[len(h.list)-1]
	h.list = h.list[0 : len(h.list)-1]
	return item
}

// time complexity: O(n*log(m)), n is number of redis object, m is heap capacity. m if far less than n
func (h *redisHeap) Append(x model.RedisObject) {
	// heap is full, skip
	if x.GetSize() <= h.minSize && h.Len() >= h.capacity {
		return
	}
	// if heap is full, pop min object
	if h.Len() >= h.capacity {
		// assert h.minIndex >= 0
		heap.Remove(h, h.minIndex)
	}
	heap.Push(h, x)
	// update h.minSize
	h.minSize = 1<<31 - 1
	for i := h.Len() - 1; i >= 0; i-- { //
		o := h.list[i]
		if o.GetSize() < h.minSize {
			h.minSize = o.GetSize()
			h.minIndex = i
		}
	}
}

func (h *redisHeap) Dump() []model.RedisObject {
	result := make([]model.RedisObject, 0, h.Len())
	for h.Len() > 0 {
		o := heap.Pop(h).(model.RedisObject)
		result = append(result, o)
	}
	return result
}

func newRedisHeap(cap int) *redisHeap {
	list := make([]model.RedisObject, 0, cap)
	h := &redisHeap{
		list:     list,
		capacity: cap,
		minIndex: -1,
	}
	heap.Init(h)
	return h
}

// FindBiggestKeys read rdb file and find the largest N keys.
// The invoker owns output, FindBiggestKeys won't close it
func FindBiggestKeys(rdbFilename string, topN int, output *os.File) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
	}
	if topN <= 0 {
		return errors.New("n must greater than 0")
	}
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	p := core.NewDecoder(rdbFile)
	topList := newRedisHeap(topN)
	err = p.Parse(func(object model.RedisObject) bool {
		topList.Append(object)
		return true
	})
	if err != nil {
		return err
	}
	_, err = output.WriteString("database,key,type,size,size_readable,element_count\n")
	if err != nil {
		return fmt.Errorf("write header failed: %v", err)
	}
	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()
	for topList.Len() > 0 {
		object := heap.Pop(topList).(model.RedisObject)
		err = csvWriter.Write([]string{
			strconv.Itoa(object.GetDBIndex()),
			object.GetKey(),
			object.GetType(),
			strconv.Itoa(object.GetSize()),
			bytefmt.FormatSize(uint64(object.GetSize())),
			strconv.Itoa(object.GetElemCount()),
		})
		if err != nil {
			return fmt.Errorf("csv write failed: %v", err)
		}
	}
	return nil
}
