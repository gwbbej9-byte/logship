package main

import (
	"sync"
	"time"
)

// FileStat is a snapshot of one tailed file.
type FileStat struct {
	Path         string     `json:"path"`
	BytesRead    int64      `json:"bytes_read"`
	LastLineTime *time.Time `json:"last_line_time"`
}

// Health is GET /healthz.
type Health struct {
	Files        []FileStat `json:"files"`
	BytesRead    int64      `json:"bytes_read"`
	LastLineTime *time.Time `json:"last_line_time"`
}

// Tracker aggregates per-file byte counts and last-line times.
type Tracker struct {
	mu    sync.Mutex
	order []string
	files map[string]*liveStat
}

type liveStat struct {
	path  string
	bytes int64
	last  time.Time
	ok    bool
}

func NewTracker(paths []string) *Tracker {
	t := &Tracker{files: make(map[string]*liveStat, len(paths))}
	for _, p := range paths {
		if _, ok := t.files[p]; ok {
			continue
		}
		t.order = append(t.order, p)
		t.files[p] = &liveStat{path: p}
	}
	return t
}

func (t *Tracker) AddBytes(path string, n int) {
	if n <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	st := t.ensure(path)
	st.bytes += int64(n)
}

func (t *Tracker) RecordLine(line Line) {
	t.mu.Lock()
	defer t.mu.Unlock()
	st := t.ensure(line.File)
	st.last = line.Time
	st.ok = true
}

func (t *Tracker) Snapshot() Health {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := Health{Files: make([]FileStat, 0, len(t.order))}
	var latest time.Time
	var have bool
	for _, p := range t.order {
		st := t.files[p]
		fs := FileStat{Path: st.path, BytesRead: st.bytes}
		if st.ok {
			ts := st.last.UTC()
			fs.LastLineTime = &ts
			if !have || ts.After(latest) {
				latest = ts
				have = true
			}
		}
		out.Files = append(out.Files, fs)
		out.BytesRead += st.bytes
	}
	if have {
		ts := latest
		out.LastLineTime = &ts
	}
	return out
}

func (t *Tracker) ensure(path string) *liveStat {
	st, ok := t.files[path]
	if ok {
		return st
	}
	st = &liveStat{path: path}
	t.files[path] = st
	t.order = append(t.order, path)
	return st
}
