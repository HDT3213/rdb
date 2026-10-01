package model

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func mkId(ms, seq uint64) *StreamId {
	return &StreamId{Ms: ms, Sequence: seq}
}

func mkMsg(ms, seq uint64, deleted bool) *StreamMessage {
	return &StreamMessage{Id: mkId(ms, seq), Fields: map[string]string{"k": "v"}, Deleted: deleted}
}

func mkEntry(msgs ...*StreamMessage) *StreamEntry {
	return &StreamEntry{FirstMsgId: msgs[0].Id, Fields: []string{"k"}, Msgs: msgs}
}

func mkNack(ms, seq, deliveryTime, deliveryCount uint64) *StreamNAck {
	return &StreamNAck{Id: mkId(ms, seq), DeliveryTime: deliveryTime, DeliveryCount: deliveryCount}
}

func mkConsumer(name string, pending ...*StreamId) *StreamConsumer {
	return &StreamConsumer{Name: name, Pending: pending}
}

func messageIds(msgs []*StreamMessage) []string {
	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		text, _ := m.Id.MarshalText()
		ids = append(ids, string(text))
	}
	return ids
}

func backlogRows(rows []*ConsumerBacklog) []string {
	texts := make([]string, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, strings.Join([]string{
			row.Group, row.Consumer,
			strconv.Itoa(row.Pending),
			strconv.FormatUint(row.OldestDeliveryTime, 10),
			strconv.FormatUint(row.MaxDeliveryCount, 10),
			strconv.Itoa(row.Undelivered),
			strconv.Itoa(row.Dangling),
		}, "/"))
	}
	return texts
}

func TestPendingOwners(t *testing.T) {
	group := &StreamGroup{
		Name: "g1",
		Pending: []*StreamNAck{
			mkNack(1, 0, 100, 1), mkNack(2, 0, 200, 1), mkNack(3, 0, 300, 1), mkNack(7, 2, 100, 1), mkNack(7, 2, 300, 4),
		},
		Consumers: []*StreamConsumer{
			mkConsumer("carol", mkId(1, 0), mkId(9, 9)),
			mkConsumer("alice", mkId(1, 0), mkId(1, 0), mkId(2, 0)),
		},
	}
	want := map[string][]string{
		"1-0": {"alice", "carol"},
		"2-0": {"alice"},
		"3-0": {},
		"7-2": {},
	}
	if got := group.PendingOwners(); !reflect.DeepEqual(got, want) {
		t.Errorf("PendingOwners() = %v, want %v", got, want)
	}
}

