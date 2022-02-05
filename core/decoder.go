// Package core is RDB core core
package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/hdt3213/rdb/model"
	"io"
	"strconv"
	"time"
)

// Decoder is an instance of rdb parsing process
type Decoder struct {
	input     *bufio.Reader
	readCount int
	buffer    []byte
}

// NewDecoder create a new RDB core
func NewDecoder(reader io.Reader) *Decoder {
	parser := new(Decoder)
	parser.input = bufio.NewReader(reader)
	parser.buffer = make([]byte, 8)
	return parser
}

var magicNumber = []byte("REDIS")

const (
	minVersion = 1
	maxVersion = 9
)

const (
	opCodeIdle         = 248 /* LRU idle time. */
	opCodeFreq         = 249 /* LFU frequency. */
	opCodeAux          = 250 /* RDB aux field. */
	opCodeResizeDB     = 251 /* Hash table resize hint. */
	opCodeExpireTimeMs = 252 /* Expire time in milliseconds. */
	opCodeExpireTime   = 253 /* Old expire time in seconds. */
	opCodeSelectDB     = 254 /* DB number of the following keys. */
	opCodeEOF          = 255
)

const (
	typeString = iota
	typeList
	typeSet
	typeZset
	typeHash
	typeZset2 /* ZSET version 2 with doubles stored in binary. */
	typeModule
	typeModule2 // Module should import module entity, not support at present.
	_
	typeHashZipMap
	typeListZipList
	typeSetIntSet
	typeZsetZipList
	typeHashZipList
	typeListQuickList
	typeStreamListPacks
)

// checkHeader checks whether input has valid RDB file header
func (dec *Decoder) checkHeader() error {
	header := make([]byte, 9)
	err := dec.readFull(header)
	if err == io.EOF {
		return errors.New("empty file")
	}
	if err != nil {
		return fmt.Errorf("io error: %v", err)
	}
	if !bytes.Equal(header[0:5], magicNumber) {
		return errors.New("file is not a RDB file")
	}
	version, err := strconv.Atoi(string(header[5:]))
	if err != nil {
		return fmt.Errorf("%s is not valid version number", string(header[5:]))
	}
	if version < minVersion || version > maxVersion {
		return fmt.Errorf("cannot parse version: %d", version)
	}
	return nil
}

func (dec *Decoder) readObject(flag byte, base *model.BaseObject) (model.RedisObject, error) {
	switch flag {
	case typeString:
		bs, err := dec.readString()
		if err != nil {
			return nil, err
		}
		return &model.StringObject{
			BaseObject: base,
			Value:      bs,
		}, nil
	case typeList:
		list, err := dec.readList()
		if err != nil {
			return nil, err
		}
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeSet:
		set, err := dec.readSet()
		if err != nil {
			return nil, err
		}
		return &model.SetObject{
			BaseObject: base,
			Members:    set,
		}, nil
	case typeSetIntSet:
		set, err := dec.readIntSet()
		if err != nil {
			return nil, err
		}
		return &model.SetObject{
			BaseObject: base,
			Members:    set,
		}, nil
	case typeHash:
		hash, err := dec.readHashMap()
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject: base,
			Hash:       hash,
		}, nil
	case typeListZipList:
		list, err := dec.readZipList()
		if err != nil {
			return nil, err
		}
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeListQuickList:
		list, err := dec.readQuickList()
		if err != nil {
			return nil, err
		}
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeHashZipMap:
		m, err := dec.readZipMapHash()
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject: base,
			Hash:       m,
		}, nil
	case typeHashZipList:
		m, err := dec.readZipListHash()
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject: base,
			Hash:       m,
		}, nil
	case typeZset:
		entries, err := dec.readZSet(false)
		if err != nil {
			return nil, err
		}
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	case typeZset2:
		entries, err := dec.readZSet(true)
		if err != nil {
			return nil, err
		}
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	case typeZsetZipList:
		entries, err := dec.readZipListZSet()
		if err != nil {
			return nil, err
		}
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	}
	return nil, fmt.Errorf("unknown type flag: %b", flag)
}

func (dec *Decoder) parse(cb func(object model.RedisObject) bool) error {
	var dbIndex int
	var expireMs int64
	for {
		b, err := dec.readByte()
		if err != nil {
			return err
		}
		if b == opCodeEOF {
			break
		} else if b == opCodeSelectDB {
			dbIndex64, _, err := dec.readLength()
			if err != nil {
				return err
			}
			dbIndex = int(dbIndex64)
			continue
		} else if b == opCodeExpireTime {
			err = dec.readFull(dec.buffer)
			if err != nil {
				return err
			}
			expireMs = int64(binary.LittleEndian.Uint64(dec.buffer)) * 1000
			continue
		} else if b == opCodeExpireTimeMs {
			err = dec.readFull(dec.buffer)
			if err != nil {
				return err
			}
			expireMs = int64(binary.LittleEndian.Uint64(dec.buffer))
			continue
		} else if b == opCodeResizeDB {
			_, _, err := dec.readLength()
			if err != nil {
				return err
			}
			_, _, err = dec.readLength()
			if err != nil {
				err = errors.New("Parse Aux value failed: " + err.Error())
				break
			}
			// todo
			continue
		} else if b == opCodeAux {
			_, err := dec.readString()
			if err != nil {
				return err
			}
			_, err = dec.readString()
			if err != nil {
				err = errors.New("Parse Aux value failed: " + err.Error())
				break
			}
			// todo: return aux
			continue
		} else if b == opCodeFreq {
			_, err = dec.readByte()
			if err != nil {
				return err
			}
		} else if b == opCodeIdle {
			_, _, err = dec.readLength()
			if err != nil {
				return err
			}
		}
		begPos := dec.readCount
		key, err := dec.readString()
		if err != nil {
			return err
		}
		keySize := dec.readCount - begPos
		base := &model.BaseObject{
			DB:  dbIndex,
			Key: string(key),
		}
		if expireMs > 0 {
			expiration := time.Unix(0, expireMs*int64(time.Millisecond))
			base.Expiration = &expiration
		}
		begPos = dec.readCount
		obj, err := dec.readObject(b, base)
		if err != nil {
			return err
		}
		obj.SetSize(dec.readCount - begPos + keySize)
		toBeContinued := cb(obj)
		if !toBeContinued {
			break
		}
	}
	return nil
}

// Parse parses rdb and callback
func (dec *Decoder) Parse(cb func(object model.RedisObject) bool) error {
	err := dec.checkHeader()
	if err != nil {
		return err
	}
	return dec.parse(cb)
}
