package core

import (
	"encoding/binary"
	"fmt"
	"strconv"

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

func (dec *Decoder) readStreamListPacks(version uint) (*model.StreamObject, error) {
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

	if version >= 2 {
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
	groups, err := dec.readStreamGroups(version)
	if err != nil {
		return nil, err
	}
	stream.Groups = groups
	stream.Version = version
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
		name, err := dec.readListPackEntryAsString(buf, cursor)
		if err != nil {
			return nil, fmt.Errorf("read field name of stream entry failed: %v", err)
		}
		masterFieldNames[i] = string(name)
	}
	// read lp count of master entry
	if _, err = dec.readListPackEntryAsString(buf, cursor); err != nil {
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
				fieldNameBin, err := dec.readListPackEntryAsString(buf, cursor)
				if err != nil {
					return nil, fmt.Errorf("read stream item field name failed: %v", err)
				}
				fieldName = unsafeBytes2Str(fieldNameBin)
			}
			fieldValue, err := dec.readListPackEntryAsString(buf, cursor)
			if err != nil {
				return nil, fmt.Errorf("read stream item field value failed: %v", err)
			}
			msg.Fields[fieldName] = unsafeBytes2Str(fieldValue)
		}
		// read lp count
		if _, err = dec.readListPackEntryAsString(buf, cursor); err != nil {
			return nil, fmt.Errorf("read fields end flag failed: %v", err)
		}
		msgs = append(msgs, msg)
	}
	return &model.StreamEntry{
		Fields: masterFieldNames,
		Msgs:   msgs,
	}, nil
}

