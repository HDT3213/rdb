package helper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
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

// StreamingPrefixAnalyse reads an RDB file and aggregates memory usage by key prefix
// using constant memory (no radix tree). Keys are split by separator up to maxDepth.
// When trackMem is true, heap usage stats are printed to stderr every 1M keys.
func StreamingPrefixAnalyse(rdbFilename string, topN int, maxDepth int, separator string, output *os.File, options ...interface{}) error {
	return streamingPrefixAnalyse(rdbFilename, topN, maxDepth, separator, false, output, options...)
}

// StreamingPrefixAnalyseWithMemTrack is like StreamingPrefixAnalyse but prints
// heap memory usage to stderr periodically.
func StreamingPrefixAnalyseWithMemTrack(rdbFilename string, topN int, maxDepth int, separator string, output *os.File, options ...interface{}) error {
	return streamingPrefixAnalyse(rdbFilename, topN, maxDepth, separator, true, output, options...)
}

func streamingPrefixAnalyse(rdbFilename string, topN int, maxDepth int, separator string, trackMem bool, output *os.File, options ...interface{}) error {
	if rdbFilename == "" {
		return errors.New("src file path is required")
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

	// flat map: "db:prefix" -> stats. One entry per unique prefix, typically 1K-50K entries.
	prefixes := make(map[string]*prefixStats)
	var keysSeen int

	err = dec.Parse(func(object model.RedisObject) bool {
		key := object.GetKey()
		db := object.GetDBIndex()
		size := object.GetSize()

		// extract prefix at each depth level and accumulate
		var charMode bool
		var parts []string
		if separator == "" {
			// per-character split: prefix at depth N = first N chars of key
			charMode = true
			for _, ch := range key {
				parts = append(parts, string(ch))
			}
		} else {
			parts = strings.SplitN(key, separator, maxDepth+1)
		}
		var limit int
		if charMode {
			limit = len(parts)
			if limit > maxDepth {
				limit = maxDepth
			}
		} else {
			// only emit prefixes that actually group keys —
			// skip depth == len(parts) since that's the full key, not a prefix
			limit = len(parts) - 1
			if limit > maxDepth {
				limit = maxDepth
			}
		}
		for depth := 1; depth <= limit; depth++ {
			var prefix string
			if charMode {
				prefix = strings.Join(parts[:depth], "")
			} else {
				prefix = strings.Join(parts[:depth], separator)
				prefix += separator + "*"
			}
			mapKey := strconv.Itoa(db) + "\x00" + prefix

			s := prefixes[mapKey]
			if s == nil {
				s = &prefixStats{}
				prefixes[mapKey] = s
			}
			s.size += size
			s.keyCount++
		}

		keysSeen++
		if keysSeen%1_000_000 == 0 {
			if trackMem {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Fprintf(os.Stderr, "Processed %d keys, %d unique prefixes | heap: %s, sys: %s\n",
					keysSeen, len(prefixes),
					bytefmt.FormatSize(m.HeapAlloc),
					bytefmt.FormatSize(m.Sys))
			} else {
				fmt.Fprintf(os.Stderr, "Processed %d keys, %d unique prefixes\n", keysSeen, len(prefixes))
			}
		}
		return true
	})
	if err != nil {
		return err
	}

	if trackMem {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Fprintf(os.Stderr, "Done: %d keys, %d unique prefixes | peak heap: %s, sys: %s\n",
			keysSeen, len(prefixes),
			bytefmt.FormatSize(m.HeapAlloc),
			bytefmt.FormatSize(m.Sys))
	} else {
		fmt.Fprintf(os.Stderr, "Done: %d keys, %d unique prefixes\n", keysSeen, len(prefixes))
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
