package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// Level is a parsed log severity.
type Level int

const (
	LevelUnknown Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "LOG"
	}
}

func (l Level) JSON() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel reads a JSON "level" / "lvl" / "severity" field, then falls
// back to plain ERROR / WARN / INFO / DEBUG tokens.
func ParseLevel(line string) Level {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "{") {
		if lvl, ok := levelFromJSON(trimmed); ok {
			return lvl
		}
	}
	return levelFromTokens(trimmed)
}

func levelFromJSON(line string) (Level, bool) {
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		return LevelUnknown, false
	}
	for _, key := range []string{"level", "lvl", "severity"} {
		if v, ok := obj[key]; ok {
			return levelFromString(stringify(v)), true
		}
	}
	return LevelUnknown, false
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return strings.Trim(string(b), `"`)
	}
}

func levelFromString(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug", "dbg", "trace", "verbose":
		return LevelDebug
	case "info", "information", "notice", "inf":
		return LevelInfo
	case "warn", "warning", "wrn":
		return LevelWarn
	case "error", "err", "fatal", "panic", "crit", "critical", "alert", "emerg", "emergency":
		return LevelError
	default:
		return LevelUnknown
	}
}

func levelFromTokens(line string) Level {
	found := LevelUnknown
	for _, tok := range tokens(line) {
		switch levelFromString(tok) {
		case LevelError:
			return LevelError
		case LevelWarn:
			if found < LevelWarn {
				found = LevelWarn
			}
		case LevelInfo:
			if found < LevelInfo {
				found = LevelInfo
			}
		case LevelDebug:
			if found < LevelDebug {
				found = LevelDebug
			}
		}
	}
	return found
}

func parseLevelFilter(s string) (map[Level]bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	allowed := make(map[Level]bool)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		low := strings.ToLower(part)
		if low == "unknown" || low == "log" {
			allowed[LevelUnknown] = true
			continue
		}
		lvl := levelFromString(part)
		if lvl == LevelUnknown {
			return nil, fmt.Errorf("unknown --level %q", part)
		}
		allowed[lvl] = true
	}
	if len(allowed) == 0 {
		return nil, nil
	}
	return allowed, nil
}

func levelAllowed(filter map[Level]bool, l Level) bool {
	if filter == nil {
		return true
	}
	return filter[l]
}

func tokens(s string) []string {
	out := make([]string, 0, 8)
	start := -1
	for i, r := range s {
		if unicode.IsLetter(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			out = append(out, s[start:i])
			start = -1
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}
