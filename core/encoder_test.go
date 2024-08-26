package core

import (
	"bytes"
	"github.com/hdt3213/rdb/model"
	"math/rand"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	auxMap := map[string]string{
		"redis-ver":    "4.0.6",
		"redis-bits":   "64",
		"aof-preamble": "0",
	}
	type valTTLPair struct {
		Value string
		TTL   uint64
	}
	strMap := map[string]*valTTLPair{
		"a": {Value: "a", TTL: uint64(time.Now().Add(time.Hour).Unix())},
		"b": {Value: "b", TTL: uint64(time.Now().Add(time.Minute).Unix())},
		"c": {Value: "c"},
		"1": {Value: "1"},
	}
	for i := 0; i < 1000; i++ {
		strMap[RandString(rand.Intn(20))] = &valTTLPair{
			Value: RandString(rand.Intn(50)),
		}
	}

	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	err := enc.WriteHeader()
	if err != nil {
		t.Error(err)
		return
	}
	for k, v := range auxMap {
		err = enc.WriteAux(k, v)
		if err != nil {
			t.Error(err)
			return
		}
	}
	var ttlCount uint64
	for _, v := range strMap {
		if v.TTL > 0 {
			ttlCount++
		}
	}
	err = enc.WriteDBHeader(0, uint64(len(strMap)), ttlCount)
	if err != nil {
		t.Error(err)
		return
	}

	for k, v := range strMap {
		var opts []interface{}
		if v.TTL > 0 {
			opts = append(opts, WithTTL(v.TTL))
		}
		err = enc.WriteStringObject(k, []byte(v.Value), opts...)
		if err != nil {
			t.Error(err)
			return
		}
	}
	err = enc.WriteEnd()
	if err != nil {
		t.Error(err)
		return
	}
	dec := NewDecoder(buf).WithSpecialOpCode()
	err = dec.Parse(func(object model.RedisObject) bool {
		switch o := object.(type) {
		case *model.StringObject:
			expect := strMap[o.GetKey()]
			if expect == nil {
				t.Errorf("unexpected object: %s", o.GetKey())
				return true
			}
			if expect.Value != string(o.Value) {
				t.Errorf("object: %s with wrong value", o.GetKey())
				return true
			}
			if o.GetExpiration() == nil {
				if expect.TTL > 0 {
					t.Errorf("object: %s with wrong ttl", o.GetKey())
					return true
				}
			} else {
				ttl := o.GetExpiration().UnixNano() / int64(time.Millisecond)
				if expect.TTL != uint64(ttl) {
					t.Errorf("object: %s with wrong ttl", o.GetKey())
					return true
				}
			}
		case *model.AuxObject:
			expect := auxMap[o.GetKey()]
			if expect == "" {
				t.Errorf("unexpected aux: %s", o.GetKey())
				return true
			}
			if expect != o.Value {
				t.Errorf("object: %s with wrong value", o.GetKey())
				return true
			}
		}
		return true
	})
	if err != nil {
		t.Error(err)
	}
}
