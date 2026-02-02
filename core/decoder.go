// Package core is RDB core core
package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hdt3213/rdb/memprofiler"
	"github.com/hdt3213/rdb/model"
)

// Decoder is an instance of rdb parsing process
type Decoder struct {
	input     *bufio.Reader
	readCount int
	buffer    []byte

	withSpecialOpCode bool
	withSpecialTypes  map[string]ModuleTypeHandleFunc

	valkey     bool
	rdbVersion int
}

// NewDecoder creates a new RDB decoder
func NewDecoder(reader io.Reader) *Decoder {
	parser := new(Decoder)
	parser.input = bufio.NewReader(reader)
	parser.buffer = make([]byte, 8)
	parser.withSpecialTypes = make(map[string]ModuleTypeHandleFunc)
	return parser
}

// WithSpecialOpCode enables returning model.AuxObject to callback
func (dec *Decoder) WithSpecialOpCode() *Decoder {
	dec.withSpecialOpCode = true
	return dec
}

// WithSpecialType enables returning redis module data structure to callback
func (dec *Decoder) WithSpecialType(moduleType string, f ModuleTypeHandleFunc) *Decoder {
	dec.withSpecialTypes[moduleType] = f
	return dec
}

var magicNumberRedis = []byte("REDIS")
var magicNumberValkey = []byte("VALKEY")

const (
	minVersion       = 1
	maxVersion       = 12
	minVersionValkey = 80
	maxVersionValkey = 80
)

const (
	opCodeSlotImport   = 243 /* Slot import state. */
	opCodeSlotInfo     = 244 /* Foreign slot info, safe to ignore. */
	opCodeFunction     = 245 /* function library data */
	opCodeModuleAux    = 247 /* Module auxiliary data. */
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
	typeModule2 // Module value parser should be registered with Decoder.WithSpecialType
	_
	typeHashZipMap
	typeListZipList
	typeSetIntSet
	typeZsetZipList
	typeHashZipList
	typeListQuickList
	typeStreamListPacks
	typeHashListPack
	typeZsetListPack
	typeListQuickList2
	typeStreamListPacks2
	typeSetListPack
	typeStreamListPacks3
	typeHashWithHfeRc         // rdb 12 (only redis 7.4 rc)
	typeHashListPackWithHfeRc // rdb 12 (only redis 7.4 rc)
	typeHashWithHfe           // since rdb 12 (redis 7.4)
	typeHashListPackWithHfe   // since rdb 12 (redis 7.4)

	typeHash2 = typeHashWithHfeRc // Hash with field-level expiration (Valkey 9+)
)

const (
	EB_EXPIRE_TIME_MAX     int64 = 0x0000FFFFFFFFFFFF
	EB_EXPIRE_TIME_INVALID int64 = EB_EXPIRE_TIME_MAX + 1
	HFE_MAX_ABS_TIME_MSEC  int64 = EB_EXPIRE_TIME_MAX >> 2
)

var encodingMap = map[int]string{
	typeString:                model.StringEncoding,
	typeList:                  model.ListEncoding,
	typeSet:                   model.SetEncoding,
	typeZset:                  model.ZSetEncoding,
	typeHash:                  model.HashEncoding,
	typeZset2:                 model.ZSet2Encoding,
	typeHashZipMap:            model.ZipMapEncoding,
	typeListZipList:           model.ZipListEncoding,
	typeSetIntSet:             model.IntSetEncoding,
	typeZsetZipList:           model.ZipListEncoding,
	typeHashZipList:           model.ZipListEncoding,
	typeListQuickList:         model.QuickListEncoding,
	typeStreamListPacks:       model.ListPackEncoding,
	typeStreamListPacks2:      model.ListPackEncoding,
	typeHashListPack:          model.ListPackEncoding,
	typeZsetListPack:          model.ListPackEncoding,
	typeListQuickList2:        model.QuickList2Encoding,
	typeSetListPack:           model.ListPackEncoding,
	typeHashWithHfeRc:         model.HashExEncoding,
	typeHashListPackWithHfeRc: model.ListPackExEncoding,
	typeHashWithHfe:           model.HashExEncoding,
	typeHashListPackWithHfe:   model.ListPackExEncoding,
	// typeHash2: model.HashExEncoding, // same 22 as typeHashWithHfeRc
}

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
	var versionString string
	if bytes.HasPrefix(header, magicNumberRedis) {
		dec.valkey = false
		versionString = string(bytes.TrimPrefix(header, magicNumberRedis))
	} else if bytes.HasPrefix(header, magicNumberValkey) {
		dec.valkey = true
		versionString = string(bytes.TrimPrefix(header, magicNumberValkey))
	} else {
		return errors.New("file is not a RDB file")
	}
	version, err := strconv.Atoi(versionString)
	if err != nil {
		return fmt.Errorf("%s is not valid version number", versionString)
	}
	if !dec.valkey && (version < minVersion || version > maxVersion) {
		return fmt.Errorf("cannot parse version: %d", version)
	}
	if dec.valkey && (version < minVersionValkey || version > maxVersionValkey) {
		return fmt.Errorf("cannot parse version: %d", version)
	}
	dec.rdbVersion = version
	return nil
}

