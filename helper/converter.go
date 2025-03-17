package helper

import (
	"errors"
	"fmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
	"os"
)

// ToAOF read rdb file and convert to aof file (Redis Serialization )
func ToAOF(rdbFilename string, aofFilename string, options ...interface{}) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
	}
	if aofFilename == "" {
		return errors.New("output file path is required")
	}
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	aofFile, err := os.Create(aofFilename)
	if err != nil {
		return fmt.Errorf("create json %s failed, %v", aofFilename, err)
	}
	defer func() {
		_ = aofFile.Close()
	}()

	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return err
	}
	return dec.Parse(func(object model.RedisObject) bool {
		cmdLines := ObjectToCmd(object, options...)
		data := CmdLinesToResp(cmdLines)
		_, err = aofFile.Write(data)
		if err != nil {
			fmt.Printf("write failed: %v", err)
			return true
		}
		return true
	})
}
