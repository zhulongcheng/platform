package tsm_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"github.com/influxdata/platform/storage/tsm"
)

func TestChunkedCompactionPlanner_Plan_Min(t *testing.T) {
	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return []tsm1.FileStat{
					{
						Path: "01-01.tsm1",
						Size: 1 * 1024 * 1024,
					},
					{
						Path: "02-01.tsm1",
						Size: 1 * 1024 * 1024,
					},
					{
						Path: "03-1.tsm1",
						Size: 251 * 1024 * 1024,
					},
				}
			},
		},
	)

	tsm := cp.Plan(time.Now())
	if exp, got := 0, len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

// Ensure that if there are older files that can be compacted together but a newer
// file that is in a larger step, the older ones will get compacted.
func TestChunkedCompactionPlanner_Plan_CombineSequence(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-04.tsm1",
			Size: 128 * 1024 * 1024,
		},
		{
			Path: "02-04.tsm1",
			Size: 128 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 128 * 1024 * 1024,
		},
		{
			Path: "04-04.tsm1",
			Size: 128 * 1024 * 1024,
		},
		{
			Path: "06-02.tsm1",
			Size: 67 * 1024 * 1024,
		},
		{
			Path: "07-02.tsm1",
			Size: 128 * 1024 * 1024,
		},
		{
			Path: "08-01.tsm1",
			Size: 251 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{data[0], data[1], data[2], data[3]}
	tsm := cp.Plan(time.Now())
	if exp, got := len(expFiles), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

// Ensure that the planner grabs the smallest compaction step
func TestChunkedCompactionPlanner_Plan_MultipleGroups(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "02-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "04-04.tsm1",
			Size: 129 * 1024 * 1024,
		},
		{
			Path: "05-04.tsm1",
			Size: 129 * 1024 * 1024,
		},
		{
			Path: "06-04.tsm1",
			Size: 129 * 1024 * 1024,
		},
		{
			Path: "07-04.tsm1",
			Size: 129 * 1024 * 1024,
		},
		{
			Path: "08-04.tsm1",
			Size: 129 * 1024 * 1024,
		},
		{
			Path: "09-04.tsm1", // should be skipped
			Size: 129 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(&fakeFileStore{
		PathsFn: func() []tsm1.FileStat {
			return data
		},
	})

	expFiles := []tsm1.FileStat{data[0], data[1], data[2], data[3],
		data[4], data[5], data[6], data[7]}
	tsm := cp.Plan(time.Now())

	if got, exp := len(tsm), 2; got != exp {
		t.Fatalf("compaction group length mismatch: got %v, exp %v", got, exp)
	}

	if exp, got := len(expFiles[:4]), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	if exp, got := len(expFiles[4:]), len(tsm[1]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles[:4] {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}

	for i, p := range expFiles[4:] {
		if got, exp := tsm[1][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

// Ensure that the planner grabs the smallest compaction step
func TestChunkedCompactionPlanner_Plan_SkipChunks(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "01-05.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "01-06.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "01-07.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "01-08.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "02-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "04-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "05-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(&fakeFileStore{
		PathsFn: func() []tsm1.FileStat {
			return data
		},
	})

	expFiles := []tsm1.FileStat{data[5], data[6], data[7], data[8]}
	tsm := cp.Plan(time.Now())

	if got, exp := len(tsm), 1; got != exp {
		t.Fatalf("compaction group length mismatch: got %v, exp %v", got, exp)
	}

	if exp, got := len(expFiles), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

// Ensure that the planner grabs the smallest compaction step
func TestChunkedCompactionPlanner_PlanLevel_SmallestCompactionStep(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-03.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "04-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "05-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "07-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "08-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "09-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "10-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "11-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "12-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{data[4], data[5], data[6], data[7], data[8], data[9], data[10], data[11]}
	tsm := cp.PlanLevel(1)
	if exp, got := len(expFiles), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

func TestChunkedCompactionPlanner_PlanLevel_SplitFile(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-03.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-03.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "04-03.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "05-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{data[0], data[1], data[2], data[3], data[4]}
	tsm := cp.PlanLevel(3)
	if exp, got := len(expFiles), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

func TestChunkedCompactionPlanner_PlanLevel_IsolatedHighLevel(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-02.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-02.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-03.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "04-02.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "05-02.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-02.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		})

	expFiles := []tsm1.FileStat{}
	tsm := cp.PlanLevel(3)
	if exp, got := len(expFiles), len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

func TestChunkedCompactionPlanner_PlanLevel3_MinFiles(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-03.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-01.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "04-01.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "05-02.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{}
	tsm := cp.PlanLevel(3)
	if exp, got := len(expFiles), len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

func TestChunkedCompactionPlanner_PlanLevel2_MinFiles(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "02-04.tsm1",
			Size: 251 * 1024 * 1024,
		},

		{
			Path: "03-02.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "03-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{}
	tsm := cp.PlanLevel(2)
	if exp, got := len(expFiles), len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

func TestChunkedCompactionPlanner_PlanLevel_Tombstone(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path:         "01-03.tsm1",
			Size:         251 * 1024 * 1024,
			HasTombstone: true,
		},
		{
			Path: "02-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-01.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "04-01.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "05-02.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{data[0], data[1]}
	tsm := cp.PlanLevel(3)
	if exp, got := len(expFiles), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

func TestChunkedCompactionPlanner_PlanLevel_Multiple(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-01.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-01.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "04-01.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "05-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "07-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "08-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles1 := []tsm1.FileStat{data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7]}

	tsm := cp.PlanLevel(1)
	if exp, got := len(expFiles1), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles1 {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

func TestChunkedCompactionPlanner_PlanLevel_InUse(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-01.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-01.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "04-01.tsm1",
			Size: 10 * 1024 * 1024,
		},
		{
			Path: "05-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "06-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "07-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "08-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "09-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "10-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "11-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "12-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "13-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "14-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "15-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "16-01.tsm1",
			Size: 1 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles1 := data[0:8]
	expFiles2 := data[8:16]

	tsm := cp.PlanLevel(1)
	if exp, got := len(expFiles1), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles1 {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}

	if exp, got := len(expFiles2), len(tsm[1]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles2 {
		if got, exp := tsm[1][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}

	cp.Release(tsm[1:])

	tsm = cp.PlanLevel(1)
	if exp, got := len(expFiles2), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles2 {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

func TestChunkedCompactionPlanner_PlanOptimize_NoLevel4(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-03.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-03.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-03.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{}
	tsm := cp.PlanOptimize()
	if exp, got := len(expFiles), len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

func TestChunkedCompactionPlanner_PlanOptimize_Empty(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-04.tsm1",
			Size: 251 * 1024 * 1024,
		},
		{
			Path: "02-04.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "04-04.tsm1",
			Size: 1 * 1024 * 1024,
		},
		{
			Path: "05-03.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "06-04.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
		{
			Path: "07-03.tsm1",
			Size: 2 * 1024 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	tsm := cp.PlanOptimize()
	if exp, got := 0, len(tsm); exp != got {
		t.Fatalf("group length mismatch: got %v, exp %v", got, exp)
	}
}

// Ensure that the planner will not return files that are over the max
// allowable size
func TestChunkedCompactionPlanner_Plan_SkipMaxSizeFiles(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-01.tsm1",
			Size: 2049 * 1024 * 1024,
		},
		{
			Path: "02-02.tsm1",
			Size: 2049 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	tsm := cp.Plan(time.Now())
	if exp, got := 0, len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

// Ensure that the planner will compact files that are past the smallest step
// size even if there is a single file in the smaller step size
func TestChunkedCompactionPlanner_Plan_CompactsMiddleSteps(t *testing.T) {
	data := []tsm1.FileStat{
		{
			Path: "01-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "02-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "03-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "04-04.tsm1",
			Size: 64 * 1024 * 1024,
		},
		{
			Path: "05-02.tsm1",
			Size: 2 * 1024 * 1024,
		},
	}

	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return data
			},
		},
	)

	expFiles := []tsm1.FileStat{data[0], data[1], data[2], data[3]}
	tsm := cp.Plan(time.Now())
	if exp, got := len(expFiles), len(tsm[0]); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}

	for i, p := range expFiles {
		if got, exp := tsm[0][i], p.Path; got != exp {
			t.Fatalf("tsm file mismatch: got %v, exp %v", got, exp)
		}
	}
}

func TestChunkedCompactionPlanner_Plan_LargeGeneration(t *testing.T) {
	cp := tsm.NewChunkedCompactionPlanner(
		&fakeFileStore{
			PathsFn: func() []tsm1.FileStat {
				return []tsm1.FileStat{
					{
						Path: "000000278-000000006.tsm",
						Size: 2148340232,
					},
					{
						Path: "000000278-000000007.tsm",
						Size: 2148356556,
					},
					{
						Path: "000000278-000000008.tsm",
						Size: 167780181,
					},
					{
						Path: "000000278-000047040.tsm",
						Size: 2148728539,
					},
					{
						Path: "000000278-000047041.tsm",
						Size: 701863692,
					},
				}
			},
		},
	)

	tsm := cp.Plan(time.Now())
	if exp, got := 0, len(tsm); got != exp {
		t.Fatalf("tsm file length mismatch: got %v, exp %v", got, exp)
	}
}

func MustOpenTSMReader(name string) *tsm1.TSMReader {
	f, err := os.Open(name)
	if err != nil {
		panic(fmt.Sprintf("open file: %v", err))
	}

	r, err := tsm1.NewTSMReader(f)
	if err != nil {
		panic(fmt.Sprintf("new reader: %v", err))
	}
	return r
}

type fakeFileStore struct {
	PathsFn      func() []tsm1.FileStat
	lastModified time.Time
	blockCount   int
	readers      []*tsm1.TSMReader
}

func (w *fakeFileStore) Stats() []tsm1.FileStat {
	return w.PathsFn()
}

func (w *fakeFileStore) NextGeneration() int {
	return 1
}

func (w *fakeFileStore) LastModified() time.Time {
	return w.lastModified
}

func (w *fakeFileStore) BlockCount(path string, idx int) int {
	return w.blockCount
}

func (w *fakeFileStore) TSMReader(path string) *tsm1.TSMReader {
	r := MustOpenTSMReader(path)
	w.readers = append(w.readers, r)
	return r
}

func (w *fakeFileStore) Close() {
	for _, r := range w.readers {
		r.Close()
	}
	w.readers = nil
}
