package core

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hdt3213/rdb/model"
)

func TestWriteStreamObject(t *testing.T) {
	// Create a minimal test stream object
	stream := &model.StreamObject{
		BaseObject: &model.BaseObject{
			Key: "astream",
		},
		Version: 1, // Use version 1 for simplicity
		Length:  0, // Empty stream
		LastId: &model.StreamId{
			Ms:       0,
			Sequence: 0,
		},
		Entries: []*model.StreamEntry{}, // Empty entries
		Groups:  []*model.StreamGroup{}, // Empty groups
	}

	// Encode the stream object
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	err := encoder.WriteHeader()
	if err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	err = encoder.WriteDBHeader(0, 1, 0)
	if err != nil {
		t.Fatalf("Failed to write DB header: %v", err)
	}

	err = encoder.WriteStreamObject("astream", stream)
	if err != nil {
		t.Fatalf("Failed to write stream object: %v", err)
	}

	err = encoder.WriteEnd()
	if err != nil {
		t.Fatalf("Failed to write end: %v", err)
	}

	// Decode the stream object
	decodeStreamObject(t, &buf, stream)
}

func TestWriteStreamObjectWithEntries(t *testing.T) {
	stream := &model.StreamObject{
		BaseObject: &model.BaseObject{
			Key: "astream",
		},
		Version: 1,
		Length:  1,
		LastId: &model.StreamId{
			Ms:       1640995200000,
			Sequence: 0,
		},
		Entries: []*model.StreamEntry{
			{
				FirstMsgId: &model.StreamId{
					Ms:       1640995200000,
					Sequence: 0,
				},
				Fields: []string{"field1"},
				Msgs: []*model.StreamMessage{
					{
						Id: &model.StreamId{
							Ms:       1640995200000,
							Sequence: 0,
						},
						Fields: map[string]string{
							"field1": "value1",
						},
						Deleted: false,
					},
				},
			},
		},
		Groups: []*model.StreamGroup{},
	}

	// Test that encoding doesn't crash
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	err := encoder.WriteHeader()
	if err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	err = encoder.WriteDBHeader(0, 1, 0)
	if err != nil {
		t.Fatalf("Failed to write DB header: %v", err)
	}

	err = encoder.WriteStreamObject("astream", stream)
	if err != nil {
		t.Fatalf("Failed to write stream object: %v", err)
	}

	err = encoder.WriteEnd()
	if err != nil {
		t.Fatalf("Failed to write end: %v", err)
	}

	// Verify that we can at least write the data without errors
	if buf.Len() == 0 {
		t.Error("No data was written")
	}

	// Decode the stream object
	decodeStreamObject(t, &buf, stream)
}

func TestWriteStreamObjectVersion2(t *testing.T) {
	// Test with version 2 stream structure (similar to stream_listpacks_2.rdb)
	stream := &model.StreamObject{
		BaseObject: &model.BaseObject{
			Key: "astream",
		},
		Version: 2,
		Length:  2,
		LastId: &model.StreamId{
			Ms:       1681085312465,
			Sequence: 0,
		},
		FirstId: &model.StreamId{
			Ms:       1681085300799,
			Sequence: 0,
		},
		MaxDeletedId: &model.StreamId{
			Ms:       0,
			Sequence: 0,
		},
		AddedEntriesCount: 2,
		Entries: []*model.StreamEntry{
			{
				FirstMsgId: &model.StreamId{
					Ms:       1681085300799,
					Sequence: 0,
				},
				Fields: []string{"a", "b", "c"},
				Msgs: []*model.StreamMessage{
					{
						Id: &model.StreamId{
							Ms:       1681085300799,
							Sequence: 0,
						},
						Fields: map[string]string{
							"a": "1",
							"b": "2",
							"c": "3",
						},
						Deleted: false,
					},
					{
						Id: &model.StreamId{
							Ms:       1681085312465,
							Sequence: 0,
						},
						Fields: map[string]string{
							"a": "2",
							"b": "3",
							"c": "4",
						},
						Deleted: false,
					},
				},
			},
		},
		Groups: []*model.StreamGroup{},
	}

	// Encode the stream object
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	err := encoder.WriteHeader()
	if err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	err = encoder.WriteDBHeader(0, 1, 0)
	if err != nil {
		t.Fatalf("Failed to write DB header: %v", err)
	}

	err = encoder.WriteStreamObject("astream", stream)
	if err != nil {
		t.Fatalf("Failed to write stream object: %v", err)
	}

	err = encoder.WriteEnd()
	if err != nil {
		t.Fatalf("Failed to write end: %v", err)
	}

	// Verify that we can at least write the data without errors
	if buf.Len() == 0 {
		t.Error("No data was written")
	}

	// Decode the stream object
	decodeStreamObject(t, &buf, stream)
}

