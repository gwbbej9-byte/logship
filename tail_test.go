package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func waitLine(t *testing.T, ch <-chan Line, sub string, d time.Duration) Line {
	t.Helper()
	deadline := time.NewTimer(d)
	defer deadline.Stop()
	for {
		select {
		case line := <-ch:
			if strings.Contains(line.Text, sub) {
				return line
			}
		case <-deadline.C:
			t.Fatalf("timeout waiting for %q", sub)
		}
	}
	panic("unreachable")
}

func TestFollowCancelStops(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan Line, 8)
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, path, ch, nil)
	}()
	cancel()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("Follow err=%v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Follow did not stop after cancel")
	}
}

func TestFollowSeesAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := make(chan Line, 8)
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, path, ch, nil)
	}()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("hello-from-test\n"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	waitLine(t, ch, "hello-from-test", 2*time.Second)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Follow did not return after cancel")
	}
}

func TestFollowExistingThenAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("seed-line\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ch := make(chan Line, 8)
	go func() { _ = Follow(ctx, path, ch, nil) }()

	waitLine(t, ch, "seed-line", 2*time.Second)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("second-line\n"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()

	waitLine(t, ch, "second-line", 2*time.Second)
	cancel()
}

func TestFollowTruncate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("before-trunc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ch := make(chan Line, 16)
	go func() { _ = Follow(ctx, path, ch, nil) }()
	waitLine(t, ch, "before-trunc", 2*time.Second)

	if err := os.Truncate(path, 0); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("after-trunc\n"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	_ = f.Close()
	waitLine(t, ch, "after-trunc", 2*time.Second)
	cancel()
}

func TestFollowRotate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	if err := os.WriteFile(path, []byte("on-old-inode\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ch := make(chan Line, 16)
	go func() { _ = Follow(ctx, path, ch, nil) }()
	waitLine(t, ch, "on-old-inode", 2*time.Second)

	rotated := filepath.Join(dir, "app.log.1")
	if err := os.Rename(path, rotated); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("on-new-inode\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitLine(t, ch, "on-new-inode", 2*time.Second)
	cancel()
}
