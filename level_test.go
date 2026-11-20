package main

import "testing"

func TestParseLevelJSON(t *testing.T) {
	cases := []struct {
		in   string
		want Level
	}{
		{`{"level":"error","msg":"boom"}`, LevelError},
		{`{"level":"ERROR"}`, LevelError},
		{`{"level":"fatal"}`, LevelError},
		{`{"lvl":"warn"}`, LevelWarn},
		{`{"severity":"warning"}`, LevelWarn},
		{`{"level":"info","msg":"ok"}`, LevelInfo},
		{`{"level":"INFO"}`, LevelInfo},
		{`{"level":"debug"}`, LevelDebug},
		{`{"level":"trace"}`, LevelDebug},
		{`{"msg":"no level here"}`, LevelUnknown},
		{`{not json`, LevelUnknown},
	}
	for _, tc := range cases {
		got := ParseLevel(tc.in)
		if got != tc.want {
			t.Errorf("ParseLevel(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestParseLevelTokens(t *testing.T) {
	cases := []struct {
		in   string
		want Level
	}{
		{"2026-11-17 ERROR failed to flush wal", LevelError},
		{"[ERR] upstream 502", LevelError},
		{"FATAL panic in worker", LevelError},
		{"2026-11-17 WARN retry queue depth=41", LevelWarn},
		{"WARNING disk 87%", LevelWarn},
		{"INFO started worker pool", LevelInfo},
		{"NOTICE loaded config", LevelInfo},
		{"DEBUG binding listener", LevelDebug},
		{"TRACE gc cycle", LevelDebug},
		{"hello world", LevelUnknown},
		{"errorcode=12 without token", LevelUnknown},
		{"  mixed INFO then ERROR later  ", LevelError},
	}
	for _, tc := range cases {
		got := ParseLevel(tc.in)
		if got != tc.want {
			t.Errorf("ParseLevel(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestParseLevelFilter(t *testing.T) {
	all, err := parseLevelFilter("")
	if err != nil || all != nil {
		t.Fatalf("empty: %v %v", all, err)
	}
	if !levelAllowed(all, LevelError) || !levelAllowed(all, LevelDebug) {
		t.Fatal("empty filter must allow all")
	}
	f, err := parseLevelFilter("error,WARN")
	if err != nil {
		t.Fatal(err)
	}
	if !levelAllowed(f, LevelError) || !levelAllowed(f, LevelWarn) {
		t.Fatal("error,warn should pass")
	}
	if levelAllowed(f, LevelInfo) || levelAllowed(f, LevelDebug) {
		t.Fatal("info/debug should be filtered")
	}
	if _, err := parseLevelFilter("nope"); err == nil {
		t.Fatal("want error for unknown level")
	}
}

func TestLevelStringJSON(t *testing.T) {
	if LevelError.String() != "ERROR" {
		t.Fatalf("String: %s", LevelError.String())
	}
	if LevelError.JSON() != "error" {
		t.Fatalf("JSON: %s", LevelError.JSON())
	}
	if LevelUnknown.String() != "LOG" {
		t.Fatalf("unknown String: %s", LevelUnknown.String())
	}
}
