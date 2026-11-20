package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "logship: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return fmt.Errorf("missing command")
	}
	switch args[0] {
	case "tail":
		return cmdTail(args[1:])
	case "-h", "-help", "--help", "help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func cmdTail(args []string) error {
	fs := flag.NewFlagSet("tail", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	listen := fs.String("listen", "", "HTTP listen address, e.g. :8080")
	jsonOut := fs.Bool("json-out", false, "emit JSON lines instead of ANSI color")
	levelFlag := fs.String("level", "", "comma-separated levels to emit (debug,info,warn,error); default all")
	fs.Usage = usage
	files, flagArgs, err := splitTailArgs(args)
	if err != nil {
		return err
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		files = append(files, extra...)
	}
	if len(files) == 0 {
		usage()
		return fmt.Errorf("tail requires at least one FILE")
	}
	filter, err := parseLevelFilter(*levelFlag)
	if err != nil {
		return err
	}
	for i, p := range files {
		abs, err := filepath.Abs(p)
		if err == nil {
			files[i] = abs
		}
	}

	tracker := NewTracker(files)
	ring := NewRing(1024)
	out := make(chan Line, 256)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var followWG sync.WaitGroup
	for _, path := range files {
		followWG.Add(1)
		go func(p string) {
			defer followWG.Done()
			if err := Follow(ctx, p, out, tracker); err != nil && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "logship: %s: %v\n", p, err)
			}
		}(path)
	}
	go func() {
		followWG.Wait()
		close(out)
	}()

	httpDone := make(chan struct{})
	if *listen != "" {
		go func() {
			defer close(httpDone)
			if err := serveHTTP(ctx, *listen, tracker, ring); err != nil && ctx.Err() == nil {
				fmt.Fprintf(os.Stderr, "logship: http %s: %v\n", *listen, err)
				cancel()
			}
		}()
		fmt.Fprintf(os.Stderr, "logship  --listen %s  GET /healthz  GET /lines?n=50\n", *listen)
	} else {
		close(httpDone)
	}

	color := !*jsonOut && stdoutTTY()
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	for {
		select {
		case <-ctx.Done():
			followWG.Wait()
			<-httpDone
			return nil
		case line, ok := <-out:
			if !ok {
				<-httpDone
				return nil
			}
			if !levelAllowed(filter, line.Level) {
				continue
			}
			ring.Push(line)
			if *jsonOut {
				_ = enc.Encode(map[string]any{
					"file":  line.File,
					"level": line.Level.JSON(),
					"line":  line.Text,
					"ts":    line.Time.UTC().Format(time.RFC3339Nano),
				})
				continue
			}
			fmt.Println(formatLine(line, color))
		}
	}
}

func splitTailArgs(args []string) (files, flags []string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			files = append(files, args[i+1:]...)
			break
		}
		if strings.HasPrefix(a, "-") {
			name, hasVal, val := cutFlag(a)
			switch name {
			case "--listen", "-listen", "--level", "-level":
				flags = append(flags, a)
				if !hasVal {
					need := "--listen"
					if name == "--level" || name == "-level" {
						need = "--level"
					}
					if i+1 >= len(args) {
						return nil, nil, fmt.Errorf("%s needs a value", need)
					}
					i++
					flags = append(flags, args[i])
				}
			case "--json-out", "-json-out":
				flags = append(flags, a)
			case "-h", "-help", "--help":
				flags = append(flags, a)
			default:
				if hasVal {
					_ = val
				}
				return nil, nil, fmt.Errorf("unknown flag %s", name)
			}
			continue
		}
		files = append(files, a)
	}
	return files, flags, nil
}

func cutFlag(a string) (name string, hasVal bool, val string) {
	if i := strings.IndexByte(a, '='); i >= 0 {
		return a[:i], true, a[i+1:]
	}
	return a, false, ""
}

func usage() {
	fmt.Fprintf(os.Stderr, `logship — concurrent log tailer

Usage:
  logship tail FILE [FILE...] [--listen ADDR] [--json-out] [--level LEVELS]

Flags:
  --listen ADDR   serve GET /healthz and GET /lines?n=50 on ADDR
  --json-out      emit JSON lines instead of ANSI color
  --level LEVELS  comma-separated: debug,info,warn,error (default: all)

Reads each file from the start, then follows like tail -F
(truncate + rotate). One goroutine per file.

Examples:
  logship tail samples/app.log samples/api.log
  logship tail samples/app.log samples/api.log --listen :8080
`)
}
