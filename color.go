package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ansiReset  = "\033[0m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiBold   = "\033[1m"
)

func stdoutTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func levelANSI(l Level) string {
	switch l {
	case LevelError:
		return ansiBold + ansiRed
	case LevelWarn:
		return ansiBold + ansiYellow
	case LevelInfo:
		return ansiGreen
	case LevelDebug:
		return ansiDim
	default:
		return ""
	}
}

func formatLine(line Line, color bool) string {
	tag := filepath.Base(line.File)
	lvl := fmt.Sprintf("%-5s", line.Level.String())
	if !color {
		return fmt.Sprintf("[%s] %s %s", tag, lvl, line.Text)
	}
	return fmt.Sprintf("%s[%s]%s %s%s%s %s%s%s",
		ansiCyan, tag, ansiReset,
		levelANSI(line.Level), lvl, ansiReset,
		levelANSI(line.Level), line.Text, ansiReset,
	)
}
