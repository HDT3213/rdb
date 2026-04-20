package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/hdt3213/rdb/bytefmt"
	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

type StatsReport struct {
	File           string               `json:"file"`
	RedisVersion   string               `json:"redisVersion"`
	RDBVersion     int                  `json:"rdbVersion"`
	KeyStats       KeyStatistics        `json:"keyStatistics"`
	MemoryStats    MemoryStatistics     `json:"memoryStatistics"`
	ExpireStats    ExpirationStatistics `json:"expirationStatistics"`
	LFUStats       *LFUStatistics       `json:"lfuStatistics,omitempty"`
	TopLargestKeys []KeyInfo            `json:"topLargestKeys"`
	TopHottestKeys []KeyWithFreq        `json:"topHottestKeys,omitempty"`
}

type KeyStatistics struct {
	Total  int64            `json:"total"`
	ByType map[string]int64 `json:"byType"`
	ByDB   map[int]int64    `json:"byDatabase"`
}

type MemoryStatistics struct {
	Total      int64                  `json:"total"`
	TotalFmt   string                 `json:"totalReadable"`
	ByType     map[string]MemoryStats `json:"byType"`
	ByEncoding map[string]MemoryStats `json:"byEncoding"`
}

type MemoryStats struct {
	Bytes      int64   `json:"bytes"`
	Percentage float64 `json:"percentage"`
	Readable   string  `json:"readable"`
}

type ExpirationStatistics struct {
	WithTTL    int64   `json:"withTTL"`
	WithoutTTL int64   `json:"withoutTTL"`
	Percentage float64 `json:"percentage"`
}

type LFUStatistics struct {
	Available bool    `json:"available"`
	Policy    string  `json:"policy"`
	Total     int64   `json:"total"`
	MinFreq   int64   `json:"minFreq"`
	MaxFreq   int64   `json:"maxFreq"`
	AvgFreq   float64 `json:"avgFreq"`
}

type KeyInfo struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Readable string `json:"readable"`
}

type KeyWithFreq struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	Freq int64  `json:"freq"`
}

type sizedKeyInfo struct {
	key  string
	typ  string
	size int
}

func (s *sizedKeyInfo) GetSize() int {
	return s.size
}

type statsCollector struct {
	report     *StatsReport
	lfuMin     int64
	lfuMax     int64
	lfuSum     int64
	lfuCount   int64
	lfuPolicy  string
	hasLFU     bool
	topLargest []*sizedKeyInfo
	topHottest []*KeyWithFreq
	topN       int
}

type freqTopList struct {
	list     []*KeyWithFreq
	capacity int
}

func (tl *freqTopList) add(key string, objType string, freq int64) {
	if freq <= 0 {
		return
	}
	info := &KeyWithFreq{
		Key:  key,
		Type: objType,
		Freq: freq,
	}
	index := sort.Search(len(tl.list), func(i int) bool {
		return tl.list[i].Freq <= freq
	})
	tl.list = append(tl.list, info)
	copy(tl.list[index+1:], tl.list[index:])
	tl.list[index] = info
	if len(tl.list) > tl.capacity {
		tl.list = tl.list[:tl.capacity]
	}
}

func newFreqToplist(cap int) *freqTopList {
	return &freqTopList{
		capacity: cap,
	}
}

func newStatsCollector(topN int) *statsCollector {
	sc := &statsCollector{
		report: &StatsReport{
			KeyStats: KeyStatistics{
				ByType: make(map[string]int64),
				ByDB:   make(map[int]int64),
			},
			MemoryStats: MemoryStatistics{
				ByType:     make(map[string]MemoryStats),
				ByEncoding: make(map[string]MemoryStats),
			},
			TopLargestKeys: make([]KeyInfo, 0),
		},
		topLargest: make([]*sizedKeyInfo, 0, topN),
		topHottest: make([]*KeyWithFreq, 0, topN),
		topN:       topN,
	}
	return sc
}

