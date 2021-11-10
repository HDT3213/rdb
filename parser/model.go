package parser

import (
	"encoding/json"
	"time"
)

const (
	// StringType is redis string
	StringType = "string"
	// StringType is redis list
	ListType = "list"
	// SetType is redis set
	SetType = "set"
	// HashType is redis hash
	HashType = "hash"
	// ZSetType is redis sorted set
	ZSetType = "zset"
)

// CallbackFunc process redis object
type CallbackFunc func(object RedisObject) bool

// RedisObject represents a redis object
type RedisObject interface {
	GetType() string
	GetKey() string
	GetDBIndex() int
	GetExpiration() *time.Time
}

// BaseObject is basement of redis object
type BaseObject struct {
	DB         int        `json:"db"`
	Key        string     `json:"key"`
	Expiration *time.Time `json:"expiration"` // expiration time, expiration of persistent object is nil
}

func (o *BaseObject) GetKey() string {
	return o.Key
}

func (o *BaseObject) GetDBIndex() int {
	return o.DB
}

func (o *BaseObject) GetExpiration() *time.Time {
	return o.Expiration
}

// StringObject stores a string object
type StringObject struct {
	*BaseObject
	Value []byte
}

// GetType returns redis object type
func (o *StringObject) GetType() string {
	return StringType
}

func (o *StringObject) MarshalJSON() ([]byte, error) {
	o2 := struct {
		*BaseObject
		Value string `json:"value"`
	}{
		BaseObject: o.BaseObject,
		Value:      string(o.Value),
	}
	return json.Marshal(o2)
}

// ListObject stores a list object
type ListObject struct {
	*BaseObject
	Values [][]byte
}

// GetType returns redis object type
func (o *ListObject) GetType() string {
	return ListType
}

func (o *ListObject) MarshalJSON() ([]byte, error) {
	values := make([]string, len(o.Values))
	for i, v := range o.Values {
		values[i] = string(v)
	}
	o2 := struct {
		*BaseObject
		Values []string `json:"values"`
	}{
		BaseObject: o.BaseObject,
		Values:     values,
	}
	return json.Marshal(o2)
}

// HashObject stores a hash object
type HashObject struct {
	*BaseObject
	Hash map[string][]byte
}

// GetType returns redis object type
func (o *HashObject) GetType() string {
	return HashType
}

func (o *HashObject) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	for k, v := range o.Hash {
		m[k] = string(v)
	}
	o2 := struct {
		*BaseObject
		Hash map[string]string `json:"hash"`
	}{
		BaseObject: o.BaseObject,
		Hash:       m,
	}
	return json.Marshal(o2)
}

// SetObject stores a set object
type SetObject struct {
	*BaseObject
	Members [][]byte
}

// GetType returns redis object type
func (o *SetObject) GetType() string {
	return SetType
}

func (o *SetObject) MarshalJSON() ([]byte, error) {
	values := make([]string, len(o.Members))
	for i, v := range o.Members {
		values[i] = string(v)
	}
	o2 := struct {
		*BaseObject
		Members []string `json:"members"`
	}{
		BaseObject: o.BaseObject,
		Members:    values,
	}
	return json.Marshal(o2)
}

// ZSetEntry is a key-score in sorted set
type ZSetEntry struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

// ZSetObject stores a sorted set object
type ZSetObject struct {
	*BaseObject
	Entries []*ZSetEntry
}

// GetType returns redis object type
func (o *ZSetObject) GetType() string {
	return ZSetType
}
