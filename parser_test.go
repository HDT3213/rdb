package main

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
)

func compareFileByLine(fn1, fn2 string) (bool, error) {
	f1, err := os.Open(fn1)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f1.Close()
	}()
	f2, err := os.Open(fn1)
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
			return false, nil
		}
		if sc1.Text() != sc2.Text() {
			return false, nil
		}
	}
	return true, nil
}

func TestParse(t *testing.T) {
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
	}
	for _, filename := range testCases {
		srcRdb := filepath.Join("cases", filename+".rdb")
		actualJson := filepath.Join("tmp", filename+".jsons")
		expectJson := filepath.Join("cases", filename+".jsons")
		err = ToJsons(srcRdb, actualJson)
		if err != nil {
			t.Errorf("error occurs during parse %s, err: %v", filename, err)
			continue
		}
		equals, err := compareFileByLine(actualJson, expectJson)
		if err != nil {
			t.Errorf("error occurs during compare %s, err: %v", filename, err)
			continue
		}
		if !equals {
			t.Errorf("result is not equal of %s", filename)
			continue
		}
	}
}
