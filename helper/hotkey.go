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

type hotKeyEntry struct {
	object model.RedisObject
	freq   int64
}

type hotKeyList struct {
	list     []*hotKeyEntry
	capacity int
}

func (hl *hotKeyList) add(entry *hotKeyEntry) {
	index := sort.Search(len(hl.list), func(i int) bool {
		return hl.list[i].freq <= entry.freq
	})
	hl.list = append(hl.list, entry)
	copy(hl.list[index+1:], hl.list[index:])
	hl.list[index] = entry
	if len(hl.list) > hl.capacity {
		hl.list = hl.list[:hl.capacity]
	}
}

func newHotKeyList(cap int) *hotKeyList {
	return &hotKeyList{
		capacity: cap,
	}
}

// FindHotKeys read rdb file and find the hottest N keys by LFU frequency.
//
// IMPORTANT: This function only works when the RDB file was generated from a Redis instance
// configured with LFU eviction policy (maxmemory-policy allkeys-lfu or volatile-lfu).
// Under other eviction policies, the RDB file does not contain LFU frequency data,
// and the result will be empty.
//
// Keys without LFU information are skipped.
// The invoker owns output, FindHotKeys won't close it.
func FindHotKeys(rdbFilename string, topN int, output *os.File, options ...interface{}) error {
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
	top := newHotKeyList(topN)
	err = dec.Parse(func(object model.RedisObject) bool {
		evict, ok := object.(model.EvictionInfo)
		if !ok {
			return true
		}
		freq := evict.GetFreq()
		if freq < 0 {
			return true // no LFU info, skip
		}
		top.add(&hotKeyEntry{
			object: object,
			freq:   freq,
		})
		return true
	})
	if err != nil {
		return err
	}
	_, err = output.WriteString("database,key,type,size,size_readable,freq\n")
	if err != nil {
		return fmt.Errorf("write header failed: %v", err)
	}
	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()
	for _, entry := range top.list {
		err = csvWriter.Write([]string{
			strconv.Itoa(entry.object.GetDBIndex()),
			entry.object.GetKey(),
			entry.object.GetType(),
			strconv.Itoa(entry.object.GetSize()),
			bytefmt.FormatSize(uint64(entry.object.GetSize())),
			strconv.Itoa(int(entry.freq)),
		})
		if err != nil {
			return fmt.Errorf("csv write failed: %v", err)
		}
	}
	return nil
}
