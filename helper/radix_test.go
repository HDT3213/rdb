package helper

import "testing"

func TestRadix(t *testing.T) {
	words := []string{
		"a",
		"b",
		"abbd",
		"abba",
		"abc",
	}
	tree := newRadixTree()
	for i, word := range words {
		tree.insert(word, i+1)
	}
	expectSizeMap := map[string]int {
		"": 15,
		"a": 13,
		"b": 2,
		"ab": 12,
		"abb": 7,
		"abc": 5,
		"abbd": 3,
		"abba": 4,
	}
	actualSizeMap := make(map[string]int)
	tree.traverse(func(node *radixNode, depth int) bool {
		actualSizeMap[node.fullpath] = int(node.totalSize)
		return true
	})
	for prefix, expectSize := range expectSizeMap {
		actualSize := actualSizeMap[prefix]
		if expectSize != actualSize {
			t.Error("wrong size for " + prefix)
		}
	}
}