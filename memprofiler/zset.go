package memprofiler

import (
	"github.com/hdt3213/rdb/model"
	"math"
	"math/rand"
)

func skipListOverhead(size int) int {
	return 2*sizeOfPointer() + hashtableOverhead(size) + (2*sizeOfPointer() + 16)
}

func skipListEntryOverhead() int {
	return hashTableEntryOverhead() + 2*sizeOfPointer() + 8 + int(math.Round(float64(sizeOfPointer()+8)*MathExpectationOfRandomLevel))
}

// MathExpectationOfRandomLevel is mathematical expectation of zsetRandomLevel(), used to guarantee the stable results
const MathExpectationOfRandomLevel = 1.33

func zsetRandomLevel() int {
	const maxLevel = 32
	const p = 0.25
	i := 1
	r := rand.Intn(0xFFFF)
	for r < 0xFFFF/4 {
		i++
		r = rand.Intn(0xFFFF)
		if i >= maxLevel {
			return maxLevel
		}
	}
	return i
}

func sizeOfZSetObject(o *model.ZSetObject) int {
	if o.GetEncoding() == model.ZipListEncoding {
		extra := o.Extra.(*model.ZiplistDetail)
		return extra.RawStringSize
	} else if o.GetEncoding() == model.ListPackEncoding {
		extra := o.Extra.(*model.ListpackDetail)
		return extra.RawStringSize
	}
	size := skipListOverhead(len(o.Entries))
	for _, entry := range o.Entries {
		size += sizeOfString(entry.Member) + 8 + skipListEntryOverhead() // size of score is 8 (double)
	}
	return size
}
