package helper

import (
	"github.com/hdt3213/rdb/model"
	"math/rand"
	"sort"
	"strconv"
	"testing"
)

func TestRedisHeap_Append(t *testing.T) {
	sizeMap := make(map[int]struct{}) // The behavior when encountering objects of the same size is undefined
	topN := 100
	n := topN * 10
	objects := make([]model.RedisObject, 0)
	for i := 0; i < n; i++ {
		var size int
		for {
			size = rand.Intn(n * 10)
			if _, ok := sizeMap[size]; !ok {
				sizeMap[size] = struct{}{}
				break
			}
		}
		o := &model.StringObject{
			BaseObject: &model.BaseObject{
				Key:  strconv.Itoa(i),
				Size: size,
			},
		}
		objects = append(objects, o)
	}
	topList := newRedisHeap(topN)
	for _, o := range objects {
		topList.Append(o)
	}
	actual := topList.Dump()
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].GetSize() > objects[j].GetSize()
	})
	expect := objects[0:topN]
	for i := 0; i < topN; i++ {
		o1 := actual[i]
		o2 := expect[i]
		if o1.GetSize() != o2.GetSize() {
			t.Errorf("wrong answer at index: %d", i)
		}
	}
}
