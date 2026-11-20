package main

import (
	"sync"
	"testing"
	"time"
)

func TestTrackerConcurrentAppendSnapshot(t *testing.T) {
	paths := []string{"a.log", "b.log"}
	tr := NewTracker(paths)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := paths[i%2]
			for j := 0; j < 50; j++ {
				tr.AddBytes(p, 3)
				tr.RecordLine(Line{File: p, Time: time.Now()})
				_ = tr.Snapshot()
			}
		}(i)
	}
	wg.Wait()
	h := tr.Snapshot()
	if h.BytesRead != 8*50*3 {
		t.Fatalf("bytes_read=%d", h.BytesRead)
	}
	if h.LastLineTime == nil {
		t.Fatal("last_line_time nil")
	}
	if len(h.Files) != 2 {
		t.Fatalf("files=%d", len(h.Files))
	}
}
