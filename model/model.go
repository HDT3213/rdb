package model

import (
	"encoding/json"
	"time"
)

const (
	// StringType is redis string
	StringType = "string"
	// ListType is redis list
	ListType = "list"
	// SetType is redis set
	SetType = "set"
	// HashType is redis hash
	HashType = "hash"
	// ZSetType is redis sorted set
	ZSetType = "zset"
	// AuxType is redis metadata key-value pair
	AuxType = "aux"
	// DBSizeType is for RDB_OPCODE_RESIZEDB
	DBSizeType = "dbsize"
	// StreamType is a redis stream
	StreamType = "stream"
)

const (
	// StringEncoding for string
	StringEncoding = "string"
	// ListEncoding is formed by a length encoding and some string
	ListEncoding = "list"
	// SetEncoding is formed by a length encoding and some string
	SetEncoding = "set"
	// ZSetEncoding is formed by a length encoding and some string
	ZSetEncoding = "zset"
	// HashEncoding is formed by a length encoding and some string
	HashEncoding = "hash"
	// ZSet2Encoding is zset version2 which stores doubles in binary format
	ZSet2Encoding = "zset2"
	// ZipMapEncoding has been deprecated
	ZipMapEncoding = "zipmap"
	// ZipListEncoding  stores data in contiguous memory
	ZipListEncoding = "ziplist"
	// IntSetEncoding is a ordered list of integers
	IntSetEncoding = "intset"
	// QuickListEncoding is a list of ziplist
	QuickListEncoding = "quicklist"
	// ListPackEncoding is a new replacement for ziplist
	ListPackEncoding = "listpack"
	// QuickList2Encoding is a list of listpack
	QuickList2Encoding = "quicklist2"
)

// CallbackFunc process redis object
type CallbackFunc func(object RedisObject) bool

// RedisObject is interface for a redis object
type RedisObject interface {
	// GetType returns redis type of object: string/list/set/hash/zset
	GetType() string
	// GetKey returns key of object
	GetKey() string
	// GetDBIndex returns db index of object
	GetDBIndex() int
	// GetExpiration returns expiration time, expiration of persistent object is nil
	GetExpiration() *time.Time
	// GetSize returns rdb value size in Byte
	GetSize() int
	// GetElemCount returns number of elements in list/set/hash/zset
	GetElemCount() int
	// GetEncoding returns encoding of object
	GetEncoding() string

	DetermineFields([]string) RedisObject
}

// BaseObject is basement of redis object
type BaseObject struct {
	DB         int         `json:"db"`                   // DB is db index of redis object
	Key        string      `json:"key,omitempty"`        // Key is key of redis object
	Expiration *time.Time  `json:"expiration,omitempty"` // Expiration is expiration time, expiration of persistent object is nil
	Size       int         `json:"size,omitempty"`       // Size is rdb value size in Byte
	Type       string      `json:"type,omitempty"`       // Type is one of string/list/set/hash/zset
	Encoding   string      `json:"encoding,omitempty"`   // Encoding is the exact encoding method
	Extra      interface{} `json:"-"`                    // Extra stores more detail of encoding for memory profiler and other usages
}

func (o *BaseObject) GetType() string {
	return o.Type
}

// GetKey returns key of object
func (o *BaseObject) GetKey() string {
	return o.Key
}

// GetDBIndex returns db index of object
func (o *BaseObject) GetDBIndex() int {
	return o.DB
}

// GetEncoding returns encoding of object
func (o *BaseObject) GetEncoding() string {
	return o.Encoding
}

// GetExpiration returns expiration time, expiration of persistent object is nil
func (o *BaseObject) GetExpiration() *time.Time {
	return o.Expiration
}

// GetSize  returns rdb value size in Byte
func (o *BaseObject) GetSize() int {
	return o.Size
}

// GetElemCount returns number of elements in list/set/hash/zset
func (o *BaseObject) GetElemCount() int {
	return 0
}

func (o *BaseObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}

	var obj = &BaseObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				obj.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				obj.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				obj.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				obj.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				obj.Encoding = o.Encoding
			}
		}
	}
	return obj
}

// StringObject stores a string object
type StringObject struct {
	*BaseObject
	Value []byte `json:"value,omitempty"`
}

// GetType returns redis object type
func (o *StringObject) GetType() string {
	return StringType
}

// MarshalJSON marshal []byte as string
func (o *StringObject) MarshalJSON() ([]byte, error) {
	o2 := struct {
		*BaseObject
		Value string `json:"value,omitempty"`
	}{
		BaseObject: o.BaseObject,
		Value:      string(o.Value),
	}
	return json.Marshal(o2)
}

func (o *StringObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}
	var obj = &StringObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "value":
			if o.Value != nil {
				obj.Value = o.Value
			}
		}
	}
	obj.BaseObject = baseObject

	return obj
}

// ListObject stores a list object
type ListObject struct {
	*BaseObject
	Values [][]byte `json:"values,omitempty"`
}

// GetType returns redis object type
func (o *ListObject) GetType() string {
	return ListType
}

// GetElemCount returns number of elements in list/set/hash/zset
func (o *ListObject) GetElemCount() int {
	return len(o.Values)
}

// MarshalJSON marshal []byte as string
func (o *ListObject) MarshalJSON() ([]byte, error) {
	values := make([]string, len(o.Values))
	for i, v := range o.Values {
		values[i] = string(v)
	}
	o2 := struct {
		*BaseObject
		Values []string `json:"values,omitempty"`
	}{
		BaseObject: o.BaseObject,
		Values:     values,
	}
	return json.Marshal(o2)
}

