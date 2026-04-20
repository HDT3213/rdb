package helper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStats(t *testing.T) {
	srcRdb := filepath.Join("../cases", "memory.rdb")
	report, err := Stats(srcRdb, 10)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if report.KeyStats.Total == 0 {
		t.Error("Expected non-zero total keys")
	}

	if report.MemoryStats.Total == 0 {
		t.Error("Expected non-zero total memory")
	}

	textOutput := report.ToText()
	if textOutput == "" {
		t.Error("Expected non-empty text output")
	}

	t.Logf("Total Keys: %d", report.KeyStats.Total)
	t.Logf("Total Memory: %s", report.MemoryStats.TotalFmt)
	t.Logf("Text Output:\n%s", textOutput)
}

func TestStatsWithFilters(t *testing.T) {
	srcRdb := filepath.Join("../cases", "memory.rdb")
	report, err := Stats(srcRdb, 5, WithRegexOption("s"))
	if err != nil {
		t.Fatalf("Stats with regex filter failed: %v", err)
	}

	t.Logf("Filtered Total Keys: %d", report.KeyStats.Total)
}

func TestStatsTopN(t *testing.T) {
	srcRdb := filepath.Join("../cases", "memory.rdb")
	report, err := Stats(srcRdb, 3)
	if err != nil {
		t.Fatalf("Stats with topN failed: %v", err)
	}

	if len(report.TopLargestKeys) > 3 {
		t.Errorf("Expected at most 3 largest keys, got %d", len(report.TopLargestKeys))
	}
}

func TestStatsExpiration(t *testing.T) {
	srcRdb := filepath.Join("../cases", "expiration.rdb")
	report, err := Stats(srcRdb, 10)
	if err != nil {
		t.Fatalf("Stats for expiration case failed: %v", err)
	}

	t.Logf("Keys with TTL: %d", report.ExpireStats.WithTTL)
	t.Logf("Keys without TTL: %d", report.ExpireStats.WithoutTTL)
}

func TestStatsJSON(t *testing.T) {
	srcRdb := filepath.Join("../cases", "memory.rdb")
	report, err := Stats(srcRdb, 5)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	jsonOutput, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(jsonOutput) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	t.Logf("JSON Output:\n%s", string(jsonOutput))
}

func TestStatsFileNotFound(t *testing.T) {
	_, err := Stats("nonexistent.rdb", 10)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestStatsOutput(t *testing.T) {
	tmpFile := "/tmp/stats_test_output.txt"
	defer os.Remove(tmpFile)

	srcRdb := filepath.Join("../cases", "memory.rdb")
	report, err := Stats(srcRdb, 5)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	f, createErr := os.Create(tmpFile)
	if createErr != nil {
		t.Fatalf("Failed to create temp file: %v", createErr)
	}
	defer f.Close()

	_, writeErr := f.WriteString(report.ToText())
	if writeErr != nil {
		t.Fatalf("Failed to write to temp file: %v", writeErr)
	}
}
