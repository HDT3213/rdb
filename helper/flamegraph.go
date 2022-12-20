package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/d3flame"
	"github.com/hdt3213/rdb/model"
	"os"
	"strconv"
	"strings"
)

// TrimThreshold is the min count of keys to enable trim
var TrimThreshold = 1000

// FlameGraph draws flamegraph in web page to analysis memory usage pattern
func FlameGraph(rdbFilename string, port int, separators []string, options ...interface{}) (chan<- struct{}, error) {
	if rdbFilename == "" {
		return nil, errors.New("src file path is required")
	}
	if port == 0 {
		port = 16379 // default port
	}
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return nil, fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()

	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return nil, err
	}
	root := &d3flame.FlameItem{
		Name:     "root",
		Children: make(map[string]*d3flame.FlameItem),
	}
	var count int
	err = dec.Parse(func(object model.RedisObject) bool {
		count++
		addObject(root, separators, object)
		return true
	})
	if err != nil {
		return nil, err
	}
	totalSize := 0
	for _, v := range root.Children {
		totalSize += v.Value
	}
	root.Value = totalSize
	if count >= TrimThreshold {
		trimData(root)
	}
	data, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	return d3flame.Web(data, port), nil
}

func split(s string, separators []string) []string {
	sep := ":"
	if len(separators) > 0 {
		sep = separators[0]
	}
	for i := 1; i < len(separators); i++ {
		s = strings.ReplaceAll(s, separators[i], sep)
	}
	return strings.Split(s, sep)
}

func addObject(root *d3flame.FlameItem, separators []string, object model.RedisObject) {
	node := root
	parts := split(object.GetKey(), separators)
	parts = append([]string{"db:" + strconv.Itoa(object.GetDBIndex())}, parts...)
	for _, part := range parts {
		if node.Children[part] == nil {
			n := &d3flame.FlameItem{
				Name:     part,
				Children: make(map[string]*d3flame.FlameItem),
			}
			node.AddChild(n)
		}
		node = node.Children[part]
		node.Value += object.GetSize()
	}
}

// bigNodeThreshold is the min size
var bigNodeThreshold = 1024 * 1024 // 1MB

func trimData(root *d3flame.FlameItem) {
	// trim long tail
	queue := []*d3flame.FlameItem{
		root,
	}
	for len(queue) > 0 {
		// Aggregate leaf nodes
		node := queue[0]
		queue = queue[1:]
		leafSum := 0
		for key, child := range node.Children {
			if len(child.Children) == 0 && child.Value < bigNodeThreshold { // child is a leaf node
				delete(node.Children, key) // remove small leaf keys
				leafSum += child.Value
			}
			queue = append(queue, child) // reserve big key
		}
		if leafSum > 0 {
			n := &d3flame.FlameItem{
				Name:  "others",
				Value: leafSum,
			}
			node.AddChild(n)
		}
	}
}
