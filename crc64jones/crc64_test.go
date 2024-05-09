package crc64jones

import (
	"io"
	"testing"
)

func TestGolden(t *testing.T) {
	c := New()
	in := "123456789"
	io.WriteString(c, in)
	s := c.Sum64()
	if out := uint64(0xe9c6d914c4b8d9ca); s != out {
		t.Fatalf("jones crc64(%s) = 0x%x want 0x%x", in, s, out)
	}
}
