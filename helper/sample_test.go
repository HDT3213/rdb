package helper

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hdt3213/rdb/model"
)

type mockDecoder struct {
	objects []model.RedisObject
}

func (m *mockDecoder) Parse(cb func(object model.RedisObject) bool) error {
	for _, obj := range m.objects {
		if !cb(obj) {
			break
		}
	}
	return nil
}

func generateMockObjects(n int) []model.RedisObject {
	objs := make([]model.RedisObject, n)
	for i := 0; i < n; i++ {
		objs[i] = &model.StringObject{
			BaseObject: &model.BaseObject{
				DB:  0,
				Key: "key_" + strconv.Itoa(i),
			},
			Value: []byte("val"),
		}
	}
	return objs
}

func TestSampleDecoderBasic(t *testing.T) {
	total := 1000
	mock := &mockDecoder{objects: generateMockObjects(total)}
	rnd := rand.New(rand.NewSource(42))
	sampler := newSampleDecoderWithRand(mock, 0.5, rnd)

	var sampled []model.RedisObject
	err := sampler.Parse(func(object model.RedisObject) bool {
		sampled = append(sampled, object)
		return true
	})
	if err != nil {
		t.Fatalf("unexpected error during parse: %v", err)
	}

	count := len(sampled)
	// With seed 42, 1000 items and rate 0.5, count should be reasonably close to 500 (e.g. 400 to 600)
	if count < 400 || count > 600 {
		t.Fatalf("expected around 500 sampled items, got %d", count)
	}
}

func TestSampleDecoderEarlyTermination(t *testing.T) {
	total := 100
	mock := &mockDecoder{objects: generateMockObjects(total)}
	rnd := rand.New(rand.NewSource(42))
	sampler := newSampleDecoderWithRand(mock, 1.0, rnd)

	count := 0
	err := sampler.Parse(func(object model.RedisObject) bool {
		count++
		return count < 5 // stop after 5 items
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected early termination at 5 items, got %d", count)
	}
}

func TestWrapDecoderWithSampleOption(t *testing.T) {
	mock := &mockDecoder{objects: generateMockObjects(10)}

	// Test invalid negative rate
	_, err := wrapDecoder(mock, WithSampleOption(-0.5))
	if err == nil {
		t.Fatal("expected error for negative sample rate")
	}

	// Test invalid zero rate
	_, err = wrapDecoder(mock, WithSampleOption(0))
	if err == nil {
		t.Fatal("expected error for zero sample rate")
	}

	// Test invalid rate > 1
	_, err = wrapDecoder(mock, WithSampleOption(1.5))
	if err == nil {
		t.Fatal("expected error for sample rate > 1")
	}

	// Test valid rate == 1.0
	dec, err := wrapDecoder(mock, WithSampleOption(1.0))
	if err != nil {
		t.Fatalf("unexpected error for sample rate 1.0: %v", err)
	}
	var count int
	err = dec.Parse(func(object model.RedisObject) bool {
		count++
		return true
	})
	if err != nil || count != 10 {
		t.Fatalf("expected all 10 items for rate 1.0, got %d (err: %v)", count, err)
	}

	// Test WithSample alias function
	dec2, err := wrapDecoder(mock, WithSample(0.5))
	if err != nil {
		t.Fatalf("unexpected error for WithSample(0.5): %v", err)
	}
	if dec2 == nil {
		t.Fatal("expected non-nil decoder")
	}
}

func TestSampleIntegrationWithRDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rdb_sample_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	srcRdb := filepath.Join("../cases", "memory.rdb")

	// 1. Test MemoryProfile with sampling
	memOut := filepath.Join(tmpDir, "memory_sampled.csv")
	err = MemoryProfile(srcRdb, memOut, WithSampleOption(1.0))
	if err != nil {
		t.Fatalf("MemoryProfile with sample 1.0 failed: %v", err)
	}
	if fi, err := os.Stat(memOut); err != nil || fi.Size() == 0 {
		t.Fatalf("expected non-empty memory csv output, got err=%v", err)
	}

	// 2. Test FindBiggestKeys with sampling
	bigkeyFile, err := os.Create(filepath.Join(tmpDir, "bigkey_sampled.csv"))
	if err != nil {
		t.Fatalf("failed to create bigkey file: %v", err)
	}
	defer func() {
		_ = bigkeyFile.Close()
	}()
	err = FindBiggestKeys(srcRdb, 5, bigkeyFile, WithSampleOption(1.0))
	if err != nil {
		t.Fatalf("FindBiggestKeys with sample 1.0 failed: %v", err)
	}

	// 3. Test ToJsons with sampling
	jsonOut := filepath.Join(tmpDir, "sampled.json")
	err = ToJsons(srcRdb, jsonOut, WithSampleOption(1.0), WithConcurrent(1))
	if err != nil {
		t.Fatalf("ToJsons with sample 1.0 failed: %v", err)
	}

	// 4. Test invalid sample rate propagation
	err = MemoryProfile(srcRdb, filepath.Join(tmpDir, "invalid.csv"), WithSampleOption(-0.1))
	if err == nil {
		t.Fatal("expected error from MemoryProfile with invalid sample rate")
	}
}

func TestSampleIntegrationToAOF(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rdb_aof_sample_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	srcRdb := filepath.Join("../cases", "memory.rdb")
	aofOut := filepath.Join(tmpDir, "sampled.aof")
	err = ToAOF(srcRdb, aofOut, WithSampleOption(1.0))
	if err != nil {
		t.Fatalf("ToAOF with sample 1.0 failed: %v", err)
	}
	if fi, err := os.Stat(aofOut); err != nil || fi.Size() == 0 {
		t.Fatalf("expected non-empty aof output, got err=%v", err)
	}
}

func TestSampleIntegrationPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rdb_prefix_sample_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	srcRdb := filepath.Join("../cases", "tree.rdb")

	sepOut, err := os.Create(filepath.Join(tmpDir, "sep_prefix.csv"))
	if err != nil {
		t.Fatalf("failed to create sep prefix file: %v", err)
	}
	err = SepPrefixAnalyse(srcRdb, 10, 3, []string{":"}, sepOut, WithSampleOption(1.0))
	_ = sepOut.Close()
	if err != nil {
		t.Fatalf("SepPrefixAnalyse with sample 1.0 failed: %v", err)
	}
	if fi, err := os.Stat(sepOut.Name()); err != nil || fi.Size() == 0 {
		t.Fatal("expected non-empty SepPrefixAnalyse output")
	}

	prefixOut, err := os.Create(filepath.Join(tmpDir, "prefix.csv"))
	if err != nil {
		t.Fatalf("failed to create prefix file: %v", err)
	}
	err = PrefixAnalyse(srcRdb, 10, 3, prefixOut, WithSampleOption(1.0))
	_ = prefixOut.Close()
	if err != nil {
		t.Fatalf("PrefixAnalyse with sample 1.0 failed: %v", err)
	}
	if fi, err := os.Stat(prefixOut.Name()); err != nil || fi.Size() == 0 {
		t.Fatal("expected non-empty PrefixAnalyse output")
	}
}

