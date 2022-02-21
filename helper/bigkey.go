package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
	"os"
	"strconv"
)

type redisTreeSet struct {
	set      *treeset.Set
	capacity int
}

func (h *redisTreeSet) GetMin() model.RedisObject {
	iter := h.set.Iterator()
	iter.End()
	if iter.Prev() {
		raw := iter.Value()
		return raw.(model.RedisObject)
	}
	return nil
}

// time complexity: O(n*log(m)), n is number of redis object, m is heap capacity. m if far less than n
func (h *redisTreeSet) Append(x model.RedisObject) {
	// if heap is full && x.Size > minSize, then pop min
	if h.set.Size() == h.capacity {
		min := h.GetMin()
		if min.GetSize() < x.GetSize() {
			h.set.Remove(min)
		}
	}
	h.set.Add(x)
}

func (h *redisTreeSet) Dump() []model.RedisObject {
	result := make([]model.RedisObject, 0, h.set.Size())
	iter := h.set.Iterator()
	for iter.Next() {
		result = append(result, iter.Value().(model.RedisObject))
	}
	return result
}

func newRedisHeap(cap int) *redisTreeSet {
	s := treeset.NewWith(func(a, b interface{}) int {
		o1 := a.(model.RedisObject)
		o2 := b.(model.RedisObject)
		return o2.GetSize() - o1.GetSize() // desc order
	})
	return &redisTreeSet{
		set:      s,
		capacity: cap,
	}
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
	iter := topList.set.Iterator()
	for iter.Next() {
		object := iter.Value().(model.RedisObject)
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
