package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"
)

const followPoll = 150 * time.Millisecond

// Follow tails path from offset 0 (existing content, then new writes).
// Handles truncation (size shrinks) and rotation (file identity changes).
// Retries if the file is missing. Returns when ctx is done.
func Follow(ctx context.Context, path string, out chan<- Line, tracker *Tracker) error {
	var (
		f       *os.File
		offset  int64
		ident   os.FileInfo
		buf     []byte
		scratch = make([]byte, 4096)
	)
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	reopen := func() error {
		if f != nil {
			_ = f.Close()
			f = nil
		}
		nf, err := openFollow(path)
		if err != nil {
			return err
		}
		if _, err := nf.Seek(offset, io.SeekStart); err != nil {
			_ = nf.Close()
			return err
		}
		info, err := nf.Stat()
		if err != nil {
			_ = nf.Close()
			return err
		}
		f = nf
		ident = info
		return nil
	}

	emit := func(raw []byte) {
		text := string(bytes.TrimRight(raw, "\r"))
		if text == "" {
			return
		}
		line := Line{
			File:  path,
			Text:  text,
			Level: ParseLevel(text),
			Time:  time.Now(),
		}
		if tracker != nil {
			tracker.RecordLine(line)
		}
		select {
		case <-ctx.Done():
		case out <- line:
		}
	}

	consume := func() error {
		n, rerr := f.Read(scratch)
		if n > 0 {
			offset += int64(n)
			if tracker != nil {
				tracker.AddBytes(path, n)
			}
			buf = append(buf, scratch[:n]...)
			for {
				i := bytes.IndexByte(buf, '\n')
				if i < 0 {
					break
				}
				emit(buf[:i])
				buf = buf[i+1:]
			}
		}
		return rerr
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		st, err := os.Stat(path)
		if err != nil {
			if f != nil {
				_ = f.Close()
				f = nil
				ident = nil
			}
			if sleep(ctx, followPoll) != nil {
				return ctx.Err()
			}
			continue
		}

		if f == nil {
			offset = 0
			buf = buf[:0]
			if err := reopen(); err != nil {
				if sleep(ctx, followPoll) != nil {
					return ctx.Err()
				}
				continue
			}
		} else if ident != nil && !os.SameFile(ident, st) {
			for {
				rerr := consume()
				if rerr == io.EOF || rerr != nil {
					break
				}
			}
			offset = 0
			buf = buf[:0]
			if err := reopen(); err != nil {
				if sleep(ctx, followPoll) != nil {
					return ctx.Err()
				}
				continue
			}
		} else if st.Size() < offset {
			offset = 0
			buf = buf[:0]
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				if err := reopen(); err != nil {
					if sleep(ctx, followPoll) != nil {
						return ctx.Err()
					}
					continue
				}
			}
		}

		rerr := consume()
		if rerr == io.EOF {
			if sleep(ctx, followPoll) != nil {
				return ctx.Err()
			}
			continue
		}
		if rerr != nil {
			_ = f.Close()
			f = nil
			ident = nil
			if sleep(ctx, followPoll) != nil {
				return ctx.Err()
			}
		}
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