func (sc *statsCollector) processObject(object model.RedisObject) {
	objType := object.GetType()

	if objType == model.AuxType || objType == model.FunctionsType {
		return
	}

	sc.report.KeyStats.Total++
	sc.report.KeyStats.ByType[objType]++

	db := object.GetDBIndex()
	sc.report.KeyStats.ByDB[db]++

	size := int64(object.GetSize())
	sc.report.MemoryStats.Total += size

	memType, ok := sc.report.MemoryStats.ByType[objType]
	if !ok {
		memType = MemoryStats{}
	}
	memType.Bytes += size
	sc.report.MemoryStats.ByType[objType] = memType

	encoding := object.GetEncoding()
	memEnc, ok := sc.report.MemoryStats.ByEncoding[encoding]
	if !ok {
		memEnc = MemoryStats{}
	}
	memEnc.Bytes += size
	sc.report.MemoryStats.ByEncoding[encoding] = memEnc

	if object.GetExpiration() != nil {
		sc.report.ExpireStats.WithTTL++
	} else {
		sc.report.ExpireStats.WithoutTTL++
	}

	if evictionInfo, ok := object.(model.EvictionInfo); ok {
		freq := evictionInfo.GetFreq()
		if freq > 0 {
			sc.hasLFU = true
			sc.lfuCount++
			sc.lfuSum += freq
			if sc.lfuMin == 0 || freq < sc.lfuMin {
				sc.lfuMin = freq
			}
			if freq > sc.lfuMax {
				sc.lfuMax = freq
			}
			sc.addTopHottest(object.GetKey(), objType, freq)
		}
	}

	sc.addTopLargest(object.GetKey(), objType, object.GetSize())
}

func (sc *statsCollector) addTopLargest(key, objType string, size int) {
	info := &sizedKeyInfo{
		key:  key,
		typ:  objType,
		size: size,
	}

	if len(sc.topLargest) < sc.topN {
		sc.topLargest = append(sc.topLargest, info)
		if len(sc.topLargest) == sc.topN {
			sc.sortTopLargest()
		}
		return
	}

	if size > sc.topLargest[0].size {
		sc.topLargest[0] = info
		sc.sortTopLargest()
	}
}

func (sc *statsCollector) sortTopLargest() {
	sort.Slice(sc.topLargest, func(i, j int) bool {
		return sc.topLargest[i].size > sc.topLargest[j].size
	})
}

func (sc *statsCollector) addTopHottest(key, objType string, freq int64) {
	info := &KeyWithFreq{
		Key:  key,
		Type: objType,
		Freq: freq,
	}

	if len(sc.topHottest) < sc.topN {
		sc.topHottest = append(sc.topHottest, info)
		if len(sc.topHottest) == sc.topN {
			sc.sortTopHottest()
		}
		return
	}

	if freq > sc.topHottest[0].Freq {
		sc.topHottest[0] = info
		sc.sortTopHottest()
	}
}

func (sc *statsCollector) sortTopHottest() {
	sort.Slice(sc.topHottest, func(i, j int) bool {
		return sc.topHottest[i].Freq > sc.topHottest[j].Freq
	})
}

