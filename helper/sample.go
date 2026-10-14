package helper

import (
	"math/rand"
	"time"

	"github.com/hdt3213/rdb/model"
)

// SampleOption enables sampling keys with a rate in (0, 1]
type SampleOption float64

// WithSampleOption creates a SampleOption from sample rate
func WithSampleOption(rate float64) SampleOption {
	return SampleOption(rate)
}

// WithSample creates a SampleOption from sample rate
func WithSample(rate float64) SampleOption {
	return SampleOption(rate)
}

type sampleDecoder struct {
	dec  decoder
	rate float64
	rnd  *rand.Rand
}

func newSampleDecoder(dec decoder, rate float64) *sampleDecoder {
	return &sampleDecoder{
		dec:  dec,
		rate: rate,
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func newSampleDecoderWithRand(dec decoder, rate float64, rnd *rand.Rand) *sampleDecoder {
	return &sampleDecoder{
		dec:  dec,
		rate: rate,
		rnd:  rnd,
	}
}

func (d *sampleDecoder) Parse(cb func(object model.RedisObject) bool) error {
	return d.dec.Parse(func(object model.RedisObject) bool {
		if d.rnd.Float64() < d.rate {
			return cb(object)
		}
		return true
	})
}
