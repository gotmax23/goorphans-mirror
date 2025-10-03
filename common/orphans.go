package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	OrphansJSON    = "orphans.json"
	OrphansTXT     = "orphans.txt"
	OrphanUID      = "orphan"
	OrphansBaseURL = "https://a.gtmx.me/orphans/"
)
const week time.Duration = time.Hour * 24 * 7

type GolangExemption int

const (
	GolangExemptionMust GolangExemption = iota
	GolangExemptionOptional
	GolangExemptionIgnore
	GolangExemptionOnly
)

var ToGolangExemption = map[string]GolangExemption{
	"must":     GolangExemptionMust,
	"optional": GolangExemptionOptional,
	"ignore":   GolangExemptionIgnore,
	"only":     GolangExemptionOnly,
}

var FromGolangExemption = map[GolangExemption]string{
	GolangExemptionMust:     "must",
	GolangExemptionOptional: "optional",
	GolangExemptionIgnore:   "ignore",
	GolangExemptionOnly:     "only",
}

func (ge GolangExemption) String() string {
	return FromGolangExemption[ge]
}

func (ge GolangExemption) MarshalText() ([]byte, error) {
	return []byte(ge.String()), nil
}

func (ge *GolangExemption) UnmarshalText(text []byte) error {
	s := string(text)
	var ok bool
	*ge, ok = ToGolangExemption[s]
	if !ok {
		return fmt.Errorf("invalid GolangExemption value: %q", s)
	}
	return nil
}

// Weeks returns weeks represented as a [time.Duration]
func Weeks(weeks int) time.Duration {
	return week * time.Duration(weeks)
}

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

type OrphanedFilterOptions struct {
	Duration        time.Duration
	GolangExemption GolangExemption
}

func (o *Orphans) OrphanedFilter(options OrphanedFilterOptions) (r []string, err error) {
	if options.GolangExemption == GolangExemptionMust && len(o.GolangExemptions) == 0 {
		return r, fmt.Errorf("GolangExemptionMust but no exemptions were listed")
	}
	exemptions := mapset.NewThreadUnsafeSet(o.GolangExemptions...)
	for _, p := range o.Orphans {
		if options.Duration != 0 {
			t := o.StatusChange[p]
			elapsed := time.Since(t)
			if elapsed < options.Duration {
				continue
			}
		}
		if exemptions.Contains(p) {
			if options.GolangExemption == GolangExemptionMust ||
				options.GolangExemption == GolangExemptionOptional {
				continue
			}
		} else if options.GolangExemption == GolangExemptionOnly {
			continue
		}
		r = append(r, p)
	}
	return r, nil
}
