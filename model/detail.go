package model

// QuicklistDetail stores detail for quicklist
type QuicklistDetail struct {
	// ZiplistStruct stores each ziplist within quicklist
	ZiplistStruct [][][]byte
}

// Quicklist2Detail stores detail for quicklist2
type Quicklist2Detail struct {
	// NodeEncodings means node of quicklist is QuicklistNodeContainerPlain or QuicklistNodeContainerPacked
	NodeEncodings []int
	// ListPackEntrySize stores sizes of each listPackEntry is node encoding is QuicklistNodeContainerPacked
	ListPackEntrySize [][]uint32
}

// IntsetDetail stores detail for intset
type IntsetDetail struct {
	RawStringSize int
}

// QuicklistNodeContainerPlain means node of quicklist is normal string
const QuicklistNodeContainerPlain = 1

// QuicklistNodeContainerPacked means node of quicklist is list pack
const QuicklistNodeContainerPacked = 2

// ZiplistDetail stores detail for ziplist
type ZiplistDetail struct {
	RawStringSize int
}

// ListpackDetail stores detail for listpack
type ListpackDetail struct {
	RawStringSize int
}
