package tsm

import (
	"sort"
	"sync"
	"time"

	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
)

const (
	maxTSMFileSize   = uint32(2 * 1024 * 1024 * 1024) // 2GB
	defaultChunkSize = uint64(maxTSMFileSize) * 5     // 10 GB chunks
)

// ChunkedCompactionPlanner implements CompactionPlanner using a strategy to roll up
// multiple generations of TSM files into larger files in stages.  It attempts
// to minimize the number of TSM files on disk while rolling up a bounded number
// of files.  Once a set of files hits a chunk threshold, they are excluded from further
// compactions.  This file was ported over
// from github.com/influxdata/influxdb/tsdb/engine/tsm1/compact.go
type ChunkedCompactionPlanner struct {
	FileStore fileStore

	// lastPlanCheck is the last time Plan was called
	lastPlanCheck time.Time

	mu sync.RWMutex
	// lastFindGenerations is the last time findGenerations was run
	lastFindGenerations time.Time

	// lastGenerations is the last set of generations found by findGenerations
	lastGenerations tsmGenerations

	// filesInUse is the set of files that have been returned as part of a plan and might
	// be being compacted.  Two plans should not return the same file at any given time.
	filesInUse map[string]struct{}
}

type fileStore interface {
	Stats() []tsm1.FileStat
	LastModified() time.Time
	BlockCount(path string, idx int) int
}

func NewChunkedCompactionPlanner(fs fileStore) *ChunkedCompactionPlanner {
	return &ChunkedCompactionPlanner{
		FileStore:  fs,
		filesInUse: make(map[string]struct{}),
	}
}

func (c *ChunkedCompactionPlanner) SetFileStore(fs *tsm1.FileStore) {
	c.FileStore = fs
}

// FullyCompacted returns true if the shard is fully compacted.
func (c *ChunkedCompactionPlanner) FullyCompacted() bool {
	gens := c.findGenerations(false)
	return len(gens) <= 1 && !gens.hasTombstones()
}

// ForceFull is a no-op
func (c *ChunkedCompactionPlanner) ForceFull() {
}

// PlanLevel returns a set of TSM files to rewrite for a specific level.
func (c *ChunkedCompactionPlanner) PlanLevel(level int) []tsm1.CompactionGroup {
	// Determine the generations from all files on disk.  We need to treat
	// a generation conceptually as a single file even though it may be
	// split across several files in sequence.
	generations := c.findGenerations(true)

	// If there is only one generation and no tombstones, then there's nothing to
	// do.
	if len(generations) <= 1 && !generations.hasTombstones() {
		return nil
	}

	// Group each generation by level such that two adjacent generations in the same
	// level become part of the same group.
	var currentGen tsmGenerations
	var groups []tsmGenerations
	for i := 0; i < len(generations); i++ {
		cur := generations[i]

		// See if this generation is orphan'd which would prevent it from being further
		// compacted until a final full compactin runs.
		if i < len(generations)-1 && cur.level() < generations[i+1].level() {
			currentGen = append(currentGen, cur)
			continue
		}

		if len(currentGen) == 0 || currentGen.level() == cur.level() {
			currentGen = append(currentGen, cur)
			continue
		}
		groups = append(groups, currentGen)

		currentGen = tsmGenerations{cur}
	}

	if len(currentGen) > 0 {
		groups = append(groups, currentGen)
	}

	// Remove any groups in the wrong level
	var levelGroups []tsmGenerations
	for _, cur := range groups {
		if cur.level() == level {
			levelGroups = append(levelGroups, cur)
		}
	}

	minGenerations := 4
	if level == 1 {
		minGenerations = 8
	}

	var cGroups []tsm1.CompactionGroup
	for _, group := range levelGroups {
		for _, chunk := range group.chunk(minGenerations) {
			var cGroup tsm1.CompactionGroup
			var hasTombstones bool
			for _, gen := range chunk {
				if gen.hasTombstones() {
					hasTombstones = true
				}
				for _, file := range gen.files {
					cGroup = append(cGroup, file.Path)
				}
			}

			if len(chunk) < minGenerations && !hasTombstones {
				continue
			}

			cGroups = append(cGroups, cGroup)
		}
	}

	if !c.acquire(cGroups) {
		return nil
	}

	return cGroups
}

