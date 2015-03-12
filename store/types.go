package store

import (
	"time"
)

type Measure struct {
	Name     string
	Value    int
	Baseline int
}

type CommitMeasure struct {
	CommitHash string
	Timestamp  time.Time
	Committer  string
	Measures   []Measure
}

type Exclusion struct {
	Committer string
	Excuse    string
	Measure   []string
}

type ByName []Measure

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (cm *CommitMeasure) String() string {
	return cm.CommitHash
}
