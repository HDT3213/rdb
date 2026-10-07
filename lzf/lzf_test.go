package lzf

import (
	"math"
	"math/rand"
	"strings"
	"testing"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandString create a random string no longer than n
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestLzf(t *testing.T) {
	for i := 0; i < 10; i++ {
		str := strings.Repeat(RandString(128), 10)
		compressed, err := Compress([]byte(str))
		if err != nil {
			t.Error(err)
			return
		}
		decompressed, err := Decompress(compressed, len(compressed), len(str))
		if err != nil {
			t.Error(err)
			return
		}
		if str != string(decompressed) {
			t.Error("wrong decompressed")
			return
		}
	}
}

func TestDecompressRejectsImpossibleOutLen(t *testing.T) {
	// one literal byte cannot expand to math.MaxInt bytes
	if _, err := Decompress([]byte{0x00}, 1, math.MaxInt); err == nil {
		t.Fatal("expect an error")
	}
	// the largest valid ratio is accepted
	if _, err := Decompress([]byte{0x00, 'a'}, 2, 88*2); err != nil {
		t.Fatal(err)
	}
	if _, err := Decompress([]byte{0x00, 'a'}, 2, 88*2+1); err == nil {
		t.Fatal("expect an error")
	}
}

func TestDecompressEmptyInput(t *testing.T) {
	out, err := Decompress(nil, 0, 0)
	if err != nil || out != nil {
		t.Fatalf("got %v, %v", out, err)
	}
}
