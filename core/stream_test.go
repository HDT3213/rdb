package core

import (
	"bytes"
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
	decoder := NewDecoder(&buf)

	// Read objects until we find our stream
	var decodedStream *model.StreamObject
	err = decoder.Parse(func(obj model.RedisObject) bool {
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
}

func TestWriteStreamObjectWithEntries(t *testing.T) {
	// Create a test stream object with one entry
	stream := &model.StreamObject{
		BaseObject: &model.BaseObject{
			Key: "test-stream",
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

	err = encoder.WriteStreamObject("test-stream", stream)
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
}
