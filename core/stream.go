package core

import (
	"encoding/binary"
	"fmt"

	"github.com/hdt3213/rdb/model"
)

const (
	// StreamItemFlagNone means No special flags.
	StreamItemFlagNone = 0
	// StreamItemFlagDeleted means entry was deleted
	StreamItemFlagDeleted = 1 << 0
	// StreamItemFlagSameFields means entry has the same fields as the master entry
	StreamItemFlagSameFields = 1 << 1
)

func (dec *Decoder) readStreamListPacks(isV2 bool) (*model.StreamObject, error) {
	entries, err := dec.readStreamEntries()
	if err != nil {
		return nil, err
	}
	steamLen, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	lastId, err := dec.readStreamId()
	if err != nil {
		return nil, err
	}
	stream := &model.StreamObject{
		Entries: entries,
		Length:  steamLen,
		LastId:  lastId,
	}

	if isV2 {
		firstId, err := dec.readStreamId()
		if err != nil {
			return nil, err
		}
		stream.FirstId = firstId
		maxDeletedId, err := dec.readStreamId()
		if err != nil {
			return nil, err
		}
		stream.MaxDeletedId = maxDeletedId
		addedCount, _, err := dec.readLength()
		if err != nil {
			return nil, err
		}
		stream.AddedEntriesCount = addedCount
	}
	groups, err := dec.readStreamGroups(isV2)
	if err != nil {
		return nil, err
	}
	stream.Groups = groups
	stream.IsV2 = isV2
	return stream, nil
}

func (dec *Decoder) readStreamId() (*model.StreamId, error) {
	ms, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	seq, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	return &model.StreamId{
		Ms:       ms,
		Sequence: seq,
	}, nil
}

// readStreamEntries read entries
func (dec *Decoder) readStreamEntries() ([]*model.StreamEntry, error) {
	length, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	var result []*model.StreamEntry
	for i := uint64(0); i < length; i++ {
		header, err := dec.readString()
		if err != nil {
			return nil, err
		}
		cursor := 0
		msBin, err := readBytes(header, &cursor, 8)
		if err != nil {
			return nil, err
		}
		ms := binary.BigEndian.Uint64(msBin)
		seqBin, err := readBytes(header, &cursor, 8)
		if err != nil {
			return nil, err
		}
		seq := binary.BigEndian.Uint64(seqBin)
		firstId := &model.StreamId{
			Ms:       ms,
			Sequence: seq,
		}

		buf, err := dec.readString()
		if err != nil {
			return nil, err
		}
		cursor = 0
		// skip 4Byte total-bytes + 2Byte num-elements
		_, _ = readBytes(buf, &cursor, 6)
		entry, err := dec.readStreamEntryContent(buf, &cursor, firstId)
		if err != nil {
			return nil, err
		}
		entry.FirstMsgId = firstId
		result = append(result, entry)
	}
	return result, nil
}

// readStreamEntryContent read messages in a stream entry
func (dec *Decoder) readStreamEntryContent(buf []byte, cursor *int, firstId *model.StreamId) (*model.StreamEntry, error) {
	// read count
	count, err := dec.readListPackEntryAsInt(buf, cursor)
	if err != nil {
		return nil, fmt.Errorf("read stream entry count failed: %v", err)
	}
	deleted, err := dec.readListPackEntryAsInt(buf, cursor)
	if err != nil {
		return nil, fmt.Errorf("read stream entry deleted count failed: %v", err)
	}

	// read field names of master entry
	fieldNum0, err := dec.readListPackEntryAsInt(buf, cursor)
	if err != nil {
		return nil, fmt.Errorf("read stream field number failed: %v", err)
	}
	masterFieldNum := int(fieldNum0)
	masterFieldNames := make([]string, masterFieldNum)
	for i := 0; i < masterFieldNum; i++ {
		name, _, err := dec.readListPackEntry(buf, cursor)
		if err != nil {
			return nil, fmt.Errorf("read field name of stream entry failed: %v", err)
		}
		masterFieldNames[i] = string(name)
	}
	// read end flag
	if _, _, err = dec.readListPackEntry(buf, cursor); err != nil {
		return nil, fmt.Errorf("read fields end flag failed: %v", err)
	}

	total := count + deleted
	msgs := make([]*model.StreamMessage, 0, total)
	for i := int64(0); i < total; i++ {
		flag, err := dec.readListPackEntryAsInt(buf, cursor)
		if err != nil {
			return nil, fmt.Errorf("read stream item flag failed: %v", err)
		}
		ms, err := dec.readListPackEntryAsInt(buf, cursor)
		if err != nil {
			return nil, fmt.Errorf("read stream item id ms failed: %v", err)
		}
		seq, err := dec.readListPackEntryAsInt(buf, cursor)
		if err != nil {
			return nil, fmt.Errorf("read stream item id seq failed: %v", err)
		}
		// ms and seq may be negative
		msgId := &model.StreamId{
			Ms:       uint64(ms + int64(firstId.Ms)),
			Sequence: uint64(seq + int64(firstId.Sequence)),
		}
		fieldNum := masterFieldNum
		if flag&StreamItemFlagSameFields == 0 {
			fieldNum0, err := dec.readListPackEntryAsInt(buf, cursor)
			if err != nil {
				return nil, fmt.Errorf("read stream item field number failed: %v", err)
			}
			fieldNum = int(fieldNum0)
		}
		msg := &model.StreamMessage{
			Id:      msgId,
			Fields:  make(map[string]string, masterFieldNum),
			Deleted: flag&StreamItemFlagDeleted > 0,
		}

		for i := 0; i < fieldNum; i++ {
			var fieldName string
			if flag&StreamItemFlagSameFields > 0 {
				fieldName = masterFieldNames[i]
			} else {
				fieldNameBin, _, err := dec.readListPackEntry(buf, cursor)
				if err != nil {
					return nil, fmt.Errorf("read stream item field name failed: %v", err)
				}
				fieldName = unsafeBytes2Str(fieldNameBin)
			}
			fieldValue, _, err := dec.readListPackEntry(buf, cursor)
			if err != nil {
				return nil, fmt.Errorf("read stream item field value failed: %v", err)
			}
			msg.Fields[fieldName] = unsafeBytes2Str(fieldValue)
		}
		// read end flag
		if _, _, err = dec.readListPackEntry(buf, cursor); err != nil {
			return nil, fmt.Errorf("read fields end flag failed: %v", err)
		}
		msgs = append(msgs, msg)
	}
	return &model.StreamEntry{
		Fields: masterFieldNames,
		Msgs:   msgs,
	}, nil
}