func (dec *Decoder) readObject(flag byte, base *model.BaseObject) (model.RedisObject, error) {
	base.Encoding = encodingMap[int(flag)]
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
		set, extra, err := dec.readIntSet()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
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
		list, extra, err := dec.readQuickList()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
		return &model.ListObject{
			BaseObject: base,
			Values:     list,
		}, nil
	case typeListQuickList2:
		list, extra, err := dec.readQuickList2()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
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
		m, extra, err := dec.readZipListHash()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
		return &model.HashObject{
			BaseObject: base,
			Hash:       m,
		}, nil
	case typeHashListPack:
		m, extra, err := dec.readListPackHash()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
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
		entries, extra, err := dec.readZipListZSet()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	case typeZsetListPack:
		entries, extra, err := dec.readListPackZSet()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
		return &model.ZSetObject{
			BaseObject: base,
			Entries:    entries,
		}, nil
	case typeStreamListPacks, typeStreamListPacks2, typeStreamListPacks3:
		var version uint = 1
		if flag == typeStreamListPacks2 {
			version = 2
		} else if flag == typeStreamListPacks3 {
			version = 3
		}
		stream, err := dec.readStreamListPacks(version)
		if err != nil {
			return nil, err
		}
		stream.BaseObject = base
		return stream, nil
	case typeModule2:
		moduleType, val, err := dec.readModuleType()
		if err != nil {
			return nil, err
		}
		return &model.ModuleTypeObject{
			BaseObject: base,
			ModuleType: moduleType,
			Value:      val,
		}, nil
	case typeSetListPack:
		set, extra, err := dec.readListPackSet()
		if err != nil {
			return nil, err
		}
		base.Extra = extra
		return &model.SetObject{
			BaseObject: base,
			Members:    set,
		}, nil
	case typeHashWithHfe, typeHashWithHfeRc:
		hash, expire, err := dec.readHashMapEx(func() bool { return flag == typeHashWithHfeRc }())
		if err != nil {
			return nil, err
		}
		return &model.HashObject{
			BaseObject:       base,
			Hash:             hash,
			FieldExpirations: expire,
		}, nil
	case typeHashListPackWithHfe, typeHashListPackWithHfeRc:
		m, e, extra, err := dec.readListPackHashEx(func() bool { return flag == typeHashListPackWithHfeRc }())
		if err != nil {
			return nil, err
		}
		base.Extra = extra
		return &model.HashObject{
			BaseObject:       base,
			Hash:             m,
			FieldExpirations: e,
		}, nil
	}
	return nil, fmt.Errorf("unknown type flag: %b", flag)
}

