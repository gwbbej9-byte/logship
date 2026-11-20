package main

import "sync"

// Ring is a fixed-capacity FIFO of tailed lines.
type Ring struct {
	mu   sync.Mutex
	buf  []Line
	head int
	len  int
	cap  int
}

func NewRing(capacity int) *Ring {
	if capacity < 1 {
		capacity = 1
	}
	return &Ring{
		buf: make([]Line, capacity),
		cap: capacity,
	}
}

func (r *Ring) Push(line Line) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.len < r.cap {
		r.buf[(r.head+r.len)%r.cap] = line
		r.len++
		return
	}
	r.buf[r.head] = line
	r.head = (r.head + 1) % r.cap
}

func (r *Ring) Last(n int) []Line {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n <= 0 || r.len == 0 {
		return nil
	}
	if n > r.len {
		n = r.len
	}
	out := make([]Line, n)
	start := r.head + r.len - n
	for i := 0; i < n; i++ {
		out[i] = r.buf[(start+i)%r.cap]
	}
	return out
}

func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.len
}
