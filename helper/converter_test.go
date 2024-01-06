package helper

import (
	"bufio"
	"github.com/bytedance/sonic"
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
	// SortMapKeys will cause performance losses, only enabled during test
	jsonEncoder = sonic.ConfigStd
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
		"stream_listoacks_3",
		"issue27",
		"set_listpack",
		"stream_listpacks_2",
		"stream_listpacks_1",
		"listpack",
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
		srcRdb := filepath.Join("../cases", filename+".rdb")
		actualJSON := filepath.Join("tmp", filename+".json")
		expectJSON := filepath.Join("../cases", filename+".json")
		err = ToJsons(srcRdb, actualJSON)
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
	err = ToJsons("cases/memory.rdb", "")
	if err == nil || err.Error() != "output file path is required" {
		t.Error("failed when empty output")
	}
	err = ToJsons("", "tmp/memory.rdb")
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
	srcRdb := filepath.Join("../cases", "memory.rdb")
	actualJSON := filepath.Join("tmp", "memory_regex.json")
	expectJSON := filepath.Join("../cases", "memory_regex.json")
	err = ToJsons(srcRdb, actualJSON, WithRegexOption("^l.*"))
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
	errJson := filepath.Join("tmp", "memory_regex_err.json")
	err = ToJsons(srcRdb, errJson, WithRegexOption(`(i)\1`))
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
	srcRdb := filepath.Join("../cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory.aof")
	expectFile := filepath.Join("../cases", "memory.aof")
	err = ToAOF(srcRdb, actualFile, lexOrder{})
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
	err = ToAOF("cases/memory.rdb", "")
	if err == nil || err.Error() != "output file path is required" {
		t.Error("failed when empty output")
	}
	err = ToAOF("", "tmp/err.rdb")
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
	srcRdb := filepath.Join("../cases", "memory.rdb")
	actualFile := filepath.Join("tmp", "memory_regex.aof")
	expectFile := filepath.Join("../cases", "memory_regex.aof")
	err = ToAOF(srcRdb, actualFile, WithRegexOption("^l.*"), lexOrder{})
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
	errFile := filepath.Join("tmp", "memory_regex.err.aof")
	err = ToAOF(srcRdb, errFile, WithRegexOption(`(i)\1`))
	if err == nil {
		t.Error("expect error")
	}
}
