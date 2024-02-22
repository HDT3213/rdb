package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

// PrefixAnalyse read rdb file and find the largest N keys.
// The invoker owns output, FindBiggestKeys won't close it
func PrefixAnalyse(rdbFilename string, topN int, maxDepth int, output *os.File, options ...interface{}) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
	}
	if topN < 0 {
		return errors.New("n must greater than 0")
	} else if topN == 0 {
		topN = math.MaxInt
	}
	if maxDepth == 0 {
		maxDepth = math.MaxInt
	} else {
		maxDepth += 2 // for root(depth==1) and database root(depth==2)
	}

	// decode rdb file
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return err
	}

	// prefix tree
	tree := newRadixTree()
	err = dec.Parse(func(object model.RedisObject) bool {
		key := genKey(object.GetDBIndex(), object.GetKey())
		tree.insert(key, object.GetSize())
		return true
	})
	if err != nil {
		return err
	}
	
	// get top list
	toplist := newToplist(topN)
	tree.traverse(func(node *radixNode, depth int) bool {
		if depth > maxDepth {
			return false
		}
		if depth <= 2 {
			// skip root and database root
			return true
		}
		toplist.add(node)
		return true
	})

	// write into csv
	_, err = output.WriteString("database,prefix,size,size_readable,key_count\n")
	if err != nil {
		return fmt.Errorf("write header failed: %v", err)
	}
	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()
	printNode := func(node *radixNode) error {
		db, key := parseNodeKey(node.fullpath)
		dbStr := strconv.Itoa(db)
		return csvWriter.Write([]string{
			dbStr,
			key,
			strconv.Itoa(node.totalSize),
			bytefmt.FormatSize(uint64(node.totalSize)),
			strconv.Itoa(node.keyCount),
		})
	}
	for _, n := range toplist.list {
		node := n.(*radixNode)
		err := printNode(node)
		if err != nil {
			return err
		}
	}

	return nil
}
