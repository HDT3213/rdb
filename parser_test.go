package main

import (
	"bufio"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/helper"
	"github.com/hdt3213/rdb/model"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func compareFileByLine(t *testing.T, fn1, fn2 string) (bool, error) {
	f1, err := os.Open(fn1)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f1.Close()
	}()
	f2, err := os.Open(fn2)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f2.Close()
	}()
	sc1 := bufio.NewScanner(f1)
	sc2 := bufio.NewScanner(f2)

	for {
		next1 := sc1.Scan()
		next2 := sc2.Scan()
		if !next1 && !next2 {
			break
		}
		if next1 != next2 {
			// line number is not equals
			t.Log("line number is not equal")
			return false, nil
		}
		txt1 := sc1.Text()
		txt2 := sc2.Text()
		if txt1 != txt2 {
			t.Logf("txt1: %s\ntxt2:%s", txt1, txt2)
			return false, nil
		}
	}
	return true, nil
}

func TestToJson(t *testing.T) {
	// use same time zone to ensure RedisObject.Expiration has same json value
	var cstZone = time.FixedZone("CST", 8*3600)
	time.Local = cstZone

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
		"quicklist",
		"easily_compressible_string_key",
		"empty_database",
		"hash",
		"hash_as_ziplist",
		"integer_keys",
		"intset_16",
		"intset_32",
		"intset_64",
		"keys_with_expiry",
		"linkedlist",
		"multiple_databases",
		"non_ascii_values",
		"parser_filters",
		"rdb_version_5_with_checksum",
		"rdb_version_8_with_64b_length_and_scores",
		"regular_set",
		"regular_sorted_set",
		"sorted_set_as_ziplist",
		"uncompressible_string_keys",
		"ziplist_that_compresses_easily",
		"ziplist_that_doesnt_compress",
		"ziplist_with_integers",
		"zipmap_that_compresses_easily",
		"zipmap_that_doesnt_compress",
		"zipmap_with_big_values",
		"zipmap_big_len",
	}
	for _, filename := range testCases {
		srcRdb := filepath.Join("cases", filename+".rdb")
		actualJSON := filepath.Join("cases", filename+".json")
		expectJSON := filepath.Join("cases", filename+".json")
		err = helper.ToJsons(srcRdb, actualJSON)
		if err != nil {
			t.Errorf("error occurs during parse %s, err: %v", filename, err)
			continue
		}
		equals, err := compareFileByLine(t, actualJSON, expectJSON)
		if err != nil {
			t.Errorf("error occurs during compare %s, err: %v", filename, err)
			continue
		}
		if !equals {
			t.Errorf("result is not equal of %s", filename)
			continue
		}
	}
	err = helper.ToJsons("cases/memory.rdb", "")
	if err == nil || err.Error() != "output file path is required" {
		t.Error("failed when empty output")
	}
	err = helper.ToJsons("", "tmp/memory.rdb")
	if err == nil || err.Error() != "src file path is required" {
		t.Error("failed when empty output")
	}
}

func TestToJsonWithRegex(t *testing.T) {
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
	srcRdb := filepath.Join("cases", "memory.rdb")
	actualJSON := filepath.Join("tmp", "memory_regex.json")
	expectJSON := filepath.Join("cases", "memory_regex.json")
	err = helper.ToJsons(srcRdb, actualJSON, helper.WithRegexOption("^l.*"))
	if err != nil {
		t.Errorf("error occurs during parse, err: %v", err)
		return
	}
	equals, err := compareFileByLine(t, actualJSON, expectJSON)
	if err != nil {
		t.Errorf("error occurs during compare err: %v", err)
		return
	}
	if !equals {
		t.Errorf("result is not equal")
		return
	}
	err = helper.ToJsons(srcRdb, actualJSON, helper.WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
	}
}

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
	srcRdb := filepath.Join("cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory.csv")
	expectFile := filepath.Join("cases", "memory.csv")
	err = helper.MemoryProfile(srcRdb, actualFile)
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
	err = helper.MemoryProfile("cases/memory.rdb", "")
	if err == nil || err.Error() != "output file path is required" {
		t.Error("failed when empty output")
	}
	err = helper.MemoryProfile("", "tmp/memory.rdb")
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
	srcRdb := filepath.Join("cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory_regex.csv")
	expectFile := filepath.Join("cases", "memory_regex.csv")
	err = helper.MemoryProfile(srcRdb, actualFile, helper.WithRegexOption("^l.*"))
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
	err = helper.MemoryProfile(srcRdb, actualFile, helper.WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
	}
}

func TestToAof(t *testing.T) {
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
	srcRdb := filepath.Join("cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory.aof")
	expectFile := filepath.Join("cases", "memory.aof")
	err = helper.ToAOF(srcRdb, actualFile)
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
	err = helper.ToAOF("cases/memory.rdb", "")
	if err == nil || err.Error() != "output file path is required" {
		t.Error("failed when empty output")
	}
	err = helper.ToAOF("", "tmp/memory.rdb")
	if err == nil || err.Error() != "src file path is required" {
		t.Error("failed when empty output")
	}
}