func (dec *Decoder) parse(cb func(object model.RedisObject) bool) error {
	var dbIndex int
	var expireMs int64
	var lru int64 = -1
	var lfu int64 = -1
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
			err = dec.readFull(dec.buffer[:4])
			if err != nil {
				return err
			}
			expireMs = int64(binary.LittleEndian.Uint32(dec.buffer)) * 1000
			continue
		} else if b == opCodeExpireTimeMs {
			err = dec.readFull(dec.buffer)
			if err != nil {
				return err
			}
			expireMs = int64(binary.LittleEndian.Uint64(dec.buffer))
			continue
		} else if b == opCodeResizeDB {
			keyCount, _, err := dec.readLength()
			if err != nil {
				return err
			}
			ttlCount, _, err := dec.readLength()
			if err != nil {
				err = errors.New("Parse Aux value failed: " + err.Error())
				break
			}
			if dec.withSpecialOpCode {
				obj := &model.DBSizeObject{
					BaseObject: &model.BaseObject{},
				}
				obj.DB = dbIndex
				obj.KeyCount = keyCount
				obj.TTLCount = ttlCount
				tbc := cb(obj)
				if !tbc {
					break
				}
			}
			continue
		} else if b == opCodeAux {
			key, err := dec.readString()
			if err != nil {
				return err
			}
			value, err := dec.readString()
			if err != nil {
				err = errors.New("Parse Aux value failed: " + err.Error())
				break
			}
			if dec.withSpecialOpCode {
				obj := &model.AuxObject{
					BaseObject: &model.BaseObject{},
				}
				obj.Type = model.AuxType
				obj.Key = unsafeBytes2Str(key)
				obj.Value = unsafeBytes2Str(value)
				tbc := cb(obj)
				if !tbc {
					break
				}
			}
			continue
		} else if b == opCodeFreq {
			freq, err := dec.readByte()
			if err != nil {
				return err
			}
			lfu = int64(freq)
			continue
		} else if b == opCodeIdle {
			idle, _, err := dec.readLength()
			if err != nil {
				return err
			}
			lru = int64(idle)
			continue
		} else if b == opCodeModuleAux {
			_, _, err = dec.readModuleType()
			if err != nil {
				return err
			}
			continue
		} else if b == opCodeFunction {
			functionsLua, err := dec.readString()
			if err != nil {
				return err
			}
			if dec.withSpecialOpCode {
				obj := &model.FunctionsObject{
					BaseObject: &model.BaseObject{},
				}
				obj.Key = "functions"
				obj.Type = model.FunctionsType
				obj.Encoding = "functions"
				obj.FunctionsLua = unsafeBytes2Str(functionsLua)
				tbc := cb(obj)
				if !tbc {
					break
				}
			}
			continue
		} else if b == opCodeSlotInfo {
			var err error
			var slot_id, slot_size, expires_slot_size uint64
			slot_id, _, err = dec.readLength()
			if err == nil {
				slot_size, _, err = dec.readLength()
			}
			if err == nil {
				expires_slot_size, _, err = dec.readLength()
			}
			if err != nil {
				return err
			}
			_, _, _ = slot_id, slot_size, expires_slot_size // safe to skip
			continue
		} else if b == opCodeSlotImport {
			job, err := dec.readString()
			if err != nil {
				return err
			}
			num_slot_ranges, _, err := dec.readLength()
			if err != nil {
				return err
			}
			var slot_from, slot_to uint64
			ranges := make([]string, num_slot_ranges)
			for i := uint64(0); i < num_slot_ranges; i++ {
				slot_from, _, err = dec.readLength()
				if err == nil {
					slot_to, _, err = dec.readLength()
				}
				if err != nil {
					return err
				}
				ranges[i] = fmt.Sprintf("%d-%d", slot_from, slot_to)
			}
			_, _ = job, ranges // no way other than skipping
			continue
		}
		key, err := dec.readString()
		if err != nil {
			return err
		}
		base := &model.BaseObject{
			DB:  dbIndex,
			Key: unsafeBytes2Str(key),
		}
		if expireMs > 0 {
			expiration := time.Unix(0, expireMs*int64(time.Millisecond))
			base.Expiration = &expiration
			expireMs = 0 // reset expire ms
		}
		base.IdleTime = lru
		lru = -1 // reset lru
		base.Freq = lfu
		lfu = -1 // reset lfu
		obj, err := dec.readObject(b, base)
		if err != nil {
			return err
		}
		base.Size = memprofiler.SizeOfObject(obj)
		base.Type = obj.GetType()
		tbc := cb(obj)
		if !tbc {
			break
		}
	}
	// read crc64 at the end
	_ = dec.readFull(dec.buffer)
	return nil
}

// Parse parses rdb and callback
// cb returns true to continue, returns false to stop the iteration
func (dec *Decoder) Parse(cb func(object model.RedisObject) bool) (err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = fmt.Errorf("panic: %v", err2)
		}
	}()
	err = dec.checkHeader()
	if err != nil {
		return err
	}
	return dec.parse(cb)
}

func (dec *Decoder) GetReadCount() int {
	return dec.readCount
}
