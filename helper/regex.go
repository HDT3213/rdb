package helper

import (
	"fmt"
	"github.com/hdt3213/rdb/model"
	"regexp"
	"time"
)

type decoder interface {
	Parse(cb func(object model.RedisObject) bool) error
}

type regexDecoder struct {
	reg    *regexp.Regexp
	notreg *regexp.Regexp
	dec    decoder
}

func (d *regexDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		// key not care
		if d.notreg != nil && d.notreg.MatchString(object.GetKey()) {
			return true
		} else {
			if d.reg != nil && d.reg.MatchString(object.GetKey()) {
				return cb(object)
			}
		}
		return true
	})
}

// regexWrapper returns
func regexWrapper(d decoder, expr string, notexpr string) (*regexDecoder, error) {
	reg, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("illegal regex expression: %v", expr)
	}
	if notexpr != "" {
		notreg, err := regexp.Compile(notexpr)
		if err != nil {
			return nil, fmt.Errorf("illegal notregex expression: %v", notexpr)
		}
		return &regexDecoder{
			dec:    d,
			reg:    reg,
			notreg: notreg,
		}, nil
	}
	return &regexDecoder{
		dec: d,
		reg: reg,
	}, nil
}

// RegexOption enable regex filters
type RegexOption *string
type NotRegexOption *string

// WithRegexOption creates a WithRegexOption from regex expression
func WithRegexOption(expr string) RegexOption {
	return &expr
}

func WithNotRegexOption(notExpr string) NotRegexOption {
	return &notExpr
}

// noExpiredDecoder filter all expired keys
type noExpiredDecoder struct {
	dec decoder
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

// NoExpiredOption tells decoder to filter all expired keys
type NoExpiredOption bool

// WithNoExpiredOption tells decoder to filter all expired keys
func WithNoExpiredOption() NoExpiredOption {
	return NoExpiredOption(true)
}

func wrapDecoder(dec decoder, options ...interface{}) (decoder, error) {
	var regexOpt RegexOption
	var notRegexOpt NotRegexOption
	var noExpiredOpt NoExpiredOption
	for _, opt := range options {
		switch o := opt.(type) {
		case RegexOption:
			regexOpt = o
		case NotRegexOption:
			notRegexOpt = o
		case NoExpiredOption:
			noExpiredOpt = o
		}
	}
	// if notRegexOpt not is nil, the regexOpt must not is nil
	// but regexOpt not is nil. the notRegexOpt maybe is nil
	if regexOpt != nil {
		var err error
		if notRegexOpt != nil {
			dec, err = regexWrapper(dec, *regexOpt, *notRegexOpt)
		} else {
			dec, err = regexWrapper(dec, *regexOpt, "")
		}
		if err != nil {
			return nil, err
		}
	}

	if noExpiredOpt {
		dec = &noExpiredDecoder{
			dec: dec,
		}
	}
	return dec, nil
}
