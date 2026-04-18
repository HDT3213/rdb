package helper

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdt3213/rdb/encoder"
)

// makeTestRDB creates a temporary RDB file with the given string keys (all values are "v").
func makeTestRDB(t *testing.T, path string, keys []string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create rdb failed: %v", err)
	}
	defer f.Close()

	enc := encoder.NewEncoder(f)
	if err = enc.WriteHeader(); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if err = enc.WriteDBHeader(0, uint64(len(keys)), 0); err != nil {
		t.Fatalf("write db header: %v", err)
	}
	for _, key := range keys {
		if err = enc.WriteStringObject(key, []byte("v")); err != nil {
			t.Fatalf("write string %q: %v", key, err)
		}
	}
	if err = enc.WriteEnd(); err != nil {
		t.Fatalf("write end: %v", err)
	}
}

// readCSVLines reads a file and returns non-empty lines.
func readCSVLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func TestSepPrefixAnalyse(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()

	rdbPath := filepath.Join("tmp", "sep_test.rdb")
	makeTestRDB(t, rdbPath, []string{
		"user:alice",
		"user:bob",
		"user:charlie",
		"post:100",
		"post:200",
		"stat:user:active",
		"stat:user:inactive",
		"stat:post:views",
		"solo", // no separator — should not generate any prefix
	})

	outPath := filepath.Join("tmp", "sep_test.csv")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	err = SepPrefixAnalyse(rdbPath, 0, 0, []string{":"}, outFile, )
	if err != nil {
		t.Fatalf("SepPrefixAnalyse failed: %v", err)
	}
	_ = outFile.Close()

	lines := readCSVLines(t, outPath)
	// header + data lines
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines (header + data), got %d", len(lines))
	}

	// check header
	if lines[0] != "database,prefix,size,size_readable,key_count" {
		t.Errorf("unexpected header: %s", lines[0])
	}

	// collect prefixes from output
	prefixSet := make(map[string]bool)
	for _, line := range lines[1:] {
		parts := strings.SplitN(line, ",", 3)
		if len(parts) >= 2 {
			prefixSet[parts[1]] = true
		}
	}

	// should contain expected prefixes
	for _, expected := range []string{"user:*", "post:*", "stat:*", "stat:user:*", "stat:post:*"} {
		if !prefixSet[expected] {
			t.Errorf("expected prefix %q not found in output", expected)
		}
	}

	// "solo" has no separator so should NOT appear as a prefix
	for prefix := range prefixSet {
		if strings.HasPrefix(prefix, "solo") {
			t.Errorf("unexpected prefix for key without separator: %s", prefix)
		}
	}
}

func TestSepPrefixAnalyseTopN(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()

	rdbPath := filepath.Join("tmp", "sep_topn.rdb")
	makeTestRDB(t, rdbPath, []string{
		"a:1", "a:2", "a:3",
		"b:1", "b:2",
		"c:1",
	})

	outPath := filepath.Join("tmp", "sep_topn.csv")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	err = SepPrefixAnalyse(rdbPath, 2, 0, []string{":"}, outFile)
	if err != nil {
		t.Fatalf("SepPrefixAnalyse failed: %v", err)
	}
	_ = outFile.Close()

	lines := readCSVLines(t, outPath)
	// header + at most 2 data lines
	dataLines := lines[1:]
	if len(dataLines) != 2 {
		t.Errorf("expected 2 data lines (topN=2), got %d", len(dataLines))
	}
}

func TestSepPrefixAnalyseMaxDepth(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()

	rdbPath := filepath.Join("tmp", "sep_depth.rdb")
	makeTestRDB(t, rdbPath, []string{
		"a:b:c:d", // depth 3 key
		"a:b:x:y", // depth 3 key
	})

	outPath := filepath.Join("tmp", "sep_depth.csv")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	// maxDepth=1: should only see "a:*", not "a:b:*" or deeper
	err = SepPrefixAnalyse(rdbPath, 0, 1, []string{":"}, outFile)
	if err != nil {
		t.Fatalf("SepPrefixAnalyse failed: %v", err)
	}
	_ = outFile.Close()

	lines := readCSVLines(t, outPath)
	dataLines := lines[1:]
	if len(dataLines) != 1 {
		t.Errorf("expected 1 prefix with maxDepth=1, got %d", len(dataLines))
	}
	if len(dataLines) > 0 && !strings.Contains(dataLines[0], "a:*") {
		t.Errorf("expected prefix a:*, got: %s", dataLines[0])
	}
}

func TestSepPrefixAnalyseMultiSep(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()

	rdbPath := filepath.Join("tmp", "sep_multi.rdb")
	// keys use different separators: ":" and "."
	makeTestRDB(t, rdbPath, []string{
		"user:alice",
		"user.bob",     // "." normalized to ":"
		"post:100",
		"post.200",     // "." normalized to ":"
	})

	outPath := filepath.Join("tmp", "sep_multi.csv")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	err = SepPrefixAnalyse(rdbPath, 0, 0, []string{":", "."}, outFile)
	if err != nil {
		t.Fatalf("SepPrefixAnalyse failed: %v", err)
	}
	_ = outFile.Close()

	lines := readCSVLines(t, outPath)
	prefixMap := make(map[string]string) // prefix -> full line
	for _, line := range lines[1:] {
		parts := strings.SplitN(line, ",", 3)
		if len(parts) >= 2 {
			prefixMap[parts[1]] = line
		}
	}

	// "user.bob" should be normalized to "user:bob", so "user:*" should have 2 keys
	if line, ok := prefixMap["user:*"]; ok {
		if !strings.HasSuffix(line, ",2") {
			t.Errorf("expected user:* key_count=2, got: %s", line)
		}
	} else {
		t.Error("expected prefix user:* not found")
	}

	if line, ok := prefixMap["post:*"]; ok {
		if !strings.HasSuffix(line, ",2") {
			t.Errorf("expected post:* key_count=2, got: %s", line)
		}
	} else {
		t.Error("expected prefix post:* not found")
	}
}

func TestSepPrefixAnalyseErrors(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()

	outFile, _ := os.Create(filepath.Join("tmp", "err.csv"))
	defer outFile.Close()

	// empty filename
	err = SepPrefixAnalyse("", 0, 0, []string{":"}, outFile)
	if err == nil {
		t.Error("expected error for empty filename")
	}

	// empty separators
	err = SepPrefixAnalyse("tmp/test.rdb", 0, 0, []string{}, outFile)
	if err == nil {
		t.Error("expected error for empty separators")
	}

	// non-existent file
	err = SepPrefixAnalyse("/nonexistent/file.rdb", 0, 0, []string{":"}, outFile)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestSepPrefixAnalyseNoSepKeys(t *testing.T) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll("tmp")
	}()

	rdbPath := filepath.Join("tmp", "sep_nosep.rdb")
	// all keys have no separator
	makeTestRDB(t, rdbPath, []string{
		"alpha",
		"beta",
		"gamma",
	})

	outPath := filepath.Join("tmp", "sep_nosep.csv")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatal(err)
	}
	err = SepPrefixAnalyse(rdbPath, 0, 0, []string{":"}, outFile)
	if err != nil {
		t.Fatalf("SepPrefixAnalyse failed: %v", err)
	}
	_ = outFile.Close()

	lines := readCSVLines(t, outPath)
	// header only — no prefixes when no keys contain the separator
	dataLines := lines[1:]
	if len(dataLines) != 0 {
		t.Errorf("expected 0 data lines for keys without separator, got %d", len(dataLines))
	}
}
