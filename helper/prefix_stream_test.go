package helper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStreamingPrefixAnalyse(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()
	srcRdb := filepath.Join("../cases", "tree.rdb")

	// test with empty separator (per-character), depth 0 (unlimited), top 3
	actualFile := filepath.Join("tmp", "stream_tree.csv")
	f, err := os.Create(actualFile)
	if err != nil {
		t.Fatal(err)
	}
	err = StreamingPrefixAnalyse(srcRdb, 3, 0, "", f, )
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	expectFile := filepath.Join("../cases", "tree.csv")
	equals, err := compareFileByLine(t, actualFile, expectFile)
	if err != nil {
		t.Fatalf("error comparing files: %v", err)
	}
	if !equals {
		t.Error("streaming prefix top3 result does not match expected")
	}

	// test with empty separator, depth 2
	actualFile2 := filepath.Join("tmp", "stream_tree2.csv")
	f2, err := os.Create(actualFile2)
	if err != nil {
		t.Fatal(err)
	}
	err = StreamingPrefixAnalyse(srcRdb, 0, 2, "", f2)
	if err != nil {
		t.Fatal(err)
	}
	_ = f2.Close()

	expectFile2 := filepath.Join("../cases", "tree2.csv")
	equals, err = compareFileByLine(t, actualFile2, expectFile2)
	if err != nil {
		t.Fatalf("error comparing files: %v", err)
	}
	if !equals {
		t.Error("streaming prefix depth2 result does not match expected")
	}
}
