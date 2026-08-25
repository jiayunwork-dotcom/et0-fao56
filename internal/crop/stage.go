package crop

import (
	"fmt"
	"strings"
)

type Stage string

const (
	StageSingle  Stage = "single"
	StageInitial Stage = "initial"
	StageMid     Stage = "mid"
	StageEnd     Stage = "end"
)

type Window struct {
	Stage    Stage   `json:"stage"`
	FirstDay int     `json:"firstDay"`
	LastDay  int     `json:"lastDay"`
	Days     int     `json:"days"`
	Kc       float64 `json:"kc"`
}

func (w Window) Contains(day int) bool {
	return day >= w.FirstDay && day <= w.LastDay
}

func (w Window) String() string {
	if w.Stage == StageSingle {
		return fmt.Sprintf("%s Kc=%.3f", w.Stage, w.Kc)
	}
	return fmt.Sprintf("%s day %d..%d Kc=%.3f", w.Stage, w.FirstDay, w.LastDay, w.Kc)
}

func DescribeWindows(windows []Window) string {
	parts := make([]string, 0, len(windows))
	for _, w := range windows {
		parts = append(parts, w.String())
	}
	return strings.Join(parts, "; ")
}

func SeasonLength(windows []Window) int {
	last := 0
	for _, w := range windows {
		if w.LastDay > last {
			last = w.LastDay
		}
	}
	return last
}

func WindowForDay(windows []Window, day int) (Window, error) {
	if day < 1 {
		return Window{}, fmt.Errorf("%w: got day %d", ErrGrowthDayInvalid, day)
	}
	for _, w := range windows {
		if w.Contains(day) {
			return w, nil
		}
	}
	return Window{}, fmt.Errorf("%w: day %d is past the season length %d",
		ErrGrowthDayOutside, day, SeasonLength(windows))
}
