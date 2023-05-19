package core

import (
	"bytes"
	"fmt"
	"github.com/hdt3213/rdb/model"
	"testing"
)

const testModuleType = "test-type"

func TestModuleType(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)

	key := "testkey"
	expectedModuleEncVersion := uint64(42)
	expectedStrData := "testdata123"
	expectedUInt := uint64(123)

	err := enc.WriteHeader()
	if err != nil {
		t.Error(err)
		return
	}

	err = enc.WriteDBHeader(0, 1, 0)
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.beforeWriteObject()
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.write([]byte{typeModule2})
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.writeString(key)
	if err != nil {
		t.Error(err)
		return
	}

	moduleId := createModuleId(testModuleType, expectedModuleEncVersion)
	err = enc.writeLength(moduleId)
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.writeLength(uint64(ModuleOpcodeString))
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.writeString(expectedStrData)
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.writeLength(uint64(ModuleOpcodeUInt))
	if err != nil {
		t.Error(err)
		return
	}
	err = enc.writeLength(expectedUInt)
	if err != nil {
		t.Error(err)
		return
	}

	err = enc.writeLength(uint64(ModuleOpcodeEOF))
	if err != nil {
		t.Error(err)
		return
	}
	enc.state = writtenObjectState

	err = enc.WriteEnd()
	if err != nil {
		t.Error(err)
		return
	}

	expectedResult := "expected-result"

	dec := NewDecoder(buf).WithSpecialType(testModuleType,
		func(h ModuleTypeHandler, encVersion int) (interface{}, error) {
			if encVersion != 42 {
				t.Errorf("invalid encoding version, expected %d, actual %d",
					expectedModuleEncVersion, encVersion)
				return nil, fmt.Errorf("invalid encoding version, expected %d, actual %d",
					expectedModuleEncVersion, encVersion)
			}

			opcode, err := h.ReadOpcode()
			if err != nil {
				return nil, err
			}
			if opcode != ModuleOpcodeString {
				return nil, fmt.Errorf("invalid opcode read, expected %d (string), actual %d",
					ModuleOpcodeString, opcode)
			}
			data, err := h.ReadString()
			if err != nil {
				return nil, err
			}
			if !bytes.Equal(data, []byte(expectedStrData)) {
				return nil, fmt.Errorf("invalid string data read, expected %s, actual %s",
					expectedStrData, string(data))
			}

			opcode, err = h.ReadOpcode()
			if err != nil {
				return nil, err
			}
			if opcode != ModuleOpcodeUInt {
				return nil, fmt.Errorf("invalid opcode read, expected %d (uint), actual %d",
					ModuleOpcodeUInt, opcode)
			}
			val, err := h.ReadUInt()
			if err != nil {
				return nil, err
			}
			if val != expectedUInt {
				return nil, fmt.Errorf("invalid unsigned int read, expected %d, actual %d",
					expectedUInt, val)
			}
			opcode, err = h.ReadOpcode()
			if err != nil {
				return nil, err
			}
			if opcode != ModuleOpcodeEOF {
				return nil, fmt.Errorf("invalid opcode read, expected %d (EOF), actual %d",
					ModuleOpcodeEOF, opcode)
			}
			return expectedResult, nil
		})

	err = dec.Parse(func(o model.RedisObject) bool {
		if o.GetKey() != key {
			t.Errorf("invalid object key, expected %s, actual %s", key, o.GetKey())
			return false
		}
		if o.GetType() != testModuleType {
			t.Errorf("invalid redis type, expected %s, actual %s", testModuleType, o.GetType())
			return false
		}
		mtObj, ok := o.(*model.ModuleTypeObject)
		if !ok {
			t.Errorf("invalid object type, expected model.ModuleTypeObject")
			return false
		}

		if mtObj.Value != expectedResult {
			t.Errorf("invalid return value")
			return false
		}

		return true
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCorrectModuleTypeEncodeDecode(t *testing.T) {
	moduleId := createModuleId(testModuleType, 42)
	name := moduleTypeNameByID(moduleId)
	encVersion := moduleTypeEncVersionByID(moduleId)
	if name != testModuleType {
		t.Errorf("invalid module type name, expected %s, actual %s", testModuleType, name)
	}
	if encVersion != 42 {
		t.Errorf("invalid module encoding version, expected %d, actual %d", 42, encVersion)
	}
}

func createModuleId(moduleType string, encVersion uint64) uint64 {
	moduleId := uint64(0)
	for i := 0; i < 9; i++ {
		moduleId |= uint64(charCode(moduleType[i]))
		moduleId <<= 6
	}
	moduleId <<= 4
	moduleId |= encVersion
	return moduleId
}

func charCode(c uint8) uint8 {
	for i, csChar := range []byte(ModuleTypeNameCharSet) {
		if c == csChar {
			return uint8(i)
		}
	}
	panic(fmt.Errorf("unsupported char %c", c))
}
