package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"unsafe"
)

func readBytes(buf []byte, cursor *int, size int) ([]byte, error) {
	if cursor == nil {
		return nil, errors.New("cursor is nil")
	}
	if *cursor+size > len(buf) {
		return nil, errors.New("cursor out of range")
	}
	end := *cursor + size
	result := buf[*cursor:end]
	*cursor += int(size)
	return result, nil
}

func readByte(buf []byte, cursor *int) (byte, error) {
	if cursor == nil {
		return 0, errors.New("cursor is nil")
	}
	if *cursor >= len(buf) {
		return 0, errors.New("cursor out of range")
	}
	b := buf[*cursor]
	*cursor++
	return b, nil
}

func readZipListLength(buf []byte, cursor *int) int {
	start := *cursor + 8
	end := start + 2
	// zip list buf: [0, 4] -> zlbytes, [4:8] -> zltail, [8:10] -> zllen
	size := int(binary.LittleEndian.Uint16(buf[start:end]))
	*cursor += 10
	return size
}

func (dec *Decoder) readByte() (byte, error) {
	b, err := dec.input.ReadByte()
	if err != nil {
		return 0, err
	}
	dec.readCount++
	return b, nil
}

func (dec *Decoder) readFull(buf []byte) error {
	n, err := io.ReadFull(dec.input, buf)
	if err != nil {
		return err
	}
	dec.readCount += n
	return nil
}

// readChunk is the largest length that readBytes allocates before it reads data.
// A bigger declared length may be false, so readBytes grows the buffer as data arrives.
const readChunk = 1 << 20

// readBytes reads exactly n bytes from the input.
// A length up to readChunk is allocated at once. A longer length grows the
// buffer by doubling, so memory stays near twice the bytes that arrived.
func (dec *Decoder) readBytes(n uint64) ([]byte, error) {
	if n <= readChunk {
		buf := make([]byte, n)
		if err := dec.readFull(buf); err != nil {
			return nil, err
		}
		return buf, nil
	}
	if n > math.MaxInt {
		return nil, fmt.Errorf("declared length %d is too large", n)
	}
	size := int(n)
	buf := make([]byte, readChunk)
	if err := dec.readFull(buf); err != nil {
		return nil, err
	}
	for len(buf) < size {
		next := 2 * len(buf)
		if next > size || next < 0 {
			next = size
		}
		grown := make([]byte, next)
		copy(grown, buf)
		if err := dec.readFull(grown[len(buf):]); err != nil {
			// Part of the value already arrived, so the end of the input here is a truncated value.
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		buf = grown
	}
	return buf, nil
}

// maxCapHint limits the initial capacity taken from a count in the file.
const maxCapHint = 1024

// capHint returns an initial capacity for a count read from the file.
// The count is not trusted, so append grows the slice when real data arrives.
func capHint(n uint64) int {
	if n > maxCapHint {
		return maxCapHint
	}
	return int(n)
}

// checkCount rejects a count that cannot fit in a buffer of size bytes.
// Each counted item uses at least one byte.
func checkCount(count int64, size int) error {
	if count < 0 || count > int64(size) {
		return fmt.Errorf("count %d does not fit in a buffer of %d bytes", count, size)
	}
	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandString create a random string no longer than n
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func unsafeBytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
