package memprofiler

import "github.com/hdt3213/rdb/model"

func sizeOfStreamObject(obj *model.StreamObject) int {
	size := sizeOfPointer()*2 + 8 + 16 + // size of stream struct
		sizeOfPointer() + 8*2 + // rax struct
		sizeOfStreamRaxTree(len(obj.Entries))
	if obj.Version >= 2 {
		size += 16*2 + 8 // size of 2 new streamID and a uint64
	}
	for _, group := range obj.Groups {
		size += sizeOfPointer()*2 + 16 // size of struct streamCG
		if obj.Version >= 2 {
			size += 8 // size of new field entries_read
		}
		pendingCount := len(group.Pending)
		size += sizeOfStreamRaxTree(pendingCount) +
			pendingCount*(sizeOfPointer()+8*2) //  streamNACK
		for _, consumer := range group.Consumers {
			size += sizeOfPointer()*2 + 8 + // streamConsumer
				sizeOfString(consumer.Name) +
				sizeOfStreamRaxTree(len(consumer.Pending))
		}
	}
	return size
}

func sizeOfStreamRaxTree(elementCount int) int {
	// This is a very rough estimation. The only alternative to doing an estimation,
	// is to fully build a radix tree of similar design, and elementCount the nodes.
	// There should be at least as many nodes as there are elements in the radix tree (possibly up to 3 times)
	nodeCount := int(float64(elementCount) * 2.5)
	// formula for memory estimation copied from Redis's streamRadixTreeMemoryUsage
	return 16*elementCount + 4*nodeCount + 30*sizeOfLong()*nodeCount
}