func TestWriteStreamObjectVersion3(t *testing.T) {
	// Test with version 3 stream structure
	stream := &model.StreamObject{
		BaseObject: &model.BaseObject{
			Key: "astream",
		},
		Version: 3,
		Length:  2,
		LastId: &model.StreamId{
			Ms:       0,
			Sequence: 0,
		},
		FirstId: &model.StreamId{
			Ms:       0,
			Sequence: 0,
		},
		MaxDeletedId: &model.StreamId{
			Ms:       0,
			Sequence: 0,
		},
		AddedEntriesCount: 0,
		Entries: []*model.StreamEntry{
			{
				FirstMsgId: &model.StreamId{
					Ms:       1681085300799,
					Sequence: 0,
				},
				Fields: []string{"a", "b", "c"},
				Msgs: []*model.StreamMessage{
					{
						Id: &model.StreamId{
							Ms:       1681085300799,
							Sequence: 0,
						},
						Fields: map[string]string{
							"a": "1",
							"b": "2",
							"c": "3",
						},
						Deleted: false,
					},
					{
						Id: &model.StreamId{
							Ms:       1681085312465,
							Sequence: 0,
						},
						Fields: map[string]string{
							"a": "2",
							"b": "3",
							"c": "4",
						},
						Deleted: false,
					},
				},
			},
		},
		Groups: []*model.StreamGroup{
			{
				Name: "test-group",
				LastId: &model.StreamId{
					Ms:       0,
					Sequence: 0,
				},
				EntriesRead: 0,
				Pending:     []*model.StreamNAck{},
				Consumers: []*model.StreamConsumer{
					{
						Name:       "test-consumer",
						SeenTime:   1640995200000,
						ActiveTime: 1640995200000,
						Pending:    []*model.StreamId{},
					},
				},
			},
		},
	}

	// Encode the stream object
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	err := encoder.WriteHeader()
	if err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	err = encoder.WriteDBHeader(0, 1, 0)
	if err != nil {
		t.Fatalf("Failed to write DB header: %v", err)
	}

	err = encoder.WriteStreamObject("astream", stream)
	if err != nil {
		t.Fatalf("Failed to write stream object: %v", err)
	}

	err = encoder.WriteEnd()
	if err != nil {
		t.Fatalf("Failed to write end: %v", err)
	}

	// Verify that we can at least write the data without errors
	if buf.Len() == 0 {
		t.Error("No data was written")
	}

	// Decode the stream object
	decodeStreamObject(t, &buf, stream)
}

