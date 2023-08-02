package helper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryProfile(t *testing.T) {
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
	testCases := []string{
		"memory",
		"stream_listpacks_1",
		"stream_listpacks_2",
		"set_listpack",
		"listpack",
	}
	for _, name := range testCases {
		srcRdb := filepath.Join("../cases", name+".rdb")
		actualFile := filepath.Join("tmp", name+".csv")
		expectFile := filepath.Join("../cases", name+".csv")
		err = MemoryProfile(srcRdb, actualFile)
		if err != nil {
			t.Errorf("error occurs during parse %s, err: %v", srcRdb, err)
			return
		}
		equals, err := compareFileByLine(t, actualFile, expectFile)
		if err != nil {
			t.Errorf("error occurs during compare %s, err: %v", srcRdb, err)
			return
		}
		if !equals {
			t.Errorf("result is not equal of %s", srcRdb)
			return
		}
	}
	err = MemoryProfile("../cases/memory.rdb", "")
	if err == nil || err.Error() != "output file path is required" {
		t.Error("failed when empty output")
	}
	err = MemoryProfile("", "tmp/memory.rdb")
	if err == nil || err.Error() != "src file path is required" {
		t.Error("failed when empty output")
	}
}

func TestMemoryWithRegex(t *testing.T) {
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
	actualFile := filepath.Join("tmp", "memory_regex.csv")
	expectFile := filepath.Join("../cases", "memory_regex.csv")
	err = MemoryProfile(srcRdb, actualFile, WithRegexOption("^l.*"))
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
	errFile := filepath.Join("tmp", "memory_regex_err.csv")
	err = MemoryProfile(srcRdb, errFile, WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
	}
}

func TestMemoryNoExpired(t *testing.T) {
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
	actualFile := filepath.Join("tmp", "memory_expired.csv")
	expectFile := filepath.Join("../cases", "memory_expired.csv")
	err = MemoryProfile(srcRdb, actualFile, WithNoExpiredOption())
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
}
