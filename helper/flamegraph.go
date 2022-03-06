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

func FlameGraph(rdbFilename string, port int, separator string, maxDepth int) (chan<- struct{}, error) {
	if rdbFilename == "" {
		return nil, errors.New("src file path is required")
	}
	if separator == "" {
		separator = ":"
	}
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return nil, fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	p := core.NewDecoder(rdbFile)
	root := &d3flame.FlameItem{
		Children: make(map[string]*d3flame.FlameItem),
	}
	err = p.Parse(func(object model.RedisObject) bool {
		parts := strings.Split(object.GetKey(), separator)
		node := root
		parts = append([]string{"db:" + strconv.Itoa(object.GetDBIndex())}, parts...)
		for _, part := range parts {
			if node.Children[part] == nil {
				node.Children[part] = &d3flame.FlameItem{
					Name:     part,
					Children: make(map[string]*d3flame.FlameItem),
				}
			}
			node = node.Children[part]
			node.Value += object.GetSize()
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	return d3flame.Web(data, port), nil
}
