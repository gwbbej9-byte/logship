package main

import (
	"strconv"
	"sync"
	"testing"
)

func TestRingLastWrap(t *testing.T) {
	r := NewRing(3)
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		r.Push(Line{Text: s, File: "f"})
	}
	got := r.Last(3)
	if len(got) != 3 {
		t.Fatalf("len=%d want 3", len(got))
	}
	want := []string{"c", "d", "e"}
	for i, s := range want {
		if got[i].Text != s {
			t.Errorf("Last[%d]=%q want %q", i, got[i].Text, s)
		}
	}
	if r.Len() != 3 {
		t.Fatalf("Len=%d want 3", r.Len())
	}
}

func TestRingLastN(t *testing.T) {
	r := NewRing(10)
	if r.Last(5) != nil {
		t.Fatal("empty Last should be nil")
	}
	r.Push(Line{Text: "only"})
	got := r.Last(50)
	if len(got) != 1 || got[0].Text != "only" {
		t.Fatalf("Last(50)=%v", got)
	}
	if r.Last(0) != nil {
		t.Fatal("Last(0) should be nil")
	}
}

func TestRingConcurrentPushLast(t *testing.T) {
	r := NewRing(64)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				r.Push(Line{Text: strconv.Itoa(id*1000 + j), File: "f"})
				_ = r.Last(10)
			}
		}(i)
	}
	wg.Wait()
	if r.Len() != 64 {
		t.Fatalf("Len=%d want 64", r.Len())
	}
}

func TestRingCapAtLeastOne(t *testing.T) {
	r := NewRing(0)
	r.Push(Line{Text: "x"})
	got := r.Last(1)
	if len(got) != 1 || got[0].Text != "x" {
		t.Fatalf("cap0: %+v", got)
	}
}
