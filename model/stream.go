package model

import (
	"sort"
	"strconv"
)

// StreamObject stores a stream object
type StreamObject struct {
	*BaseObject
	// Version 2 means is RDB_TYPE_STREAM_LISTPACKS_2, 3 means is RDB_TYPE_STREAM_LISTPACKS_3
	Version uint `json:"version,omitempty"`
	// Entries stores messages in stream
	Entries []*StreamEntry `json:"entries,omitempty"`
	// Groups is consumer groups of stream
	Groups []*StreamGroup `json:"groups,omitempty"`
	// Length is current number of messages inside this stream
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

// StreamEntry is a node in the underlying radix tree of redis stream, of type listpacks, which contains several messages.
// There is no need to care about which entry the message belongs to when using it.
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
	return []byte(streamIdKey(id)), nil
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
	Name       string      `json:"name"`
	SeenTime   uint64      `json:"seenTime"`
	Pending    []*StreamId `json:"pending,omitempty"`
	ActiveTime uint64      `json:"activeTime"` // since rdb11, see https://github.com/redis/redis/pull/11099
}

func streamIdKey(id *StreamId) string {
	return strconv.FormatUint(id.Ms, 10) + "-" + strconv.FormatUint(id.Sequence, 10)
}

func (id *StreamId) isAfter(other *StreamId) bool {
	if other == nil {
		return true
	}
	if id.Ms != other.Ms {
		return id.Ms > other.Ms
	}
	return id.Sequence > other.Sequence
}

// PendingOwners maps every distinct id in group.Pending, written as <ms>-<sequence>, to the sorted names of
// the consumers whose own pending list holds it. An id that no consumer holds maps to an empty slice.
// Ids in a consumer's list that are not pending entries of the group are ignored.
func (group *StreamGroup) PendingOwners() map[string][]string {
	owners := make(map[string][]string, len(group.Pending))
	for _, nack := range group.Pending {
		key := streamIdKey(nack.Id)
		if _, ok := owners[key]; !ok {
			owners[key] = []string{}
		}
	}
	for _, consumer := range group.Consumers {
		for _, id := range consumer.Pending {
			key := streamIdKey(id)
			names, ok := owners[key]
			if !ok {
				continue
			}
			held := false
			for _, name := range names {
				if name == consumer.Name {
					held = true
					break
				}
			}
			if !held {
				owners[key] = append(names, consumer.Name)
			}
		}
	}
	for key := range owners {
		sort.Strings(owners[key])
	}
	return owners
}

// UndeliveredMessages returns the live messages that group has never been delivered, ordered by ascending id:
// those with an id greater than group.LastId and no pending entry in the group.
// If the group has no LastId, every message that is not deleted and not pending is returned.
func (obj *StreamObject) UndeliveredMessages(group *StreamGroup) []*StreamMessage {
	var lastId *StreamId
	held := map[string]bool{}
	if group != nil {
		lastId = group.LastId
		for _, nack := range group.Pending {
			held[streamIdKey(nack.Id)] = true
		}
	}
	result := make([]*StreamMessage, 0)
	for _, entry := range obj.Entries {
		for _, msg := range entry.Msgs {
			if msg.Deleted || !msg.Id.isAfter(lastId) || held[streamIdKey(msg.Id)] {
				continue
			}
			result = append(result, msg)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[j].Id.isAfter(result[i].Id)
	})
	return result
}

// ConsumerBacklog is one row of the backlog report of a consumer group.
// Undelivered and Dangling are numbers of the whole group, so every row of a group carries the same pair.
type ConsumerBacklog struct {
	Group string `json:"group"`
	// Consumer is empty for the row that collects the pending entries no consumer holds
	Consumer string `json:"consumer"`
	// Pending is the number of pending entries the consumer holds
	Pending int `json:"pending"`
	// OldestDeliveryTime is the smallest delivery time among the pending entries
	OldestDeliveryTime uint64 `json:"oldestDeliveryTime"`
	// MaxDeliveryCount is the largest delivery count among the pending entries
	MaxDeliveryCount uint64 `json:"maxDeliveryCount"`
	// Undelivered is the number of messages the group has not been delivered yet
	Undelivered int `json:"undelivered"`
	// Dangling is the number of distinct pending entries of the group whose message is deleted or missing
	Dangling int `json:"dangling"`
}

func (obj *StreamObject) danglingPending(group *StreamGroup) int {
	live := map[string]bool{}
	for _, entry := range obj.Entries {
		for _, msg := range entry.Msgs {
			if !msg.Deleted {
				live[streamIdKey(msg.Id)] = true
			}
		}
	}
	dangling := 0
	counted := map[string]bool{}
	for _, nack := range group.Pending {
		key := streamIdKey(nack.Id)
		if counted[key] {
			continue
		}
		counted[key] = true
		if !live[key] {
			dangling++
		}
	}
	return dangling
}

func (group *StreamGroup) backlogRows(undelivered, dangling int) []*ConsumerBacklog {
	owners := group.PendingOwners()
	rows := make(map[string]*ConsumerBacklog)
	ordered := make([]*ConsumerBacklog, 0, len(owners))
	counted := make(map[string]bool, len(owners))
	for _, nack := range group.Pending {
		key := streamIdKey(nack.Id)
		if counted[key] {
			continue
		}
		counted[key] = true
		names := owners[key]
		if len(names) == 0 {
			names = []string{""}
		}
		for _, name := range names {
			row := rows[name]
			if row == nil {
				row = &ConsumerBacklog{
					Group:              group.Name,
					Consumer:           name,
					OldestDeliveryTime: nack.DeliveryTime,
					Undelivered:        undelivered,
					Dangling:           dangling,
				}
				rows[name] = row
				ordered = append(ordered, row)
			}
			row.Pending++
			if nack.DeliveryTime < row.OldestDeliveryTime {
				row.OldestDeliveryTime = nack.DeliveryTime
			}
			if nack.DeliveryCount > row.MaxDeliveryCount {
				row.MaxDeliveryCount = nack.DeliveryCount
			}
		}
	}
	if len(ordered) == 0 {
		ordered = append(ordered, &ConsumerBacklog{
			Group:       group.Name,
			Undelivered: undelivered,
			Dangling:    dangling,
		})
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.Pending != b.Pending {
			return a.Pending > b.Pending
		}
		if a.OldestDeliveryTime != b.OldestDeliveryTime {
			return a.OldestDeliveryTime < b.OldestDeliveryTime
		}
		return a.Consumer < b.Consumer
	})
	return ordered
}

// ConsumerBacklog returns one row for every consumer holding pending entries of a group, plus one row with an
// empty Consumer for the pending entries nobody holds. A group without pending entries yields a single row.
// Rows are ordered by group name, then by descending Pending, ascending OldestDeliveryTime and Consumer.
func (obj *StreamObject) ConsumerBacklog() []*ConsumerBacklog {
	groups := make([]*StreamGroup, 0, len(obj.Groups))
	groups = append(groups, obj.Groups...)
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})
	result := make([]*ConsumerBacklog, 0)
	for _, group := range groups {
		result = append(result, group.backlogRows(len(obj.UndeliveredMessages(group)), obj.danglingPending(group))...)
	}
	return result
}