func TestToAofWithRegex(t *testing.T) {
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
	srcRdb := filepath.Join("cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory_regex.aof")
	expectFile := filepath.Join("cases", "memory_regex.aof")
	err = helper.ToAOF(srcRdb, actualFile, helper.WithRegexOption("^l.*"))
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
	err = helper.ToAOF(srcRdb, actualFile, helper.WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
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
	srcRdb := filepath.Join("cases", "memory.rdb")
	expectFile := filepath.Join("cases", "largest.csv")
	outputFilePath := filepath.Join("tmp", "largest.csv")
	output, err := os.Create(outputFilePath)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = helper.FindBiggestKeys(srcRdb, 5, output)
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

	err = helper.FindBiggestKeys("", 5, os.Stdout)
	if err == nil || err.Error() != "src file path is required" {
		t.Error("failed when empty output")
	}
	err = helper.FindBiggestKeys("cases/memory.rdb", 0, os.Stdout)
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
	srcRdb := filepath.Join("cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory_regex.biggest.csv")
	expectFile := filepath.Join("cases", "memory_regex.biggest.csv")
	output, err := os.Create(actualFile)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = helper.FindBiggestKeys(srcRdb, 2, output, helper.WithRegexOption("^l.*"))
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

	output, err = os.Create(actualFile)
	if err != nil {
		t.Errorf("create output file failed: %v", err)
		return
	}
	err = helper.FindBiggestKeys(srcRdb, 2, output, helper.WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
	}
}

func TestFlameGraph(t *testing.T) {
	helper.TrimThreshold = 1
	srcRdb := filepath.Join("cases", "tree.rdb")
	stop, err := helper.FlameGraph(srcRdb, 18888, nil)
	if err != nil {
		t.Errorf("draw FlameGraph failed: %v", err)
	}
	resp, err := http.Get("http://localhost:18888/flamegraph")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("http %d", resp.StatusCode)
		return
	}
	resp, err = http.Get("http://localhost:18888/stacks.json")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("http %d", resp.StatusCode)
		return
	}
	stop <- struct{}{}

	stop, err = helper.FlameGraph(srcRdb, 0, nil)
	if err != nil {
		t.Errorf("FindLargestKeys failed: %v", err)
	}
	stop <- struct{}{}

	_, err = helper.FlameGraph("", 0, nil)
	if err == nil || err.Error() != "src file path is required" {
		t.Error("expect error: src file path is required")
	}
}

func TestFlameGraphWithRegex(t *testing.T) {
	srcRdb := filepath.Join("cases", "tree.rdb")
	stop, err := helper.FlameGraph(srcRdb, 18888, nil, helper.WithRegexOption("^l.*"))
	if err != nil {
		t.Errorf("FindLargestKeys failed: %v", err)
	}
	stop <- struct{}{}

	_, err = helper.FlameGraph(srcRdb, 18888, nil, helper.WithRegexOption(`(1)\2`))
	if err == nil {
		t.Errorf("expect error: %v", err)
	}
}

func TestWithSpecialObject(t *testing.T) {
	rdbFilename := filepath.Join("cases", "memory.rdb")
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		t.Errorf("open rdb %s failed, %v", rdbFilename, err)
		return
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	expectAux := map[string]string{
		"redis-ver":    "6.0.6",
		"redis-bits":   "64",
		"ctime":        "1644136130",
		"used-mem":     "1167584",
		"aof-preamble": "0",
	}
	expectKeyCount := map[int]uint64{
		0: 7,
	}
	expectTTLCount := map[int]uint64{
		0: 1,
	}
	var auxCount, dbSizeCount int
	dec := core.NewDecoder(rdbFile).WithSpecialOpCode()
	err = dec.Parse(func(object model.RedisObject) bool {
		switch o := object.(type) {
		case *model.AuxObject:
			if o.GetType() != model.AuxType {
				t.Error("aux obj with wrong type")
			}
			expectValue := expectAux[o.Key]
			if o.Value != expectValue {
				t.Errorf("aux %s has wrong value", o.GetKey())
			}
			auxCount++
		case *model.DBSizeObject:
			dbSizeCount++
			if o.KeyCount != expectKeyCount[o.DB] {
				t.Errorf("db %d has wrong key count", o.DB)
			}
			if o.TTLCount != expectTTLCount[o.DB] {
				t.Errorf("db %d has wrong ttl count", o.DB)
			}
		}
		return true
	})
	if err != nil {
		t.Error(err)
	}
	if auxCount != len(expectAux) {
		t.Error("wrong aux object count")
	}
	if dbSizeCount != 1 {
		t.Error("wrong db size object count")
	}
}