func TestWriteStreamObjectVersion4(t *testing.T) {
	stream := &model.StreamObject{
		BaseObject: &model.BaseObject{
			Key: "astream",
		},
		Version: 4,
		Length:  1,
		LastId: &model.StreamId{
			Ms:       1704557973866,
			Sequence: 0,
		},
		FirstId: &model.StreamId{
			Ms:       1704557973866,
			Sequence: 0,
		},
		MaxDeletedId: &model.StreamId{
			Ms:       0,
			Sequence: 0,
		},
		AddedEntriesCount: 1,
		Entries: []*model.StreamEntry{
			{
				FirstMsgId: &model.StreamId{
					Ms:       1704557973866,
					Sequence: 0,
				},
				Fields: []string{"name"},
				Msgs: []*model.StreamMessage{
					{
						Id: &model.StreamId{
							Ms:       1704557973866,
							Sequence: 0,
						},
						Fields:  map[string]string{"name": "Sara"},
						Deleted: false,
					},
				},
			},
		},
		Groups: []*model.StreamGroup{},
		IdmpDuration:   60000,
		IdmpMaxEntries: 100,
		IdmpProducers: []*model.StreamProducer{
			{
				Id: "producer-1",
				Entries: []*model.StreamIdmpEntry{
					{
						Iid: "req-abc-123",
						StreamId: &model.StreamId{
							Ms:       1704557973866,
							Sequence: 0,
						},
					},
				},
			},
		},
		IidsAdded:      5,
		IidsDuplicates: 1,
	}

	var buf bytes.Buffer
	encoder := NewEncoder(&buf)
	if err := encoder.WriteHeader(); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}
	if err := encoder.WriteDBHeader(0, 1, 0); err != nil {
		t.Fatalf("Failed to write DB header: %v", err)
	}
	if err := encoder.WriteStreamObject("astream", stream); err != nil {
		t.Fatalf("Failed to write stream object: %v", err)
	}
	if err := encoder.WriteEnd(); err != nil {
		t.Fatalf("Failed to write end: %v", err)
	}

	// Decode and verify
	decoder := NewDecoder(&buf)
	var decoded *model.StreamObject
	err := decoder.Parse(func(obj model.RedisObject) bool {
		if s, ok := obj.(*model.StreamObject); ok {
			decoded = s
			return false
		}
		return true
	})
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	if decoded == nil {
		t.Fatal("Failed to decode stream object")
	}
	if decoded.Version != 4 {
		t.Errorf("Version: expected 4, got %d", decoded.Version)
	}
	if decoded.IdmpDuration != 60000 {
		t.Errorf("IdmpDuration: expected 60000, got %d", decoded.IdmpDuration)
	}
	if decoded.IdmpMaxEntries != 100 {
		t.Errorf("IdmpMaxEntries: expected 100, got %d", decoded.IdmpMaxEntries)
	}
	if len(decoded.IdmpProducers) != 1 {
		t.Fatalf("IdmpProducers count: expected 1, got %d", len(decoded.IdmpProducers))
	}
	p := decoded.IdmpProducers[0]
	if p.Id != "producer-1" {
		t.Errorf("Producer Id: expected producer-1, got %s", p.Id)
	}
	if len(p.Entries) != 1 {
		t.Fatalf("Producer entries count: expected 1, got %d", len(p.Entries))
	}
	if p.Entries[0].Iid != "req-abc-123" {
		t.Errorf("IDMP Iid: expected req-abc-123, got %s", p.Entries[0].Iid)
	}
	if p.Entries[0].StreamId.Ms != 1704557973866 {
		t.Errorf("IDMP StreamId.Ms: expected 1704557973866, got %d", p.Entries[0].StreamId.Ms)
	}
	if decoded.IidsAdded != 5 {
		t.Errorf("IidsAdded: expected 5, got %d", decoded.IidsAdded)
	}
	if decoded.IidsDuplicates != 1 {
		t.Errorf("IidsDuplicates: expected 1, got %d", decoded.IidsDuplicates)
	}
}

func decodeStreamObject(t *testing.T, buf *bytes.Buffer, stream *model.StreamObject) {
	// Decode the stream object
	decoder := NewDecoder(buf)

	// Read objects until we find our stream
	var decodedStream *model.StreamObject
	var err = decoder.Parse(func(obj model.RedisObject) bool {
		if streamObj, ok := obj.(*model.StreamObject); ok && streamObj.GetKey() == "astream" {
			decodedStream = streamObj
			return false // stop parsing
		}
		return true // continue parsing
	})
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if decodedStream == nil {
		t.Fatal("Failed to decode stream object")
	}

	// Verify the decoded stream matches the original
	if decodedStream.Version != stream.Version {
		t.Errorf("Version mismatch: expected %d, got %d", stream.Version, decodedStream.Version)
	}

	if decodedStream.Length != stream.Length {
		t.Errorf("Length mismatch: expected %d, got %d", stream.Length, decodedStream.Length)
	}

	if decodedStream.LastId.Ms != stream.LastId.Ms || decodedStream.LastId.Sequence != stream.LastId.Sequence {
		t.Errorf("LastId mismatch: expected %d-%d, got %d-%d",
			stream.LastId.Ms, stream.LastId.Sequence,
			decodedStream.LastId.Ms, decodedStream.LastId.Sequence)
	}

	if len(decodedStream.Entries) != len(stream.Entries) {
		t.Errorf("Entries count mismatch: expected %d, got %d", len(stream.Entries), len(decodedStream.Entries))
	}

	if len(decodedStream.Groups) != len(stream.Groups) {
		t.Errorf("Groups count mismatch: expected %d, got %d", len(stream.Groups), len(decodedStream.Groups))
	}

	for i, entry := range decodedStream.Entries {
		if !reflect.DeepEqual(entry.Fields, stream.Entries[i].Fields) {
			t.Errorf("Fields mismatch at index %d: expected %v, got %v", i, stream.Entries[i].Fields, entry.Fields)
		}
		if !reflect.DeepEqual(entry.Msgs, stream.Entries[i].Msgs) {
			t.Errorf("Msgs mismatch at index %d: expected %v, got %v", i, stream.Entries[i].Msgs, entry.Msgs)
		}
	}
}