// PlanOptimize returns an empty plan.
func (c *ChunkedCompactionPlanner) PlanOptimize() []tsm1.CompactionGroup {
	return nil
}

// Plan returns a set of TSM files to rewrite for level 4 or higher.  The planning returns
// multiple groups if possible to allow compactions to run concurrently.
func (c *ChunkedCompactionPlanner) Plan(lastWrite time.Time) []tsm1.CompactionGroup {
	generations := c.findGenerations(true)

	// don't plan if nothing has changed in the filestore
	if c.lastPlanCheck.After(c.FileStore.LastModified()) && !generations.hasTombstones() {
		return nil
	}

	c.lastPlanCheck = time.Now()

	// If there is only one generation, return early to avoid re-compacting the same file
	// over and over again.
	if len(generations) <= 1 && !generations.hasTombstones() {
		return nil
	}

	// Need to find the ending point for level 4 files.  They will be the oldest files. We scan
	// each generation in descending break once we see a file less than 4.
	var start, end int
	for i, g := range generations {
		if g.level() <= 3 {
			break
		}
		end = i + 1
	}

	// As compactions run, the oldest files get bigger.  We don't want to re-compact them during
	// this planning if they are maxed out so skip over any we see.
	for i, g := range generations[:end] {
		if g.hasTombstones() {
			break
		}

		// Skip the generation if it's over the max chunk size
		if g.size() >= defaultChunkSize {
			start = i + 1
		}
	}

	// step is how may files to compact in a group.  We want to clamp it at 4 but also stil
	// return groups smaller than 4.
	step := 4
	if step > end {
		step = end
	}

	// slice off the generations that we'll examine
	generations = generations[start:end]

	// Loop through the generations in groups of size step and see if we can compact all (or
	// some of them as group)
	groups := []tsmGenerations{}
	for i := 0; i < len(generations); i += step {
		var skipGroup bool
		startIndex := i

		for j := i; j < i+step && j < len(generations); j++ {
			gen := generations[j]
			lvl := gen.level()

			// Skip compacting this group if there happens to be any lower level files in the
			// middle.  These will get picked up by the level compactors.
			if lvl <= 3 {
				skipGroup = true
				break
			}

			// Skip the file if it's over the max size and it contains a full block
			if gen.size() >= defaultChunkSize && !gen.hasTombstones() {
				startIndex++
				continue
			}
		}

		if skipGroup {
			continue
		}

		endIndex := i + step
		if endIndex > len(generations) {
			endIndex = len(generations)
		}
		if endIndex-startIndex > 0 {
			groups = append(groups, generations[startIndex:endIndex])
		}
	}

	if len(groups) == 0 {
		return nil
	}

	// With the groups, we need to evaluate whether the group as a whole can be compacted
	compactable := []tsmGenerations{}
	for _, group := range groups {
		//if we don't have enough generations to compact, skip it
		if len(group) < 4 && !group.hasTombstones() {
			continue
		}
		compactable = append(compactable, group)
	}

	// All the files to be compacted must be compacted in order.  We need to convert each
	// group to the actual set of files in that group to be compacted.
	var tsmFiles []tsm1.CompactionGroup
	for _, c := range compactable {
		var cGroup tsm1.CompactionGroup
		for _, group := range c {
			for _, f := range group.files {
				cGroup = append(cGroup, f.Path)
			}
		}
		sort.Strings(cGroup)
		tsmFiles = append(tsmFiles, cGroup)
	}

	if !c.acquire(tsmFiles) {
		return nil
	}
	return tsmFiles
}

