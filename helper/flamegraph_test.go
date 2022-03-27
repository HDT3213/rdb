package helper

import "testing"

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
