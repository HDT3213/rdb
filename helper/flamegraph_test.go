package helper

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestSplit(t *testing.T) {
	result := split("a:b:c", nil)
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("wrong result: %+v", result)
	}
	result = split("a.b.c", []string{"."})
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("wrong result: %+v", result)
	}
	result = split("a++b--c", []string{"++", "--"})
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("wrong result: %+v", result)
	}
}

func TestFlameGraph(t *testing.T) {
	TrimThreshold = 1
	srcRdb := filepath.Join("../cases", "tree.rdb")
	stop, err := FlameGraph(srcRdb, 18888, nil)
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

	stop, err = FlameGraph(srcRdb, 0, nil)
	if err != nil {
		t.Errorf("FindLargestKeys failed: %v", err)
	}
	stop <- struct{}{}

	_, err = FlameGraph("", 0, nil)
	if err == nil || err.Error() != "src file path is required" {
		t.Error("expect error: src file path is required")
	}
}

func TestFlameGraphWithRegex(t *testing.T) {
	srcRdb := filepath.Join("../cases", "tree.rdb")
	stop, err := FlameGraph(srcRdb, 18888, nil, WithRegexOption("^l.*"))
	if err != nil {
		t.Errorf("FindLargestKeys failed: %v", err)
	}
	stop <- struct{}{}

	_, err = FlameGraph(srcRdb, 18888, nil, WithRegexOption(`(1)\2`))
	if err == nil {
		t.Errorf("expect error: %v", err)
	}
}
