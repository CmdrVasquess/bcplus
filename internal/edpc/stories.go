package edpc

import (
	"time"
)

type Data struct {
	Stories []Story
}

type Hint struct {
	File  string
	Found time.Time
}

type Story struct {
	Id     int
	Title  string
	Author string
	Hints  []Hint
}
