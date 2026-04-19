package helper

import (
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/hdt3213/rdb/model"
)

func TestHotKeyList(t *testing.T) {
	topN := 100
	n := topN * 10
	type entry struct {
		object model.RedisObject
		freq   int64
	}
	entries := make([]*entry, 0)
	for i := 0; i < n; i++ {
		freq := int64(rand.Intn(256))
		o := &model.StringObject{
			BaseObject: &model.BaseObject{
				Key:  strconv.Itoa(i),
				Size: rand.Intn(n * 10),
				Freq: &freq,
			},
		}
		entries = append(entries, &entry{object: o, freq: freq})
	}
	hotList := newHotKeyList(topN)
	for _, e := range entries {
		hotList.add(&hotKeyEntry{object: e.object, freq: e.freq})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].freq > entries[j].freq
	})
	if len(hotList.list) != topN {
		t.Error("wrong hot key list size")
	}
	for i, actual := range hotList.list {
		expect := entries[i]
		if actual.freq != expect.freq {
			t.Errorf("wrong hot key list at index %d, expect freq %d, got %d", i, expect.freq, actual.freq)
			return
		}
	}
}

func TestFindHotKeys(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		err := os.RemoveAll("tmp")
		if err != nil {
			t.Logf("remove tmp directory failed: %v", err)
		}
	}()
	srcRdb := filepath.Join("../cases", "memory.rdb")
	outputFilePath := filepath.Join("tmp", "hotkey.csv")
	output, err := os.Create(outputFilePath)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = FindHotKeys(srcRdb, 5, output)
	if err != nil {
		t.Errorf("FindHotKeys failed: %v", err)
	}
	err = output.Close()
	if err != nil {
		t.Errorf("error occurs during close output, err: %v", err)
		return
	}

	// test error cases
	err = FindHotKeys("", 5, os.Stdout)
	if err == nil || err.Error() != "src file path is required" {
		t.Error("failed when empty src")
	}
	err = FindHotKeys("cases/memory.rdb", 0, os.Stdout)
	if err == nil || err.Error() != "n must greater than 0" {
		t.Error("failed when n <= 0")
	}
}

func TestFindHotKeysWithRegex(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		err := os.RemoveAll("tmp")
		if err != nil {
			t.Logf("remove tmp directory failed: %v", err)
		}
	}()
	srcRdb := filepath.Join("../cases", "memory.rdb")
	outputFilePath := filepath.Join("tmp", "hotkey_regex.csv")
	output, err := os.Create(outputFilePath)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = FindHotKeys(srcRdb, 5, output, WithRegexOption("^l.*"))
	if err != nil {
		t.Errorf("FindHotKeys with regex failed: %v", err)
	}
	err = output.Close()
	if err != nil {
		t.Errorf("error occurs during close output, err: %v", err)
		return
	}
}
