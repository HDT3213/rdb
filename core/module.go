package core

import (
	"errors"
	"fmt"
)

type Opcode uint8

const (
	ModuleOpcodeEOF Opcode = iota
	ModuleOpcodeSInt
	ModuleOpcodeUInt
	ModuleOpcodeFloat
	ModuleOpcodeDouble
	ModuleOpcodeString
)

type ModuleTypeHandler interface {
	ReadByte() (byte, error)
	ReadFull(buf []byte) error
	ReadOpcode() (Opcode, error)
	ReadUInt() (uint64, error)
	ReadSInt() (int64, error)
	ReadFloat32() (float32, error)
	ReadDouble() (float64, error)
	ReadString() ([]byte, error)
	ReadLength() (uint64, bool, error)
}

type moduleTypeHandlerImpl struct {
	dec *Decoder
}

func (m moduleTypeHandlerImpl) ReadByte() (byte, error) {
	return m.dec.readByte()
}

func (m moduleTypeHandlerImpl) ReadFull(buf []byte) error {
	return m.dec.readFull(buf)
}

func (m moduleTypeHandlerImpl) ReadOpcode() (Opcode, error) {
	code, _, err := m.dec.readLength()
	if err != nil {
		return 0, err
	}
	if code > 5 {
		return 0, errors.New("unknown opcode")
	}
	return Opcode(code), nil
}

func (m moduleTypeHandlerImpl) ReadUInt() (uint64, error) {
	val, _, err := m.dec.readLength()
	return val, err
}

func (m moduleTypeHandlerImpl) ReadSInt() (int64, error) {
	val, _, err := m.dec.readLength()
	return int64(val), err
}

func (m moduleTypeHandlerImpl) ReadFloat32() (float32, error) {
	return m.dec.readFloat32()
}
func (m moduleTypeHandlerImpl) ReadDouble() (float64, error) {
	return m.dec.readFloat()
}

func (m moduleTypeHandlerImpl) ReadString() ([]byte, error) {
	return m.dec.readString()
}

func (m moduleTypeHandlerImpl) ReadLength() (uint64, bool, error) {
	return m.dec.readLength()
}

type ModuleTypeHandleFunc func(handler ModuleTypeHandler, encVersion int) (interface{}, error)

func (dec *Decoder) readModuleType() (string, interface{}, error) {
	moduleId, _, err := dec.readLength()
	if err != nil {
		return "", nil, err
	}
	return dec.handleModuleType(moduleId)
}

func (dec *Decoder) handleModuleType(moduleId uint64) (string, interface{}, error) {
	moduleType := moduleTypeNameByID(moduleId)
	handler, found := dec.withSpecialTypes[moduleType]
	if !found {
		fmt.Printf("unknown module type: %s,will skip\n", moduleType)
		handler = skipModuleAuxData
	}
	encVersion := moduleTypeEncVersionByID(moduleId)
	val, err := handler(moduleTypeHandlerImpl{dec: dec}, int(encVersion))
	return moduleType, val, err
}

func moduleTypeNameByID(moduleId uint64) string {
	cset := ModuleTypeNameCharSet
	name := make([]byte, 9)
	moduleId >>= 10
	for j := 0; j < 9; j++ {
		name[8-j] = cset[moduleId&63]
		moduleId >>= 6
	}
	return string(name)
}

func moduleTypeEncVersionByID(moduleId uint64) uint64 {
	return moduleId & 1023
}

// skipModuleAuxData skips module aux data
func skipModuleAuxData(h ModuleTypeHandler, _ int) (interface{}, error) {
	opCode, err := h.ReadOpcode()
	if err != nil {
		return nil, err
	}
	for opCode != ModuleOpcodeEOF {
		switch opCode {
		case ModuleOpcodeSInt:
			_, err = h.ReadSInt()
		case ModuleOpcodeUInt:
			_, err = h.ReadUInt()
		case ModuleOpcodeFloat:
			_, err = h.ReadFloat32()
		case ModuleOpcodeDouble:
			_, err = h.ReadDouble()
		case ModuleOpcodeString:
			_, err = h.ReadString()
		default:
			err = fmt.Errorf("unknown module opcode %d", opCode)
		}
		if err != nil {
			return nil, err
		}
		opCode, err = h.ReadOpcode()
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

const ModuleTypeNameCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
