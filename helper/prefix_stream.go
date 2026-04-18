package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

type prefixStats struct {
	size     int
	keyCount int
}

// SepPrefixAnalyse reads an RDB file and aggregates memory usage by key prefix
// using a flat map (constant memory). Keys are split by the given separators up to maxDepth.
// Multiple separators are normalized to the first one before splitting.
func SepPrefixAnalyse(rdbFilename string, topN int, maxDepth int, separators []string, output *os.File, options ...interface{}) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
	}
	if len(separators) == 0 {
		return errors.New("at least one separator is required")
	}
	if topN <= 0 {
		topN = math.MaxInt
	}
	if maxDepth <= 0 {
		maxDepth = math.MaxInt
	}

	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return fmt.Errorf("open rdb %s failed, %v", rdbFilename, err)
	}
	defer rdbFile.Close()

	var dec decoder = core.NewDecoder(rdbFile)
	if dec, err = wrapDecoder(dec, options...); err != nil {
		return err
	}

	primarySep := separators[0]

	// flat map: "db\x00prefix" -> stats
	prefixes := make(map[string]*prefixStats)

	err = dec.Parse(func(object model.RedisObject) bool {
		key := object.GetKey()
		db := object.GetDBIndex()
		size := object.GetSize()

		// normalize all separators to the primary one
		normalizedKey := key
		for i := 1; i < len(separators); i++ {
			normalizedKey = strings.ReplaceAll(normalizedKey, separators[i], primarySep)
		}

		parts := strings.SplitN(normalizedKey, primarySep, maxDepth+1)

		// only emit prefixes that actually group keys —
		// skip depth == len(parts) since that's the full key, not a prefix
		limit := len(parts) - 1
		if limit > maxDepth {
			limit = maxDepth
		}

		for depth := 1; depth <= limit; depth++ {
			prefix := strings.Join(parts[:depth], primarySep) + primarySep + "*"
			mapKey := strconv.Itoa(db) + "\x00" + prefix

			s := prefixes[mapKey]
			if s == nil {
				s = &prefixStats{}
				prefixes[mapKey] = s
			}
			s.size += size
			s.keyCount++
		}

		return true
	})
	if err != nil {
		return err
	}

	// sort by size descending
	type entry struct {
		db       string
		prefix   string
		size     int
		keyCount int
	}
	entries := make([]entry, 0, len(prefixes))
	for mapKey, s := range prefixes {
		idx := strings.Index(mapKey, "\x00")
		entries = append(entries, entry{
			db:       mapKey[:idx],
			prefix:   mapKey[idx+1:],
			size:     s.size,
			keyCount: s.keyCount,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].size > entries[j].size
	})

	// write CSV
	_, err = output.WriteString("database,prefix,size,size_readable,key_count\n")
	if err != nil {
		return fmt.Errorf("write header failed: %v", err)
	}
	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()

	limit := topN
	if limit > len(entries) {
		limit = len(entries)
	}
	for i := 0; i < limit; i++ {
		e := entries[i]
		err = csvWriter.Write([]string{
			e.db,
			e.prefix,
			strconv.Itoa(e.size),
			bytefmt.FormatSize(uint64(e.size)),
			strconv.Itoa(e.keyCount),
		})
		if err != nil {
			return fmt.Errorf("csv write failed: %v", err)
		}
	}

	return nil
}