func (sc *statsCollector) finalize() {
	sc.report.MemoryStats.TotalFmt = bytefmt.FormatSize(uint64(sc.report.MemoryStats.Total))

	totalMem := sc.report.MemoryStats.Total
	for objType, memStats := range sc.report.MemoryStats.ByType {
		pct := float64(memStats.Bytes) / float64(totalMem) * 100
		sc.report.MemoryStats.ByType[objType] = MemoryStats{
			Bytes:      memStats.Bytes,
			Percentage: float64(int(pct*10)) / 10,
			Readable:   bytefmt.FormatSize(uint64(memStats.Bytes)),
		}
	}

	for encoding, memStats := range sc.report.MemoryStats.ByEncoding {
		sc.report.MemoryStats.ByEncoding[encoding] = MemoryStats{
			Bytes:    memStats.Bytes,
			Readable: bytefmt.FormatSize(uint64(memStats.Bytes)),
		}
	}

	totalKeys := sc.report.ExpireStats.Total()
	if totalKeys > 0 {
		sc.report.ExpireStats.Percentage = float64(int(float64(sc.report.ExpireStats.WithTTL)/float64(totalKeys)*1000)) / 10
	}

	if sc.hasLFU {
		sc.report.LFUStats = &LFUStatistics{
			Available: true,
			Policy:    sc.lfuPolicy,
			Total:     sc.lfuCount,
			MinFreq:   sc.lfuMin,
			MaxFreq:   sc.lfuMax,
			AvgFreq:   float64(int(float64(sc.lfuSum)/float64(sc.lfuCount)*100)) / 100,
		}
	}

	for _, item := range sc.topLargest {
		sc.report.TopLargestKeys = append(sc.report.TopLargestKeys, KeyInfo{
			Key:      item.key,
			Type:     item.typ,
			Size:     int64(item.size),
			Readable: bytefmt.FormatSize(uint64(item.size)),
		})
	}

	for _, item := range sc.topHottest {
		sc.report.TopHottestKeys = append(sc.report.TopHottestKeys, *item)
	}
}

func (s ExpirationStatistics) Total() int64 {
	return s.WithTTL + s.WithoutTTL
}

func Stats(rdbFilename string, topN int, options ...interface{}) (*StatsReport, error) {
	if rdbFilename == "" {
		return nil, fmt.Errorf("src file path is required")
	}

	rdbFile, err := os.Open(rdbFilename)
	if err != nil {
		return nil, fmt.Errorf("open rdb %s failed: %v", rdbFilename, err)
	}
	defer rdbFile.Close()

	collector := newStatsCollector(topN)
	if topN <= 0 {
		topN = 10
	}

	coreDec := core.NewDecoder(rdbFile)
	var dec decoder = coreDec
	dec, err = wrapDecoder(dec, options...)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = dec.Parse(func(object model.RedisObject) bool {
		if collector.lfuPolicy == "" {
			if evictionInfo, ok := object.(model.EvictionInfo); ok {
				if freq := evictionInfo.GetFreq(); freq > 0 {
					collector.lfuPolicy = "allkeys-lfu"
				}
			}
		}

		if object.GetExpiration() != nil && object.GetExpiration().Before(now) {
			collector.report.ExpireStats.WithTTL--
			collector.report.ExpireStats.WithoutTTL++
		}

		if auxObj, ok := object.(*model.AuxObject); ok {
			if auxObj.Key == "redis-ver" {
				collector.report.RedisVersion = auxObj.Value
			}
		}

		collector.processObject(object)
		return true
	})
	if err != nil {
		return nil, err
	}

	collector.report.RDBVersion = coreDec.GetRDBVersion()
	collector.finalize()
	collector.report.File = rdbFilename

	return collector.report, nil
}

