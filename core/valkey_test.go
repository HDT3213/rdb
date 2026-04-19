package core

import (
	"bytes"
	"testing"

	"github.com/hdt3213/rdb/model"
)

func TestValkeyClusterMetadataOpcodes(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoderValkey(buf)
	if err := enc.WriteHeader(); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if err := enc.WriteDBHeader(0, 1, 0); err != nil {
		t.Fatalf("write db header: %v", err)
	}

	if err := enc.write([]byte{opCodeSlotInfo}); err != nil {
		t.Fatalf("write slot info opcode: %v", err)
	}
	if err := enc.writeLength(42); err != nil {
		t.Fatalf("write slot id: %v", err)
	}
	if err := enc.writeLength(3); err != nil {
		t.Fatalf("write slot size: %v", err)
	}
	if err := enc.writeLength(1); err != nil {
		t.Fatalf("write expires slot size: %v", err)
	}

	if err := enc.write([]byte{opCodeSlotImport}); err != nil {
		t.Fatalf("write slot import opcode: %v", err)
	}
	if err := enc.writeString("job-1"); err != nil {
		t.Fatalf("write slot import job: %v", err)
	}
	if err := enc.writeLength(2); err != nil {
		t.Fatalf("write slot range count: %v", err)
	}
	if err := enc.writeLength(10); err != nil {
		t.Fatalf("write slot range start: %v", err)
	}
	if err := enc.writeLength(11); err != nil {
		t.Fatalf("write slot range end: %v", err)
	}
	if err := enc.writeLength(20); err != nil {
		t.Fatalf("write slot range start: %v", err)
	}
	if err := enc.writeLength(25); err != nil {
		t.Fatalf("write slot range end: %v", err)
	}

	if err := enc.WriteStringObject("cluster-key", []byte("value")); err != nil {
		t.Fatalf("write string object: %v", err)
	}
	if err := enc.WriteEnd(); err != nil {
		t.Fatalf("write end: %v", err)
	}

	var objects []model.RedisObject
	dec := NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := dec.Parse(func(object model.RedisObject) bool {
		objects = append(objects, object)
		return true
	}); err != nil {
		t.Fatalf("parse valkey cluster rdb: %v", err)
	}

	if len(objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(objects))
	}
	strObj, ok := objects[0].(*model.StringObject)
	if !ok {
		t.Fatalf("expected string object, got %T", objects[0])
	}
	if strObj.Key != "cluster-key" {
		t.Fatalf("unexpected key: %s", strObj.Key)
	}
	if string(strObj.Value) != "value" {
		t.Fatalf("unexpected value: %s", strObj.Value)
	}
}