func (dec *Decoder) readStreamGroups(isV2 bool) ([]*model.StreamGroup, error) {
	groupCount, _, err := dec.readLength()
	if err != nil {
		return nil, err
	}
	groups := make([]*model.StreamGroup, 0, int(groupCount))
	for i := uint64(0); i < groupCount; i++ {
		name, _ := dec.readString()
		if err != nil {
			return nil, err
		}
		lastId, err := dec.readStreamId()
		if err != nil {
			return nil, err
		}

		var entriesRead uint64
		if isV2 {
			entriesRead, _, err = dec.readLength()
			if err != nil {
				return nil, err
			}
		}

		// read pending list
		pendingCount, _, err := dec.readLength()
		if err != nil {
			return nil, err
		}
		pending := make([]*model.StreamNAck, 0, int(pendingCount))
		for j := uint64(0); j < pendingCount; j++ {
			if err := dec.readFull(dec.buffer); err != nil {
				return nil, err
			}
			ms := binary.BigEndian.Uint64(dec.buffer)
			if err := dec.readFull(dec.buffer); err != nil {
				return nil, err
			}
			seq := binary.BigEndian.Uint64(dec.buffer)
			streamId := &model.StreamId{
				Ms:       ms,
				Sequence: seq,
			}
			if err := dec.readFull(dec.buffer); err != nil {
				return nil, err
			}
			deliveryTime := binary.LittleEndian.Uint64(dec.buffer)
			deliveryCount, _, err := dec.readLength()
			if err != nil {
				return nil, err
			}
			pending = append(pending, &model.StreamNAck{
				Id:            streamId,
				DeliveryTime:  deliveryTime,
				DeliveryCount: deliveryCount,
			})
		}

		// read consumers
		consumerCount, _, err := dec.readLength()
		if err != nil {
			return nil, err
		}
		consumers := make([]*model.StreamConsumer, 0, int(consumerCount))
		for j := uint64(0); j < consumerCount; j++ {
			consumerName, err := dec.readString()
			if err != nil {
				return nil, err
			}
			if err := dec.readFull(dec.buffer); err != nil {
				return nil, err
			}
			seenTime := binary.LittleEndian.Uint64(dec.buffer)
			consumerPendingCount, _, err := dec.readLength()
			if err != nil {
				return nil, err
			}
			consumerPending := make([]*model.StreamId, 0, int(consumerPendingCount))
			for k := uint64(0); k < consumerPendingCount; k++ {
				if err := dec.readFull(dec.buffer); err != nil {
					return nil, err
				}
				ms := binary.BigEndian.Uint64(dec.buffer)
				if err := dec.readFull(dec.buffer); err != nil {
					return nil, err
				}
				seq := binary.BigEndian.Uint64(dec.buffer)
				consumerPending = append(consumerPending, &model.StreamId{
					Ms:       ms,
					Sequence: seq,
				})
			}
			consumers = append(consumers, &model.StreamConsumer{
				SeenTime: seenTime,
				Name:     unsafeBytes2Str(consumerName),
				Pending:  consumerPending,
			})
		}
		groups = append(groups, &model.StreamGroup{
			Name:        unsafeBytes2Str(name),
			LastId:      lastId,
			Pending:     pending,
			Consumers:   consumers,
			EntriesRead: entriesRead,
		})
	}
	return groups, nil
}
