package helper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrefixAnalyse(t *testing.T) {
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
	srcRdb := filepath.Join("../cases", "tree.rdb")

	// test top3
	actualTop3 := filepath.Join("tmp", "tree.csv")
	expectTop3 := filepath.Join("../cases", "tree.csv")
	actualTop3File, err := os.Create(actualTop3)
	if err != nil {
		t.Error(err)
		return
	}
	err = PrefixAnalyse(srcRdb, 3, 0, actualTop3File)
	if err != nil {
		t.Error(err)
		return
	}

	equals, err := compareFileByLine(t, actualTop3, expectTop3)
	if err != nil {
		t.Errorf("error occurs during compare top3, err: %v", err)
		return
	}
	if !equals {
		t.Error("result is not equal of top3")
		return
	}

	// test depth=2 
	actualDepth2 := filepath.Join("tmp", "tree2.csv")
	expectDepth2 := filepath.Join("../cases", "tree2.csv")
	actualDepth2File, err := os.Create(actualDepth2)
	if err != nil {
		t.Error(err)
		return
	}
	err = PrefixAnalyse(srcRdb, 0, 2, actualDepth2File)
	if err != nil {
		t.Error(err)
		return
	}

	equals, err = compareFileByLine(t, actualDepth2, expectDepth2)
	if err != nil {
		t.Errorf("error occurs during compare top3, err: %v", err)
		return
	}
	if !equals {
		t.Error("result is not equal of top3")
		return
	}

}