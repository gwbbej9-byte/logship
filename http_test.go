package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadHeaderTimeout(t *testing.T) {
	srv := newHTTPServer(":0", NewTracker(nil), NewRing(8))
	if srv.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout=%v want 5s", srv.ReadHeaderTimeout)
	}
}

func TestLinesBadN(t *testing.T) {
	h := newMux(NewTracker(nil), NewRing(8))
	for _, q := range []string{"n=abc", "n=0", "n=-1", "n=1.5"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/lines?"+q, nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status=%d want 400", q, rec.Code)
		}
	}
}

func TestLinesDefaultOK(t *testing.T) {
	ring := NewRing(8)
	ring.Push(Line{File: "a.log", Text: "INFO ok", Level: LevelInfo, Time: time.Now()})
	h := newMux(NewTracker(nil), ring)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/lines", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var resp linesResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.N != 50 || resp.Count != 1 {
		t.Fatalf("n=%d count=%d", resp.N, resp.Count)
	}
}

func TestHealthzLiveFields(t *testing.T) {
	path := "/tmp/a.log"
	tr := NewTracker([]string{path})
	tr.AddBytes(path, 12)
	now := time.Now().UTC().Truncate(time.Millisecond)
	tr.RecordLine(Line{File: path, Time: now})
	h := newMux(tr, NewRing(8))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body Health
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.BytesRead != 12 {
		t.Fatalf("bytes_read=%d", body.BytesRead)
	}
	if body.LastLineTime == nil {
		t.Fatal("last_line_time missing")
	}
	if len(body.Files) != 1 {
		t.Fatalf("files=%d", len(body.Files))
	}
	if body.Files[0].Path != path || body.Files[0].BytesRead != 12 {
		t.Fatalf("file=%+v", body.Files[0])
	}
	if body.Files[0].LastLineTime == nil {
		t.Fatal("file last_line_time missing")
	}
}