func TestUndeliveredMessages(t *testing.T) {
	obj := &StreamObject{Entries: []*StreamEntry{
		mkEntry(mkMsg(9, 0, false)),
		mkEntry(mkMsg(1, 0, false), mkMsg(4, 0, false), mkMsg(7, 1, false), mkMsg(7, 2, true), mkMsg(8, 5, false), mkMsg(8, 1, false)),
	}}
	tests := []struct {
		name  string
		group *StreamGroup
		want  string
	}{
		{"compares every message id, orders by id, skips deleted", &StreamGroup{LastId: mkId(4, 0)}, "7-1,8-1,8-5,9-0"},
		{"compares the sequence within a millisecond", &StreamGroup{LastId: mkId(7, 1)}, "8-1,8-5,9-0"},
		{"no last id returns every live message", &StreamGroup{}, "1-0,4-0,7-1,8-1,8-5,9-0"},
		{
			"skips messages the group still holds as pending",
			&StreamGroup{LastId: mkId(1, 0), Pending: []*StreamNAck{mkNack(4, 0, 1, 1), mkNack(8, 5, 1, 1)}},
			"7-1,8-1,9-0",
		},
		{
			"a pending entry above the last id is not undelivered",
			&StreamGroup{LastId: mkId(1, 0), Pending: []*StreamNAck{mkNack(9, 0, 1, 1)}},
			"4-0,7-1,8-1,8-5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strings.Join(messageIds(obj.UndeliveredMessages(tt.group)), ","); got != tt.want {
				t.Errorf("UndeliveredMessages() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestConsumerBacklog(t *testing.T) {
	entries := []*StreamEntry{
		mkEntry(mkMsg(1, 0, false), mkMsg(2, 0, false), mkMsg(3, 0, true), mkMsg(4, 0, false)),
	}
	tests := []struct {
		name    string
		entries []*StreamEntry
		groups  []*StreamGroup
		want    string
	}{
		{
			name:    "aggregates the entries of each consumer",
			entries: entries,
			groups: []*StreamGroup{{
				Name:    "g1",
				LastId:  mkId(2, 0),
				Pending: []*StreamNAck{mkNack(1, 0, 500, 2), mkNack(2, 0, 300, 7), mkNack(4, 0, 900, 1)},
				Consumers: []*StreamConsumer{
					mkConsumer("alice", mkId(1, 0), mkId(2, 0)),
					mkConsumer("bob", mkId(4, 0)),
				},
			}},
			want: "g1/alice/2/300/7/0/0,g1/bob/1/900/1/0/0",
		},
		{
			name:    "an entry held by two consumers counts for both",
			entries: entries,
			groups: []*StreamGroup{{
				Name:    "g1",
				LastId:  mkId(4, 0),
				Pending: []*StreamNAck{mkNack(1, 0, 400, 5)},
				Consumers: []*StreamConsumer{
					mkConsumer("bob", mkId(1, 0)),
					mkConsumer("alice", mkId(1, 0), mkId(1, 0)),
				},
			}},
			want: "g1/alice/1/400/5/0/0,g1/bob/1/400/5/0/0",
		},
		{
			name:    "a repeated pending id counts once with its first record",
			entries: entries,
			groups: []*StreamGroup{{
				Name:      "g1",
				LastId:    mkId(4, 0),
				Pending:   []*StreamNAck{mkNack(1, 0, 400, 2), mkNack(1, 0, 100, 9)},
				Consumers: []*StreamConsumer{mkConsumer("alice", mkId(1, 0))},
			}},
			want: "g1/alice/1/400/2/0/0",
		},
		{
			name:    "entries nobody holds share the row of an empty consumer name",
			entries: entries,
			groups: []*StreamGroup{{
				Name:      "g1",
				LastId:    mkId(4, 0),
				Pending:   []*StreamNAck{mkNack(1, 0, 500, 2), mkNack(2, 0, 0, 4), mkNack(4, 0, 800, 9), mkNack(5, 0, 50, 1)},
				Consumers: []*StreamConsumer{mkConsumer("alice", mkId(1, 0), mkId(7, 7)), mkConsumer("", mkId(5, 0))},
			}},
			want: "g1//3/0/9/0/1,g1/alice/1/500/2/0/1",
		},
		{
			name:    "a group without pending entries still has a row",
			entries: entries,
			groups: []*StreamGroup{
				{Name: "zulu", LastId: mkId(4, 0)},
				{Name: "alpha", LastId: mkId(1, 0), Consumers: []*StreamConsumer{mkConsumer("alice")}},
			},
			want: "alpha//0/0/0/2/0,zulu//0/0/0/0/0",
		},
		{
			name:    "rows are ordered by pending, oldest delivery time, then name",
			entries: []*StreamEntry{mkEntry(mkMsg(1, 0, false), mkMsg(2, 0, false), mkMsg(3, 0, false), mkMsg(4, 0, false), mkMsg(5, 0, false), mkMsg(6, 0, false))},
			groups: []*StreamGroup{{
				Name:   "g1",
				LastId: mkId(6, 0),
				Pending: []*StreamNAck{
					mkNack(1, 0, 700, 1), mkNack(2, 0, 200, 1), mkNack(3, 0, 200, 1), mkNack(4, 0, 900, 1), mkNack(5, 0, 100, 1), mkNack(6, 0, 100, 1),
				},
				Consumers: []*StreamConsumer{
					mkConsumer("dave", mkId(4, 0)),
					mkConsumer("carol", mkId(3, 0)),
					mkConsumer("bob", mkId(2, 0)),
					mkConsumer("alice", mkId(1, 0)),
					mkConsumer("erin", mkId(5, 0), mkId(6, 0)),
				},
			}},
			want: "g1/erin/2/100/1/0/0,g1/bob/1/200/1/0/0,g1/carol/1/200/1/0/0,g1/alice/1/700/1/0/0,g1/dave/1/900/1/0/0",
		},
		{
			name:    "undelivered and dangling are per group",
			entries: entries,
			groups: []*StreamGroup{
				{
					Name:      "behind",
					LastId:    mkId(1, 0),
					Pending:   []*StreamNAck{mkNack(2, 0, 10, 1), mkNack(3, 0, 20, 1), mkNack(9, 0, 30, 1)},
					Consumers: []*StreamConsumer{mkConsumer("alice", mkId(2, 0), mkId(3, 0), mkId(9, 0))},
				},
				{Name: "current", LastId: mkId(4, 0)},
			},
			want: "behind/alice/3/10/1/1/2,current//0/0/0/0/0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &StreamObject{Entries: tt.entries, Groups: tt.groups}
			if got := strings.Join(backlogRows(obj.ConsumerBacklog()), ","); got != tt.want {
				t.Errorf("ConsumerBacklog() = %s, want %s", got, tt.want)
			}
		})
	}
}
