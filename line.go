package main

import "time"

// Line is one tailed record.
type Line struct {
	File  string
	Text  string
	Level Level
	Time  time.Time
}
