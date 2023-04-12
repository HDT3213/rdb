package helper

import (
	"github.com/hdt3213/rdb/model"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
)

func TestRedisHeap_Append(t *testing.T) {
	sizeMap := make(map[int]struct{}) // The behavior when encountering objects of the same size is undefined
	topN := 100
	n := topN * 10
	objects := make([]model.RedisObject, 0)
	for i := 0; i < n; i++ {
		var size int
		for {
			size = rand.Intn(n * 10)
			if _, ok := sizeMap[size]; !ok {
				sizeMap[size] = struct{}{}
				break
			}
		}
		o := &model.StringObject{
			BaseObject: &model.BaseObject{
				Key:  strconv.Itoa(i),
				Size: size,
			},
		}
		objects = append(objects, o)
	}
	topList := newRedisHeap(topN)
	for _, o := range objects {
		topList.Append(o)
	}
	actual := topList.Dump()
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].GetSize() > objects[j].GetSize()
	})
	expect := objects[0:topN]
	for i := 0; i < topN; i++ {
		o1 := actual[i]
		o2 := expect[i]
		if o1.GetSize() != o2.GetSize() {
			t.Errorf("wrong answer at index: %d", i)
		}
	}
}

func TestFindLargestKeys(t *testing.T) {
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
	expectFile := filepath.Join("../cases", "largest.csv")
	outputFilePath := filepath.Join("tmp", "largest.csv")
	output, err := os.Create(outputFilePath)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = FindBiggestKeys(srcRdb, 5, output)
	if err != nil {
		t.Errorf("FindLargestKeys failed: %v", err)
	}
	err = output.Close()
	if err != nil {
		t.Errorf("error occurs during close output %s, err: %v", srcRdb, err)
		return
	}
	equals, err := compareFileByLine(t, outputFilePath, expectFile)
	if err != nil {
		t.Errorf("error occurs during compare %s, err: %v", srcRdb, err)
		return
	}
	if !equals {
		t.Errorf("result is not equal of %s", srcRdb)
		return
	}

	err = FindBiggestKeys("", 5, os.Stdout)
	if err == nil || err.Error() != "src file path is required" {
		t.Error("failed when empty output")
	}
	err = FindBiggestKeys("cases/memory.rdb", 0, os.Stdout)
	if err == nil || err.Error() != "n must greater than 0" {
		t.Error("failed when empty output")
	}
}

func TestFindBiggestKeyWithRegex(t *testing.T) {
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
	actualFile := filepath.Join("../cases", "memory_regex.biggest.csv")
	expectFile := filepath.Join("../cases", "memory_regex.biggest.csv")
	output, err := os.Create(actualFile)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = FindBiggestKeys(srcRdb, 2, output, WithRegexOption("^l.*"))
	if err != nil {
		t.Errorf("error occurs during parse, err: %v", err)
		return
	}
	equals, err := compareFileByLine(t, actualFile, expectFile)
	if err != nil {
		t.Errorf("error occurs during compare err: %v", err)
		return
	}
	if !equals {
		t.Errorf("result is not equal")
		return
	}

	err = FindBiggestKeys(srcRdb, 2, output, WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
	}
}
