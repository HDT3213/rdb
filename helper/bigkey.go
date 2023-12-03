package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

type topList struct {
	list     []model.RedisObject
	capacity int
}

func (tl *topList) add(x model.RedisObject) {
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

func newRedisHeap(cap int) *topList {
	return &topList{
		capacity: cap,
	}
}

// FindBiggestKeys read rdb file and find the largest N keys.
// The invoker owns output, FindBiggestKeys won't close it
func FindBiggestKeys(rdbFilename string, topN int, output *os.File, options ...interface{}) error {
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
	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return err
	}
	top := newRedisHeap(topN)
	err = dec.Parse(func(object model.RedisObject) bool {
		top.add(object)
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
	for _, object := range top.list {
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
