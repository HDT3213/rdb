package bytefmt

import (
	"math"
	"testing"
)

func TestFormatSize(t *testing.T) {
	if FormatSize(0) != "0" {
		t.Error("format error")
	}
	if FormatSize(123) != "123B" {
		t.Error("format error")
	}
	if FormatSize(123*(1<<10)) != "123K" {
		t.Error("format error")
	}
	if FormatSize(123*(1<<20)) != "123M" {
		t.Error("format error")
	}
	if FormatSize(123*(1<<30)) != "123G" {
		t.Error("format error")
	}
	if FormatSize(123*(1<<40)) != "123T" {
		t.Error("format error")
	}
	if FormatSize(123*(1<<50)) != "123P" {
		t.Error("format error")
	}
	if FormatSize(math.MaxUint64) != "16E" {
		t.Error("format error")
	}
}

func TestParseSize(t *testing.T) {
	if _, err := ParseSize("0"); err != invalidByteQuantityError {
		t.Error("parse error")
	}
	if _, err := ParseSize("0B"); err != invalidByteQuantityError {
		t.Error("parse error")
	}
	if _, err := ParseSize("1A"); err != invalidByteQuantityError {
		t.Error("parse error")
	}
	if b, err := ParseSize("123B"); err != nil || b != 123 {
		t.Error("parse error")
	}
	if b, err := ParseSize("123K"); err != nil || b != 123*(1<<10) {
		t.Error("format error")
	}
	if b, err := ParseSize("123M"); err != nil || b != 123*(1<<20) {
		t.Error("format error")
	}
	if b, err := ParseSize("123G"); err != nil || b != 123*(1<<30) {
		t.Error("format error")
	}
	if b, err := ParseSize("123T"); err != nil || b != 123*(1<<40) {
		t.Error("format error")
	}
	if b, err := ParseSize("123P"); err != nil || b != 123*(1<<50) {
		t.Error("format error")
	}
	if b, err := ParseSize("1E"); err != nil || b != 1<<60 {
		t.Error("format error")
	}
}