func (dec *Decoder) readStreamGroups(version uint) ([]*model.StreamGroup, error) {
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
		if version >= 2 {
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
			activeTime := seenTime
			if version >= 3 {
				if err := dec.readFull(dec.buffer); err != nil {
					return nil, err
				}
				activeTime = binary.LittleEndian.Uint64(dec.buffer)
			}
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
				SeenTime:   seenTime,
				Name:       unsafeBytes2Str(consumerName),
				Pending:    consumerPending,
				ActiveTime: activeTime,
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

// WriteStreamObject writes a stream object to RDB file
func (enc *Encoder) WriteStreamObject(key string, stream *model.StreamObject, options ...interface{}) error {
	err := enc.beforeWriteObject(options...)
	if err != nil {
		return err
	}

	// Write stream type based on version
	var streamType byte
	switch stream.Version {
	case 1:
		streamType = typeStreamListPacks
	case 2:
		streamType = typeStreamListPacks2
	case 3:
		streamType = typeStreamListPacks3
	default:
		streamType = typeStreamListPacks // default to version 1
	}

	err = enc.write([]byte{streamType})
	if err != nil {
		return err
	}

	err = enc.writeString(key)
	if err != nil {
		return err
	}

	// Write stream entries
	err = enc.writeStreamEntries(stream.Entries)
	if err != nil {
		return err
	}

	// Write stream length
	err = enc.writeLength(stream.Length)
	if err != nil {
		return err
	}

	// Write last ID
	err = enc.writeStreamId(stream.LastId)
	if err != nil {
		return err
	}

	// Write version 2+ fields if available
	if stream.Version >= 2 {
		if stream.FirstId != nil {
			err = enc.writeStreamId(stream.FirstId)
			if err != nil {
				return err
			}
		} else {
			// Write zero ID if FirstId is nil
			err = enc.writeStreamId(&model.StreamId{Ms: 0, Sequence: 0})
			if err != nil {
				return err
			}
		}

		if stream.MaxDeletedId != nil {
			err = enc.writeStreamId(stream.MaxDeletedId)
			if err != nil {
				return err
			}
		} else {
			// Write zero ID if MaxDeletedId is nil
			err = enc.writeStreamId(&model.StreamId{Ms: 0, Sequence: 0})
			if err != nil {
				return err
			}
		}

		err = enc.writeLength(stream.AddedEntriesCount)
		if err != nil {
			return err
		}
	}

	// Write stream groups
	err = enc.writeStreamGroups(stream.Groups, stream.Version)
	if err != nil {
		return err
	}

	enc.state = writtenObjectState
	return nil
}

// writeStreamId writes a stream ID
func (enc *Encoder) writeStreamId(id *model.StreamId) error {
	err := enc.writeLength(id.Ms)
	if err != nil {
		return err
	}
	return enc.writeLength(id.Sequence)
}

// writeStreamEntries writes stream entries
func (enc *Encoder) writeStreamEntries(entries []*model.StreamEntry) error {
	err := enc.writeLength(uint64(len(entries)))
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Write entry header (first message ID)
		header := make([]byte, 16)
		binary.BigEndian.PutUint64(header[0:8], entry.FirstMsgId.Ms)
		binary.BigEndian.PutUint64(header[8:16], entry.FirstMsgId.Sequence)
		err = enc.writeString(unsafeBytes2Str(header))
		if err != nil {
			return err
		}

		// Write entry content as listpack
		err = enc.writeStreamEntryContent(entry)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeStreamEntryContent writes a stream entry content as listpack
func (enc *Encoder) writeStreamEntryContent(entry *model.StreamEntry) error {
	// Calculate total messages (including deleted ones)
	totalMsgs := len(entry.Msgs)
	deletedCount := 0
	for _, msg := range entry.Msgs {
		if msg.Deleted {
			deletedCount++
		}
	}
	validCount := totalMsgs - deletedCount

	// Build listpack with proper backlen values
	var entries []listpackEntry

	// Add count and deleted count
	entries = append(entries, listpackEntry{intVal: int64(validCount)})
	entries = append(entries, listpackEntry{intVal: int64(deletedCount)})

	// Add master field names
	entries = append(entries, listpackEntry{intVal: int64(len(entry.Fields))})
	for _, field := range entry.Fields {
		entries = append(entries, listpackEntry{strVal: field})
	}
	// Add field count for master entry (this is what the decoder reads as "end flag")
	entries = append(entries, listpackEntry{strVal: strconv.Itoa(len(entry.Fields))})

	// Add messages
	for _, msg := range entry.Msgs {
		// Calculate flag
		flag := StreamItemFlagNone
		if msg.Deleted {
			flag |= StreamItemFlagDeleted
		}

		// Check if message uses same fields as master
		if len(msg.Fields) == len(entry.Fields) {
			sameFields := true
			for _, field := range entry.Fields {
				if _, exists := msg.Fields[field]; !exists {
					sameFields = false
					break
				}
			}
			if sameFields {
				flag |= StreamItemFlagSameFields
			}
		}

		// Add flag
		entries = append(entries, listpackEntry{intVal: int64(flag)})

		// Add message ID (relative to first message ID)
		msDiff := int64(msg.Id.Ms) - int64(entry.FirstMsgId.Ms)
		seqDiff := int64(msg.Id.Sequence) - int64(entry.FirstMsgId.Sequence)
		entries = append(entries, listpackEntry{intVal: msDiff})
		entries = append(entries, listpackEntry{intVal: seqDiff})

		// Add field count if not same fields
		if flag&StreamItemFlagSameFields == 0 {
			entries = append(entries, listpackEntry{intVal: int64(len(msg.Fields))})
		}

		// Add fields
		if flag&StreamItemFlagSameFields > 0 {
			// Use master field names order
			for _, field := range entry.Fields {
				value := msg.Fields[field]
				entries = append(entries, listpackEntry{strVal: value})
			}
		} else {
			// Add field names and values
			for fieldName, fieldValue := range msg.Fields {
				entries = append(entries, listpackEntry{strVal: fieldName})
				entries = append(entries, listpackEntry{strVal: fieldValue})
			}
		}

		// Add field count for this message (this is what the decoder reads as "end flag")
		entries = append(entries, listpackEntry{strVal: strconv.Itoa(len(msg.Fields))})
	}

	// Build listpack with proper backlen values
	listpackData := enc.buildListpackWithBacklen(entries)

	// Write the complete listpack
	err := enc.writeString(unsafeBytes2Str(listpackData))
	if err != nil {
		return err
	}

	return nil
}

type listpackEntry struct {
	intVal int64
	strVal string
}

// buildListpackWithBacklen builds a proper listpack with backlen values
func (enc *Encoder) buildListpackWithBacklen(entries []listpackEntry) []byte {
	var listpackData []byte
	var entrySizes []uint32

	// First pass: encode entries and calculate sizes
	for _, entry := range entries {
		var encoded []byte
		if entry.strVal != "" {
			encoded = enc.encodeListPackString(entry.strVal)
		} else {
			encoded = enc.encodeListPackInt(entry.intVal)
		}
		listpackData = append(listpackData, encoded...)
		entrySizes = append(entrySizes, uint32(len(encoded)))
	}

	// Second pass: add backlen values
	var finalListpack []byte
	for i := len(entries) - 1; i >= 0; i-- {
		// Add backlen
		backlen := enc.encodeBacklen(entrySizes[i])
		finalListpack = append(backlen, finalListpack...)
		// Add entry
		entryStart := 0
		for j := 0; j < i; j++ {
			entryStart += int(entrySizes[j])
		}
		entryEnd := entryStart + int(entrySizes[i])
		finalListpack = append(listpackData[entryStart:entryEnd], finalListpack...)
	}

	// Add header
	totalBytes := len(finalListpack) + 6 // 6 bytes for header
	header := make([]byte, 6)
	binary.LittleEndian.PutUint32(header[0:4], uint32(totalBytes))
	binary.LittleEndian.PutUint16(header[4:6], uint16(len(entries)))

	return append(header, finalListpack...)
}

// encodeBacklen encodes a backlen value
func (enc *Encoder) encodeBacklen(elementLen uint32) []byte {
	if elementLen <= 127 {
		return []byte{byte(elementLen)}
	} else if elementLen < (1<<14)-1 {
		return []byte{
			byte(0x80 | (elementLen >> 8)),
			byte(elementLen & 0xFF),
		}
	} else if elementLen < (1<<21)-1 {
		return []byte{
			byte(0xC0 | (elementLen >> 16)),
			byte((elementLen >> 8) & 0xFF),
			byte(elementLen & 0xFF),
		}
	} else if elementLen < (1<<28)-1 {
		return []byte{
			byte(0xE0 | (elementLen >> 24)),
			byte((elementLen >> 16) & 0xFF),
			byte((elementLen >> 8) & 0xFF),
			byte(elementLen & 0xFF),
		}
	} else {
		return []byte{
			0xF0,
			byte((elementLen >> 24) & 0xFF),
			byte((elementLen >> 16) & 0xFF),
			byte((elementLen >> 8) & 0xFF),
			byte(elementLen & 0xFF),
		}
	}
}

// writeStreamGroups writes stream groups
func (enc *Encoder) writeStreamGroups(groups []*model.StreamGroup, version uint) error {
	err := enc.writeLength(uint64(len(groups)))
	if err != nil {
		return err
	}

	for _, group := range groups {
		// Write group name
		err = enc.writeString(group.Name)
		if err != nil {
			return err
		}

		// Write last ID
		err = enc.writeStreamId(group.LastId)
		if err != nil {
			return err
		}

		// Write entries read (version 2+)
		if version >= 2 {
			err = enc.writeLength(group.EntriesRead)
			if err != nil {
				return err
			}
		}

		// Write pending list
		err = enc.writeLength(uint64(len(group.Pending)))
		if err != nil {
			return err
		}

		for _, pending := range group.Pending {
			// Write message ID
			msBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(msBytes, pending.Id.Ms)
			err = enc.write(msBytes)
			if err != nil {
				return err
			}

			seqBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(seqBytes, pending.Id.Sequence)
			err = enc.write(seqBytes)
			if err != nil {
				return err
			}

			// Write delivery time
			deliveryTimeBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(deliveryTimeBytes, pending.DeliveryTime)
			err = enc.write(deliveryTimeBytes)
			if err != nil {
				return err
			}

			// Write delivery count
			err = enc.writeLength(pending.DeliveryCount)
			if err != nil {
				return err
			}
		}

		// Write consumers
		err = enc.writeLength(uint64(len(group.Consumers)))
		if err != nil {
			return err
		}

		for _, consumer := range group.Consumers {
			// Write consumer name
			err = enc.writeString(consumer.Name)
			if err != nil {
				return err
			}

			// Write seen time
			seenTimeBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(seenTimeBytes, consumer.SeenTime)
			err = enc.write(seenTimeBytes)
			if err != nil {
				return err
			}

			// Write active time (version 3+)
			if version >= 3 {
				activeTimeBytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(activeTimeBytes, consumer.ActiveTime)
				err = enc.write(activeTimeBytes)
				if err != nil {
					return err
				}
			}

			// Write consumer pending list
			err = enc.writeLength(uint64(len(consumer.Pending)))
			if err != nil {
				return err
			}

			for _, pendingId := range consumer.Pending {
				// Write message ID
				msBytes := make([]byte, 8)
				binary.BigEndian.PutUint64(msBytes, pendingId.Ms)
				err = enc.write(msBytes)
				if err != nil {
					return err
				}

				seqBytes := make([]byte, 8)
				binary.BigEndian.PutUint64(seqBytes, pendingId.Sequence)
				err = enc.write(seqBytes)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// encodeListPackInt encodes an integer for listpack
func (enc *Encoder) encodeListPackInt(val int64) []byte {
	if val >= -127 && val <= 127 {
		// 0xxxxxxx, uint7
		return []byte{byte(val)}
	} else if val >= -8191 && val <= 8191 {
		// 110xxxxx yyyyyyyy, int13
		uval := uint16(val)
		if val < 0 {
			uval = uint16(8191 + val + 1)
		}
		return []byte{
			byte(0xC0 | (uval >> 8)),
			byte(uval & 0xFF),
		}
	} else if val >= -32767 && val <= 32767 {
		// 11110001 aaaaaaaa bbbbbbbb, int16
		uval := uint16(val)
		return []byte{
			0xF1,
			byte(uval & 0xFF),
			byte(uval >> 8),
		}
	} else if val >= -8388607 && val <= 8388607 {
		// 11110010 aaaaaaaa bbbbbbbb cccccccc, int24
		uval := uint32(val)
		return []byte{
			0xF2,
			byte(uval & 0xFF),
			byte((uval >> 8) & 0xFF),
			byte((uval >> 16) & 0xFF),
		}
	} else if val >= -2147483647 && val <= 2147483647 {
		// 11110011 aaaaaaaa bbbbbbbb cccccccc dddddddd, int32
		uval := uint32(val)
		return []byte{
			0xF3,
			byte(uval & 0xFF),
			byte((uval >> 8) & 0xFF),
			byte((uval >> 16) & 0xFF),
			byte((uval >> 24) & 0xFF),
		}
	} else {
		// 11110100 8Byte -> int64
		uval := uint64(val)
		result := []byte{0xF4}
		for i := 0; i < 8; i++ {
			result = append(result, byte(uval&0xFF))
			uval >>= 8
		}
		return result
	}
}

// encodeListPackString encodes a string for listpack
func (enc *Encoder) encodeListPackString(s string) []byte {
	bytes := []byte(s)
	length := len(bytes)

	if length <= 63 {
		// 10xxxxxx + content, string(len<=63)
		return append([]byte{byte(0x80 | length)}, bytes...)
	} else if length < 4096 {
		// 1110xxxx yyyyyyyy + content, string(len < 1<<12)
		header := make([]byte, 2)
		header[0] = byte(0xE0 | (length >> 8))
		header[1] = byte(length & 0xFF)
		return append(header, bytes...)
	} else {
		// 11110000 aaaaaaaa bbbbbbbb cccccccc dddddddd + content, string(len < 1<<32)
		header := make([]byte, 5)
		header[0] = 0xF0
		binary.LittleEndian.PutUint32(header[1:5], uint32(length))
		return append(header, bytes...)
	}
}
