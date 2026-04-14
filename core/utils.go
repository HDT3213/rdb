package core

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"sync/atomic"
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

// original code
/*func (dec *Decoder) readByte() (byte, error) {
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
}*/

// readByte get 1 Byte from Decoder.cache
func (dec *Decoder) readByte() (byte, error) {
	for {
		pos := dec.cachePos.Load()
		length := dec.cacheLen.Load()

		if pos >= length {
			dec.refillMu.Lock()
			if dec.cachePos.Load() < dec.cacheLen.Load() {
				dec.refillMu.Unlock()
				continue
			}
			err := dec.refillCache()
			dec.refillMu.Unlock()
			if err != nil {
				return 0, err
			}
			continue
		}

		if dec.cachePos.CompareAndSwap(pos, pos+1) {
			atomic.AddInt64(&dec.readCount, 1)
			return dec.cache[pos], nil
		}
	}
}

// readFull get len(buf) Bytes from Decoder.cache
func (dec *Decoder) readFull(buf []byte) error {
	need := len(buf)
	copied := 0

	for copied < need {
		pos := dec.cachePos.Load()
		length := dec.cacheLen.Load()

		if pos >= length {
			dec.refillMu.Lock()
			if dec.cachePos.Load() < dec.cacheLen.Load() {
				dec.refillMu.Unlock()
				continue
			}
			err := dec.refillCache()
			dec.refillMu.Unlock()
			if err != nil {
				if copied > 0 && err == io.EOF {
					return io.ErrUnexpectedEOF
				}
				return err
			}
			continue
		}

		available := int(length - pos)
		toCopy := need - copied
		if toCopy > available {
			toCopy = available
		}

		if dec.cachePos.CompareAndSwap(pos, pos+uint32(toCopy)) {
			copy(buf[copied:copied+toCopy], dec.cache[pos:pos+uint32(toCopy)])
			copied += toCopy
			atomic.AddInt64(&dec.readCount, int64(toCopy))
		}
	}
	return nil
}

// refillCache read 1MiB from disk
func (dec *Decoder) refillCache() error {
	n, err := dec.input.Read(dec.cache)

	// eof
	if n == 0 && err == nil {
		return io.EOF
	}

	if n > 0 {
		dec.cachePos.Store(0)
		dec.cacheLen.Store(uint32(n))
		return nil
	}

	return err
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
