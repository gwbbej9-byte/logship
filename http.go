package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type lineDTO struct {
	File  string `json:"file"`
	Level string `json:"level"`
	Text  string `json:"text"`
	Time  string `json:"time"`
}

type linesResp struct {
	N     int       `json:"n"`
	Count int       `json:"count"`
	Lines []lineDTO `json:"lines"`
}

func newMux(tracker *Tracker, ring *Ring) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, tracker.Snapshot())
	})
	mux.HandleFunc("/lines", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		n := 50
		if s := r.URL.Query().Get("n"); s != "" {
			v, err := strconv.Atoi(s)
			if err != nil || v < 1 {
				http.Error(w, `{"error":"n must be a positive integer"}`, http.StatusBadRequest)
				return
			}
			n = v
		}
		if n > 500 {
			n = 500
		}
		got := ring.Last(n)
		resp := linesResp{N: n, Count: len(got), Lines: make([]lineDTO, 0, len(got))}
		for _, ln := range got {
			resp.Lines = append(resp.Lines, lineDTO{
				File:  ln.File,
				Level: ln.Level.JSON(),
				Text:  ln.Text,
				Time:  ln.Time.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(w, http.StatusOK, resp)
	})
	return mux
}

func newHTTPServer(addr string, tracker *Tracker, ring *Ring) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           newMux(tracker, ring),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func serveHTTP(ctx context.Context, addr string, tracker *Tracker, ring *Ring) error {
	srv := newHTTPServer(addr, tracker, ring)
	errc := make(chan error, 1)
	go func() {
		errc <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shctx)
		<-errc
		return ctx.Err()
	case err := <-errc:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