// findGenerations groups all the TSM files by generation based
// on their filename, then returns the generations in descending order (newest first).
// If skipInUse is true, tsm files that are part of an existing compaction plan
// are not returned.
func (c *ChunkedCompactionPlanner) findGenerations(skipInUse bool) tsmGenerations {
	c.mu.Lock()
	defer c.mu.Unlock()

	last := c.lastFindGenerations

	if !last.IsZero() && c.FileStore.LastModified().Equal(last) {
		return c.lastGenerations
	}

	genTime := c.FileStore.LastModified()
	tsmStats := c.FileStore.Stats()
	generations := make(map[int]*tsmGeneration, len(tsmStats))
	for _, f := range tsmStats {
		gen, _, _ := tsm1.DefaultParseFileName(f.Path)

		// Skip any files that are assigned to a current compaction plan
		if _, ok := c.filesInUse[f.Path]; skipInUse && ok {
			continue
		}

		group := generations[gen]
		if group == nil {
			group = &tsmGeneration{
				id: gen,
			}
			generations[gen] = group
		}
		group.files = append(group.files, f)
	}

	orderedGenerations := make(tsmGenerations, 0, len(generations))
	for _, g := range generations {
		orderedGenerations = append(orderedGenerations, g)
	}
	if !orderedGenerations.IsSorted() {
		sort.Sort(orderedGenerations)
	}

	c.lastFindGenerations = genTime
	c.lastGenerations = orderedGenerations

	return orderedGenerations
}

func (c *ChunkedCompactionPlanner) acquire(groups []tsm1.CompactionGroup) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// See if the new files are already in use
	for _, g := range groups {
		for _, f := range g {
			if _, ok := c.filesInUse[f]; ok {
				return false
			}
		}
	}

	// Mark all the new files in use
	for _, g := range groups {
		for _, f := range g {
			c.filesInUse[f] = struct{}{}
		}
	}
	return true
}

// Release removes the files reference in each compaction group allowing new plans
// to be able to use them.
func (c *ChunkedCompactionPlanner) Release(groups []tsm1.CompactionGroup) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, g := range groups {
		for _, f := range g {
			delete(c.filesInUse, f)
		}
	}
}

// tsmGeneration represents the TSM files within a generation.
// 000001-01.tsm, 000001-02.tsm would be in the same generation
// 000001 each with different sequence numbers.
type tsmGeneration struct {
	id    int
	files []tsm1.FileStat
}

// size returns the total size of the files in the generation.
func (t *tsmGeneration) size() uint64 {
	var n uint64
	for _, f := range t.files {
		n += uint64(f.Size)
	}
	return n
}

// compactionLevel returns the level of the files in this generation.
func (t *tsmGeneration) level() int {
	// Level 0 is always created from the result of a cache compaction.  It generates
	// 1 file with a sequence num of 1.  Level 2 is generated by compacting multiple
	// level 1 files.  Level 3 is generate by compacting multiple level 2 files.  Level
	// 4 is for anything else.
	if _, seq, _ := tsm1.DefaultParseFileName(t.files[0].Path); seq < 4 {
		return seq
	}

	return 4
}

// hasTombstones returns true if there are keys removed for any of the files.
func (t *tsmGeneration) hasTombstones() bool {
	for _, f := range t.files {
		if f.HasTombstone {
			return true
		}
	}
	return false
}

type tsmGenerations []*tsmGeneration

func (a tsmGenerations) Len() int           { return len(a) }
func (a tsmGenerations) Less(i, j int) bool { return a[i].id < a[j].id }
func (a tsmGenerations) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a tsmGenerations) hasTombstones() bool {
	for _, g := range a {
		if g.hasTombstones() {
			return true
		}
	}
	return false
}

func (a tsmGenerations) level() int {
	var level int
	for _, g := range a {
		lev := g.level()
		if lev > level {
			level = lev
		}
	}
	return level
}

func (a tsmGenerations) chunk(size int) []tsmGenerations {
	var chunks []tsmGenerations
	for len(a) > 0 {
		if len(a) >= size {
			chunks = append(chunks, a[:size])
			a = a[size:]
		} else {
			chunks = append(chunks, a)
			a = a[len(a):]
		}
	}
	return chunks
}

func (a tsmGenerations) IsSorted() bool {
	if len(a) == 1 {
		return true
	}

	for i := 1; i < len(a); i++ {
		if a.Less(i, i-1) {
			return false
		}
	}
	return true
}
