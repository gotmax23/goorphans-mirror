package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	OrphansJSON = "orphans.json"
	OrphansTXT  = "orphans.txt"
	OrphanUID   = "orphan"
)

type Orphans struct {
	Addresses                   []string             `json:"addresses"`
	AffectedPeople              map[string][]string  `json:"affected_people"`
	AllAffectedPeople           map[string][]string  `json:"all_affected_people"`
	FtbfsBreakingDeps           []string             `json:"ftbfs_breaking_deps"`
	FtbfsNotBreakingDeps        []string             `json:"ftbfs_not_breaking_deps"`
	GolangExemptions            []string             `json:"golang_exemptions,omitempty"`
	Orphans                     []string             `json:"orphans"`
	OrphansBreakingDeps         []string             `json:"orphans_breaking_deps"`
	OrphansBreakingDepsStale    []string             `json:"orphans_breaking_deps_stale"`
	OrphansNotBreakingDeps      []string             `json:"orphans_not_breaking_deps"`
	OrphansNotBreakingDepsStale []string             `json:"orphans_not_breaking_deps_stale"`
	StatusChange                map[string]time.Time `json:"status_change"`
	// NOTE(gotmax23): started_at and finished_at are extensions I added that
	// have not yet been committed to the releng repository.
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func LoadOrphans(path string) (*Orphans, error) {
	var orphans Orphans
	file, err := os.Open(path)
	if err != nil {
		return &orphans, fmt.Errorf("failed to load orphans.json: %v", err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&orphans)
	if err != nil {
		return &orphans, fmt.Errorf("failed to load orphans.json: %v", err)
	}
	return &orphans, nil
}