func (r *StatsReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r *StatsReport) ToText() string {
	var lines []string

	lines = append(lines, "=== RDB Statistics Overview ===")
	lines = append(lines, fmt.Sprintf("File: %s", r.File))
	lines = append(lines, fmt.Sprintf("Redis Version: %s", r.RedisVersion))
	lines = append(lines, fmt.Sprintf("RDB Version: %d", r.RDBVersion))
	lines = append(lines, "")

	lines = append(lines, "--- Key Statistics ---")
	lines = append(lines, fmt.Sprintf("Total Keys: %d", r.KeyStats.Total))
	lines = append(lines, "Keys by Type:")
	typeKeys := make([]string, 0, len(r.KeyStats.ByType))
	for k := range r.KeyStats.ByType {
		typeKeys = append(typeKeys, k)
	}
	sort.Strings(typeKeys)
	for _, t := range typeKeys {
		count := r.KeyStats.ByType[t]
		pct := float64(count) / float64(r.KeyStats.Total) * 100
		lines = append(lines, fmt.Sprintf("  - %s: %d (%.1f%%)", t, count, pct))
	}
	lines = append(lines, "Keys by Database:")
	dbKeys := make([]int, 0, len(r.KeyStats.ByDB))
	for k := range r.KeyStats.ByDB {
		dbKeys = append(dbKeys, k)
	}
	sort.Ints(dbKeys)
	for _, db := range dbKeys {
		count := r.KeyStats.ByDB[db]
		pct := float64(count) / float64(r.KeyStats.Total) * 100
		lines = append(lines, fmt.Sprintf("  - DB %d: %d (%.1f%%)", db, count, pct))
	}
	lines = append(lines, "")

	lines = append(lines, "--- Memory Statistics ---")
	lines = append(lines, fmt.Sprintf("Total Memory: %s", r.MemoryStats.TotalFmt))
	lines = append(lines, "Memory by Type:")
	for _, t := range typeKeys {
		memStats := r.MemoryStats.ByType[t]
		lines = append(lines, fmt.Sprintf("  - %s: %s (%.1f%%)", t, memStats.Readable, memStats.Percentage))
	}
	lines = append(lines, "Memory by Encoding:")
	encKeys := make([]string, 0, len(r.MemoryStats.ByEncoding))
	for k := range r.MemoryStats.ByEncoding {
		encKeys = append(encKeys, k)
	}
	sort.Strings(encKeys)
	for _, enc := range encKeys {
		memStats := r.MemoryStats.ByEncoding[enc]
		lines = append(lines, fmt.Sprintf("  - %s: %s", enc, memStats.Readable))
	}
	lines = append(lines, "")

	lines = append(lines, "--- Expiration Statistics ---")
	lines = append(lines, fmt.Sprintf("Keys with TTL: %d (%.1f%%)", r.ExpireStats.WithTTL, r.ExpireStats.Percentage))
	lines = append(lines, fmt.Sprintf("Keys without TTL: %d", r.ExpireStats.WithoutTTL))
	lines = append(lines, "")

	if r.LFUStats != nil && r.LFUStats.Available {
		lines = append(lines, "--- LFU Statistics ---")
		policy := r.LFUStats.Policy
		if policy == "" {
			policy = "allkeys-lfu"
		}
		lines = append(lines, fmt.Sprintf("LFU Info Available: Yes (maxmemory-policy: %s)", policy))
		lines = append(lines, fmt.Sprintf("Total Keys with LFU: %d", r.LFUStats.Total))
		lines = append(lines, fmt.Sprintf("Min Frequency: %d", r.LFUStats.MinFreq))
		lines = append(lines, fmt.Sprintf("Max Frequency: %d", r.LFUStats.MaxFreq))
		lines = append(lines, fmt.Sprintf("Avg Frequency: %.2f", r.LFUStats.AvgFreq))
		lines = append(lines, "")
	}

	if len(r.TopLargestKeys) > 0 {
		lines = append(lines, fmt.Sprintf("--- Top %d Largest Keys ---", len(r.TopLargestKeys)))
		for i, key := range r.TopLargestKeys {
			lines = append(lines, fmt.Sprintf("#%d. %s (%s) - %s", i+1, key.Key, key.Type, key.Readable))
		}
		lines = append(lines, "")
	}

	if len(r.TopHottestKeys) > 0 {
		lines = append(lines, fmt.Sprintf("--- Top %d Hottest Keys (by LFU) ---", len(r.TopHottestKeys)))
		for i, key := range r.TopHottestKeys {
			lines = append(lines, fmt.Sprintf("#%d. %s (%s) - freq: %d", i+1, key.Key, key.Type, key.Freq))
		}
		lines = append(lines, "")
	}

	return joinLines(lines)
}

func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
