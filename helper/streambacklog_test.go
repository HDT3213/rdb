package helper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdt3213/rdb/core"
	"github.com/hdt3213/rdb/model"
)

const streamBacklogHeader = "database,key,group,consumer,pending,oldest_delivery_time,max_delivery_count,undelivered,dangling"

func runStreamBacklogReport(t *testing.T, srcRdb string, options ...interface{}) []string {
	t.Helper()
	outputPath := filepath.Join(t.TempDir(), "backlog.csv")
	output, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("create output file failed: %v", err)
	}
	err = StreamBacklogReport(srcRdb, output, options...)
	if closeErr := output.Close(); closeErr != nil {
		t.Fatalf("close output failed: %v", closeErr)
	}
	if err != nil {
		t.Fatalf("StreamBacklogReport failed: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}
	return strings.Split(strings.TrimRight(string(content), "\n"), "\n")
}

func assertLines(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("report =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestStreamBacklogReport(t *testing.T) {
	got := runStreamBacklogReport(t, filepath.Join("../cases", "stream_listoacks_3.rdb"))
	assertLines(t, got, []string{
		streamBacklogHeader,
		"0,mystream,consumer-group-name,consumer-name,1,1704557998397,1,0,0",
	})

	got = runStreamBacklogReport(t, filepath.Join("../cases", "stream_listpacks_1.rdb"))
	assertLines(t, got, []string{
		streamBacklogHeader,
		"0,listpack,g1,c1,2,1528516636879,1,144,0",
		"0,listpack,g1,c2,2,1528516649782,1,144,0",
		"0,listpack,g2,c1,1,1528516695691,1,83,0",
		"0,listpack,g3,c1,2,1528516699993,1,81,0",
		"0,listpack,g4,,0,0,0,0,0",
	})

	got = runStreamBacklogReport(t, filepath.Join("../cases", "stream_listpacks_2.rdb"))
	assertLines(t, got, []string{streamBacklogHeader})
}

func TestStreamBacklogReportWithRegex(t *testing.T) {
	srcRdb := filepath.Join("../cases", "stream_listpacks_1.rdb")
	got := runStreamBacklogReport(t, srcRdb, WithRegexOption("^listpack$"))
	if len(got) != 6 {
		t.Errorf("got %d lines, want the header and five rows", len(got))
	}
	got = runStreamBacklogReport(t, srcRdb, WithRegexOption("^no-such-key$"))
	assertLines(t, got, []string{streamBacklogHeader})
}

func TestStreamBacklogReportCsvQuotingAndOrder(t *testing.T) {
	newStream := func(key, group string, consumers ...*model.StreamConsumer) *model.StreamObject {
		return &model.StreamObject{
			BaseObject: &model.BaseObject{Key: key},
			Version:    3,
			Length:     2,
			LastId:     &model.StreamId{Ms: 2},
			Entries: []*model.StreamEntry{{
				FirstMsgId: &model.StreamId{Ms: 1},
				Fields:     []string{"k"},
				Msgs: []*model.StreamMessage{
					{Id: &model.StreamId{Ms: 1}, Fields: map[string]string{"k": "v"}},
					{Id: &model.StreamId{Ms: 2}, Fields: map[string]string{"k": "v"}},
				},
			}},
			Groups: []*model.StreamGroup{{
				Name:      group,
				LastId:    &model.StreamId{Ms: 1},
				Pending:   []*model.StreamNAck{{Id: &model.StreamId{Ms: 1}, DeliveryTime: 111, DeliveryCount: 2}},
				Consumers: consumers,
			}},
		}
	}
	pending := []*model.StreamId{{Ms: 1}}
	quoted := newStream("b,stream", "g,1",
		&model.StreamConsumer{Name: `c"1`, Pending: pending},
		&model.StreamConsumer{Name: "c2", Pending: pending},
	)
	plain := newStream("zeta", "zg", &model.StreamConsumer{Name: "zc", Pending: pending})

	srcRdb := filepath.Join(t.TempDir(), "streams.rdb")
	file, err := os.Create(srcRdb)
	if err != nil {
		t.Fatalf("create rdb failed: %v", err)
	}
	enc := core.NewEncoder(file)
	steps := []func() error{
		enc.WriteHeader,
		func() error { return enc.WriteDBHeader(0, 2, 0) },
		func() error { return enc.WriteStringObject("plain", []byte("v")) },
		func() error { return enc.WriteStreamObject(quoted.Key, quoted) },
		func() error { return enc.WriteDBHeader(3, 1, 0) },
		func() error { return enc.WriteStreamObject(plain.Key, plain) },
		enc.WriteEnd,
	}
	for i, step := range steps {
		if err := step(); err != nil {
			t.Fatalf("write rdb step %d failed: %v", i, err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close rdb failed: %v", err)
	}

	got := runStreamBacklogReport(t, srcRdb)
	assertLines(t, got, []string{
		streamBacklogHeader,
		`0,"b,stream","g,1","c""1",1,111,2,1,0`,
		`0,"b,stream","g,1",c2,1,111,2,1,0`,
		`3,zeta,zg,zc,1,111,2,1,0`,
	})
}

func TestStreamBacklogReportErrors(t *testing.T) {
	err := StreamBacklogReport("", os.Stdout)
	if err == nil || err.Error() != "src file path is required" {
		t.Errorf("empty src: got %v", err)
	}
	if err = StreamBacklogReport(filepath.Join(t.TempDir(), "missing.rdb"), os.Stdout); err == nil {
		t.Error("missing file: want an error")
	}
}
