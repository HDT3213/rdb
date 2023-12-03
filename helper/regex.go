package helper

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hdt3213/rdb/model"
)

type decoder interface {
	Parse(cb func(object model.RedisObject) bool) error
	SetFields(fields []string)
	GetFields() []string
}

type regexDecoder struct {
	reg    *regexp.Regexp
	dec    decoder
	fields []string
}

func (d *regexDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		if d.reg.MatchString(object.GetKey()) {
			return cb(object)
		}
		return true
	})
}

func (d *regexDecoder) SetFields(fields []string) {
	d.fields = fields
}

func (d *regexDecoder) GetFields() []string {
	return d.fields
}

// regexWrapper returns
func regexWrapper(d decoder, expr string) (*regexDecoder, error) {
	reg, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("illegal regex expression: %v", expr)
	}
	return &regexDecoder{
		dec: d,
		reg: reg,
	}, nil
}

// RegexOption enable regex filters
type RegexOption *string

// WithRegexOption creates a WithRegexOption from regex expression
func WithRegexOption(expr string) RegexOption {
	return &expr
}

// noExpiredDecoder filter all expired keys
type noExpiredDecoder struct {
	dec    decoder
	fields []string
}

func (d *noExpiredDecoder) Parse(cb func(object model.RedisObject) bool) error {
	now := time.Now()
	return d.dec.Parse(func(object model.RedisObject) bool {
		expiration := object.GetExpiration()
		if expiration == nil || expiration.After(now) {
			return cb(object)
		}
		return true
	})
}

func (d *noExpiredDecoder) SetFields(fields []string) {
	d.fields = fields
}

func (d *noExpiredDecoder) GetFields() []string {
	return d.fields
}

// NoExpiredOption tells decoder to filter all expired keys
type NoExpiredOption bool

// WithNoExpiredOption tells decoder to filter all expired keys
func WithNoExpiredOption() NoExpiredOption {
	return NoExpiredOption(true)
}

type FieldOption []string

func WithFieldOption(fields string) FieldOption {
	var fieldOpt FieldOption

	for _, i := range strings.Split(fields, ",") {
		fieldOpt = append(fieldOpt, i)
	}
	return fieldOpt
}

func wrapDecoder(dec decoder, options ...interface{}) (decoder, error) {
	var regexOpt RegexOption
	var noExpiredOpt NoExpiredOption
	var fieldOpt FieldOption
	for _, opt := range options {
		switch o := opt.(type) {
		case RegexOption:
			regexOpt = o
		case NoExpiredOption:
			noExpiredOpt = o
		case FieldOption:
			fieldOpt = o
		}
	}
	if regexOpt != nil {
		var err error
		dec, err = regexWrapper(dec, *regexOpt)
		if err != nil {
			return nil, err
		}
	}
	if noExpiredOpt {
		dec = &noExpiredDecoder{
			dec: dec,
		}
	}
	dec.SetFields(fieldOpt)
	return dec, nil
}