func (o *ListObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}
	var obj = &ListObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "values":
			if o.Values != nil {
				obj.Values = o.Values
			}
		}
	}
	obj.BaseObject = baseObject

	return obj
}

// HashObject stores a hash object
type HashObject struct {
	*BaseObject
	Hash map[string][]byte `json:"hash,omitempty"`
}

// GetType returns redis object type
func (o *HashObject) GetType() string {
	return HashType
}

// GetElemCount returns number of elements in list/set/hash/zset
func (o *HashObject) GetElemCount() int {
	return len(o.Hash)
}

// MarshalJSON marshal []byte as string
func (o *HashObject) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	for k, v := range o.Hash {
		m[k] = string(v)
	}
	o2 := struct {
		*BaseObject
		Hash map[string]string `json:"hash,omitempty"`
	}{
		BaseObject: o.BaseObject,
		Hash:       m,
	}
	return json.Marshal(o2)
}

func (o *HashObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}

	var obj = &HashObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "hash":
			if o.Hash != nil {
				obj.Hash = o.Hash
			}
		}
	}
	obj.BaseObject = baseObject

	return obj
}

// SetObject stores a set object
type SetObject struct {
	*BaseObject
	Members [][]byte `json:"members,omitempty"`
}

// GetType returns redis object type
func (o *SetObject) GetType() string {
	return SetType
}

// GetElemCount returns number of elements in list/set/hash/zset
func (o *SetObject) GetElemCount() int {
	return len(o.Members)
}

// MarshalJSON marshal []byte as string
func (o *SetObject) MarshalJSON() ([]byte, error) {
	values := make([]string, len(o.Members))
	for i, v := range o.Members {
		values[i] = string(v)
	}
	o2 := struct {
		*BaseObject
		Members []string `json:"members,omitempty"`
	}{
		BaseObject: o.BaseObject,
		Members:    values,
	}
	return json.Marshal(o2)
}

func (o *SetObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}
	var obj = &SetObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "members":
			if o != nil && o.Members != nil {
				obj.Members = o.Members
			}
		}
	}
	obj.BaseObject = baseObject

	return obj
}

// ZSetEntry is a key-score in sorted set
type ZSetEntry struct {
	Member string  `json:"member,omitempty"`
	Score  float64 `json:"score,omitempty"`
}

// ZSetObject stores a sorted set object
type ZSetObject struct {
	*BaseObject
	Entries []*ZSetEntry `json:"entries,omitempty"`
}

// GetType returns redis object type
func (o *ZSetObject) GetType() string {
	return ZSetType
}

// GetElemCount returns number of elements in list/set/hash/zset
func (o *ZSetObject) GetElemCount() int {
	return len(o.Entries)
}

func (o *ZSetObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}
	var obj = &ZSetObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "entries":
			if o.Entries != nil {
				obj.Entries = o.Entries
			}
		}
	}
	obj.BaseObject = baseObject

	return obj
}

// AuxObject stores redis metadata
type AuxObject struct {
	*BaseObject
	Value string `json:"value,omitempty"`
}

// GetType returns redis object type
func (o *AuxObject) GetType() string {
	return AuxType
}

// MarshalJSON marshal []byte as string
func (o *AuxObject) MarshalJSON() ([]byte, error) {
	o2 := struct {
		*BaseObject
		Value string `json:"value,omitempty"`
	}{
		BaseObject: o.BaseObject,
		Value:      string(o.Value),
	}
	return json.Marshal(o2)
}

func (o *AuxObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}
	var obj = &AuxObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "value":
			if o.Value != "" {
				obj.Value = o.Value
			}
		}
	}

	obj.BaseObject = baseObject

	return obj
}

// DBSizeObject stores db size metadata
type DBSizeObject struct {
	*BaseObject
	KeyCount uint64
	TTLCount uint64
}

// GetType returns redis object type
func (o *DBSizeObject) GetType() string {
	return DBSizeType
}

// ModuleTypeObject stores a module type object parsed by custom handler
type ModuleTypeObject struct {
	*BaseObject
	ModuleType string      `json:"moduleType,omitempty"`
	Value      interface{} `json:"value,omitempty"`
}

// GetType returns module type name
func (o *ModuleTypeObject) GetType() string {
	return o.ModuleType
}

// MarshalJSON marshal []byte as string
func (o *ModuleTypeObject) MarshalJSON() ([]byte, error) {
	o2 := struct {
		*BaseObject
		ModuleType string      `json:"moduleType,omitempty"`
		Value      interface{} `json:"value,omitempty"`
	}{
		BaseObject: o.BaseObject,
		ModuleType: o.ModuleType,
		Value:      o.Value,
	}
	return json.Marshal(o2)
}

func (o *ModuleTypeObject) DetermineFields(fields []string) RedisObject {
	if o == nil {
		return nil
	}
	var baseObject = &BaseObject{DB: o.DB}

	var obj = &ModuleTypeObject{}
	for _, i := range fields {
		switch i {
		case "key":
			if o.Key != "" {
				baseObject.Key = o.Key
			}
		case "expiration":
			if o.Expiration != nil {
				baseObject.Expiration = o.Expiration
			}
		case "size":
			if o.Size != 0 {
				baseObject.Size = o.Size
			}
		case "type":
			if o.Type != "" {
				baseObject.Type = o.Type
			}
		case "encoding":
			if o.Encoding != "" {
				baseObject.Encoding = o.Encoding
			}
		case "value":
			if o.Value != nil {
				obj.Value = o.Value
			}
		case "moduleType":
			if o.ModuleType != "" {
				obj.ModuleType = o.ModuleType
			}
		}
	}
	obj.BaseObject = baseObject

	return obj
}
