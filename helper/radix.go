package helper

import (
	"container/list"
	"strconv"
	"strings"
)

type radixNode struct {
	path      string
	end       bool
	children  []*radixNode
	totalSize int // total size of all key-value with this prefix
	keyCount  int
	fullpath  string
}

type radixTree struct{
	root *radixNode
}

func newRadixTree() *radixTree {
	return &radixTree{
		root: &radixNode{},
	}
}

func commonPrefixLen(wordA, wordB string) int {
	var i int
	for i < len(wordA) && i < len(wordB) && wordA[i] == wordB[i] {
		i++
	}
	return i
}

func (tree *radixTree) insert(word string, size int) {
	root := tree.root
	fullword := word
	node := root
walk:
	for {
		i := commonPrefixLen(word, node.path)
		if i < len(node.path) {
			// assert: node == root || i > 0, because it is the first loop or from `for _, child := range node.children`
			// split current node `rn`
			newChild := &radixNode{
				path:      node.path[i:],
				children:  node.children,
				end:       node.end,
				totalSize: node.totalSize,
				keyCount:  node.keyCount,
				fullpath:  node.fullpath,
			}
			node.children = []*radixNode{newChild}
			node.fullpath = node.fullpath[:len(node.fullpath)-(len(node.path)-i)]
			node.path = node.path[:i]
			node.end = false
		}
		// word must be a descendants of node
		if i > 0 || node == root {
			node.totalSize += size
			node.keyCount++
		}
		if i == len(word) {
			// assert node.fullpath == fullword
			node.end = true
			break
		}
		if i < len(word) {
			word = word[i:]
			// word may have common prefix with a child, recurse until no common prefix
			c := word[0]
			for _, child := range node.children {
				if child.path[0] == c {
					node = child
					continue walk
				}
			}

			// now, word has no common prefix with child
			child := &radixNode{}
			child.path = word
			child.end = true
			child.fullpath = fullword
			child.totalSize = size
			child.keyCount = 1
			node.children = append(node.children, child)
			return
		}
	}
}

func (tree *radixTree) traverse(cb func(node *radixNode, depth int) bool) {
	root := tree.root
	type nodeDepth struct {
		node  *radixNode
		depth int
	}
	queue := list.New()
	queue.PushBack(&nodeDepth{root, 1})
	for queue.Len() > 0 {
		head := queue.Front()
		queue.Remove(head)
		node2 := head.Value.(*nodeDepth)
		if !cb(node2.node, node2.depth) {
			return
		}
		depth := node2.depth + 1
		for _, child := range node2.node.children {
			queue.PushBack(&nodeDepth{child, depth})
		}
	}
}

func (n *radixNode) GetSize() int {
	return int(n.totalSize)
}

func genKey(db int, key string) string {
	return strconv.Itoa(db) + " " + key
}

func parseNodeKey(key string) (int, string) {
	if key == "" {
		return -1, ""
	}
	index := strings.Index(key, " ")
	db, _ := strconv.Atoi(key[:index])
	var realKey string
	// if key is db root, index+1 == len(key)
	if index+1 < len(key) {
		realKey = key[index+1:]
	}
	return db, realKey
}