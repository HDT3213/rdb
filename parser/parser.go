// Package parser is RDB parser core
package parser

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

// Parser is an instance of rdb parsing process
type Parser struct {
	input     *bufio.Reader
	readCount int
	buffer    []byte
}

// NewParser create a new RDB parser
func NewParser(reader io.Reader) *Parser {
	parser := new(Parser)
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
func (p *Parser) checkHeader() error {
	header := make([]byte, 9)
	err := p.readFull(header)
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

func (p *Parser) readObject(flag byte, base *model.BaseObject) (model.RedisObject, error) {
	switch flag {
	case typeString:
		bs, err := p.readString()
		if err != nil {
			return nil, err
		}
		return &model.StringObject{
			BaseObject: base,
			Value:      bs,
		}, nil
	case typeList:
		list, err := p.readList()
		if err != nil {
			return nil, err
		}
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeSet:
		set, err := p.readSet()
		if err != nil {
			return nil, err
		}
		return &model.SetObject{
			BaseObject: base,
			Members:    set,
		}, nil
	case typeSetIntSet:
		set, err := p.readIntSet()
		if err != nil {
			return nil, err
		}
		return &model.SetObject{
			BaseObject: base,
			Members:    set,
		}, nil
	case typeHash:
		hash, err := p.readHashMap()
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject: base,
			Hash:       hash,
		}, nil
	case typeListZipList:
		list, err := p.readZipList()
		if err != nil {
			return nil, err
		}
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeListQuickList:
		list, err := p.readQuickList()
		if err != nil {
			return nil, err
		}
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeHashZipMap:
		m, err := p.readZipMapHash()
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject: base,
			Hash:       m,
		}, nil
	case typeHashZipList:
		m, err := p.readZipListHash()
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject: base,
			Hash:       m,
		}, nil
	case typeZset:
		entries, err := p.readZSet(false)
		if err != nil {
			return nil, err
		}
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	case typeZset2:
		entries, err := p.readZSet(true)
		if err != nil {
			return nil, err
		}
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	case typeZsetZipList:
		entries, err := p.readZipListZSet()
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

func (p *Parser) parse(cb func(object model.RedisObject) bool) error {
	var dbIndex int
	var expireMs int64
	for {
		b, err := p.readByte()
		if err != nil {
			return err
		}
		if b == opCodeEOF {
			break
		} else if b == opCodeSelectDB {
			dbIndex64, _, err := p.readLength()
			if err != nil {
				return err
			}
			dbIndex = int(dbIndex64)
			continue
		} else if b == opCodeExpireTime {
			err = p.readFull(p.buffer)
			if err != nil {
				return err
			}
			expireMs = int64(binary.LittleEndian.Uint64(p.buffer)) * 1000
			continue
		} else if b == opCodeExpireTimeMs {
			err = p.readFull(p.buffer)
			if err != nil {
				return err
			}
			expireMs = int64(binary.LittleEndian.Uint64(p.buffer))
			continue
		} else if b == opCodeResizeDB {
			_, _, err := p.readLength()
			if err != nil {
				return err
			}
			_, _, err = p.readLength()
			if err != nil {
				err = errors.New("Parse Aux value failed: " + err.Error())
				break
			}
			// todo
			continue
		} else if b == opCodeAux {
			_, err := p.readString()
			if err != nil {
				return err
			}
			_, err = p.readString()
			if err != nil {
				err = errors.New("Parse Aux value failed: " + err.Error())
				break
			}
			// todo: return aux
			continue
		} else if b == opCodeFreq {
			_, err = p.readByte()
			if err != nil {
				return err
			}
		} else if b == opCodeIdle {
			_, _, err = p.readLength()
			if err != nil {
				return err
			}
		}
		begPos := p.readCount
		key, err := p.readString()
		if err != nil {
			return err
		}
		keySize := p.readCount - begPos
		base := &model.BaseObject{
			DB:  dbIndex,
			Key: string(key),
		}
		if expireMs > 0 {
			expiration := time.Unix(0, expireMs*int64(time.Millisecond))
			base.Expiration = &expiration
		}
		begPos = p.readCount
		obj, err := p.readObject(b, base)
		if err != nil {
			return err
		}
		obj.SetSize(p.readCount - begPos + keySize)
		toBeContinued := cb(obj)
		if !toBeContinued {
			break
		}
	}
	return nil
}

// Parse parses rdb and callback
func (p *Parser) Parse(cb func(object model.RedisObject) bool) error {
	err := p.checkHeader()
	if err != nil {
		return err
	}
	return p.parse(cb)
}
