package helper

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/model"
)

type decoder interface {
	Parse(cb func(object model.RedisObject) bool) error
}

type regexDecoder struct {
	reg *regexp.Regexp
	dec decoder
}

func (d *regexDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		if d.reg.MatchString(object.GetKey()) {
			return cb(object)
		}
		return true
	})
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

type ExpirationOption string

func WithExpirationOption(expr string) ExpirationOption {
	return ExpirationOption(expr)
}

// SizeOption filters by object size
type SizeOption string

// WithSizeOption creates a SizeOption from size expression
// expression format: "<min>~<max>", supports KB/MB/GB/TB/PB/EB units and 'inf'
func WithSizeOption(expr string) SizeOption {
	return SizeOption(expr)
}

// expirationDecoder returns entries with expiration times and expiration within the range.
type expirationDecoder struct {
	dec             decoder
	expirationRange []int64
}

// Parse returns entries with expiration times and expiration within the range.

func (d *expirationDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		expiration := object.GetExpiration()
		if expiration != nil {
			timestamp := expiration.Unix()
			if timestamp >= d.expirationRange[0] && timestamp <= d.expirationRange[1] {
				return cb(object)
			}
		}
		return true
	})
}

// sizeDecoder returns entries with size within the range.
type sizeDecoder struct {
	dec       decoder
	sizeRange []int
}

func (d *sizeDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		size := object.GetSize()
		if size >= d.sizeRange[0] && size <= d.sizeRange[1] {
			return cb(object)
		}
		return true
	})
}

// noExpirationDecoder returns entries without expiration
type noExpirationDecoder struct {
	dec decoder
}

func (d *noExpirationDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		expiration := object.GetExpiration()
		if expiration == nil {
			return cb(object)
		}
		return true
	})
}

func parseExpireExpr(s string) ([]int64, error) {
	parseValue := func(s string) (int64, error) {
		if s == "now" {
			return time.Now().Unix(), nil
		}
		if s == "inf" {
			return math.MaxInt64, nil
		}
		return strconv.ParseInt(s, 10, 64)
	}

	parts := strings.Split(s, "~")
	if len(parts) != 2 {
		return nil, errors.New("illegal expr, should be timestamp1~timestamp2")
	}

	min, err := parseValue(parts[0])
	if err != nil {
		return nil, fmt.Errorf("illegal range begin")
	}
	max, err := parseValue(parts[1])
	if err != nil {
		return nil, fmt.Errorf("illegal range end")
	}
	return []int64{min, max}, nil
}

func parseSizeExpr(s string) ([]int, error) {
	parseValue := func(s string) (int, error) {
		if s == "inf" {
			return math.MaxInt, nil
		}
		val, err := bytefmt.ParseSize(s)
		if err != nil {
			return 0, fmt.Errorf("illegal size: %v", s)
		}
		if val > uint64(math.MaxInt) {
			return math.MaxInt, nil
		}
		return int(val), nil
	}
	parts := strings.Split(s, "~")
	if len(parts) != 2 {
		return nil, errors.New("illegal expr, should be size1~size2")
	}
	min, err := parseValue(parts[0])
	if err != nil {
		return nil, fmt.Errorf("illegal range begin")
	}
	max, err := parseValue(parts[1])
	if err != nil {
		return nil, fmt.Errorf("illegal range end")
	}
	return []int{min, max}, nil
}

func wrapDecoder(dec decoder, options ...interface{}) (decoder, error) {
	var regexOpt RegexOption
	var noExpiredOpt NoExpiredOption
	var expirationOpt ExpirationOption
	var sizeOpt SizeOption
	for _, opt := range options {
		switch o := opt.(type) {
		case RegexOption:
			regexOpt = o
		case NoExpiredOption:
			noExpiredOpt = o
		case ExpirationOption:
			expirationOpt = o
		case SizeOption:
			sizeOpt = o
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
	if expirationOpt != "" {
		if expirationOpt == "noexpire" {
			dec = &noExpirationDecoder{
				dec: dec,
			}
		} else if expirationOpt == "anyexpire" {
			dec = &expirationDecoder{
				dec:             dec,
				expirationRange: []int64{0, math.MaxInt64},
			}
		} else {
			rng, err := parseExpireExpr(string(expirationOpt))
			if err != nil {
				return nil, err
			}
			dec = &expirationDecoder{
				dec:             dec,
				expirationRange: rng,
			}
		}
	}
	if sizeOpt != "" {
		rng, err := parseSizeExpr(string(sizeOpt))
		if err != nil {
			return nil, err
		}
		dec = &sizeDecoder{
			dec:       dec,
			sizeRange: rng,
		}
	}
	return dec, nil
}
