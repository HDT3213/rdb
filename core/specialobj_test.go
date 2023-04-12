package core

import (
	"github.com/hdt3213/rdb/model"
	"os"
	"path/filepath"
	"testing"
)

func TestWithSpecialObject(t *testing.T) {
	rdbFilename := filepath.Join("../cases", "memory.rdb")
	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		t.Errorf("open rdb %s failed, %v", rdbFilename, err)
		return
	}
	defer func() {
		_ = rdbFile.Close()
	}()
	expectAux := map[string]string{
		"redis-ver":    "6.0.6",
		"redis-bits":   "64",
		"ctime":        "1644136130",
		"used-mem":     "1167584",
		"aof-preamble": "0",
	}
	expectKeyCount := map[int]uint64{
		0: 7,
	}
	expectTTLCount := map[int]uint64{
		0: 1,
	}
	var auxCount, dbSizeCount int
	dec := NewDecoder(rdbFile).WithSpecialOpCode()
	err = dec.Parse(func(object model.RedisObject) bool {
		switch o := object.(type) {
		case *model.AuxObject:
			if o.GetType() != model.AuxType {
				t.Error("aux obj with wrong type")
			}
			expectValue := expectAux[o.Key]
			if o.Value != expectValue {
				t.Errorf("aux %s has wrong value", o.GetKey())
			}
			auxCount++
		case *model.DBSizeObject:
			dbSizeCount++
			if o.KeyCount != expectKeyCount[o.DB] {
				t.Errorf("db %d has wrong key count", o.DB)
			}
			if o.TTLCount != expectTTLCount[o.DB] {
				t.Errorf("db %d has wrong ttl count", o.DB)
			}
		}
		return true
	})
	if err != nil {
		t.Error(err)
	}
	if auxCount != len(expectAux) {
		t.Error("wrong aux object count")
	}
	if dbSizeCount != 1 {
		t.Error("wrong db size object count")
	}
}
