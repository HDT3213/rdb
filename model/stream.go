package model

import "strconv"

// StreamObject stores a stream object
type StreamObject struct {
	*BaseObject
	// IsV2 means is RDB_TYPE_STREAM_LISTPACKS_2`
	IsV2 bool `json:"isV2,omitempty"`
	// Entries stores messages in stream
	Entries []*StreamEntry `json:"entries,omitempty"`
	// Groups is consumer groups of stream
	Groups []*StreamGroup `json:"groups,omitempty"`
	// Length is current number of elements inside this stream
	Length uint64 `json:"len"`
	// LastId is the ID of last entry in stream
	LastId *StreamId `json:"lastId"`
	// FirstId is the ID of first entry in stream. only valid in V2
	FirstId *StreamId `json:"firstId,omitempty"`
	// MaxDeletedId is the maximal ID that was deleted in stream. only valid in V2
	MaxDeletedId *StreamId `json:"maxDeletedId,omitempty"`
	// AddedEntriesCount is count of elements added in all time. only valid in V2
	AddedEntriesCount uint64 `json:"addedEntriesCount,omitempty"`
}

func (obj *StreamObject) GetType() string {
	return StreamType
}

// StreamEntry is a group of messages in stream
// Actually, it is a node of radix tree which may contains
type StreamEntry struct {
	FirstMsgId *StreamId        `json:"firstMsgId"`
	Fields     []string         `json:"fields"`
	Msgs       []*StreamMessage `json:"msgs"`
}

// StreamMessage is a message item in stream
type StreamMessage struct {
	Id      *StreamId         `json:"id"`
	Fields  map[string]string `json:"fields"`
	Deleted bool              `json:"deleted"`
}

// StreamId is a 128-bit number composed of a milliseconds time and  a sequence counter
type StreamId struct {
	Ms       uint64 `json:"ms"`
	Sequence uint64 `json:"sequence"`
}

func (id *StreamId) MarshalText() (text []byte, err error) {
	txt := strconv.FormatUint(id.Ms, 10) + "-" + strconv.FormatUint(id.Sequence, 10)
	return []byte(txt), nil
}

// StreamGroup is a consumer group
type StreamGroup struct {
	Name      string            `json:"name"`
	LastId    *StreamId         `json:"lastId"`
	Pending   []*StreamNAck     `json:"pending,omitempty"`
	Consumers []*StreamConsumer `json:"consumers,omitempty"`
	// EntriesRead is only valid in V2. The following comments are from redis to illustrate its use
	/*In a perfect world (CG starts at 0-0, no dels, no XGROUP SETID, ...), this is the total number of
	group reads. In the real world, the reasoning behind this value is detailed at the top comment of
	streamEstimateDistanceFromFirstEverEntry(). */
	EntriesRead uint64 `json:"entriesRead,omitempty"`
}

// StreamNAck points a pending message
type StreamNAck struct {
	Id            *StreamId `json:"id"`
	DeliveryTime  uint64    `json:"deliveryTime"`
	DeliveryCount uint64    `json:"deliveryCount"`
}

// StreamConsumer is a consumer
type StreamConsumer struct {
	Name     string      `json:"name"`
	SeenTime uint64      `json:"seenTime"`
	Pending  []*StreamId `json:"pending,omitempty"`
}
