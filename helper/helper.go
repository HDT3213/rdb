package helper

import (
	"encoding/json"
	"fmt"
	"github.com/hdt3213/rdb/model"
	"github.com/hdt3213/rdb/parser"
	"os"
)

// ToJsons read rdb file and convert to json file whose each line contains a json object
func ToJsons(rdbFilename string, jsonFilename string) error {
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	jsonFile, err := os.Create(jsonFilename)
	if err != nil {
		return fmt.Errorf("create json %s failed, %v", jsonFilename, err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()
	p := parser.NewParser(rdbFile)
	return p.Parse(func(object model.RedisObject) bool {
		data, err := json.Marshal(object)
		if err != nil {
			fmt.Printf("json marshal failed: %v", err)
			return true
		}
		_, err = jsonFile.Write(data)
		if err != nil {
			fmt.Printf("write failed: %v", err)
			return true
		}
		_, err = jsonFile.WriteString("\n")
		if err != nil {
			fmt.Printf("write failed: %v", err)
			return true
		}
		return true
	})
}
