package crop

import (
	"fmt"
	"strings"
)

// Stage names one segment of the crop coefficient curve.
type Stage string

const (
	StageSingle  Stage = "single"
	StageInitial Stage = "initial"
	StageMid     Stage = "mid"
	StageEnd     Stage = "end"
)

// Window is one segment of the curve with the growth days it covers and
// the coefficient that holds there.
type Window struct {
	Stage    Stage   `json:"stage"`
	FirstDay int     `json:"firstDay"`
	LastDay  int     `json:"lastDay"`
	Days     int     `json:"days"`
	Kc       float64 `json:"kc"`
}

// Contains reports whether a growth day falls into the window.
func (w Window) Contains(day int) bool {
	return day >= w.FirstDay && day <= w.LastDay
}

// String renders the window for reports.
func (w Window) String() string {
	if w.Stage == StageSingle {
		return fmt.Sprintf("%s Kc=%.3f", w.Stage, w.Kc)
	}
	return fmt.Sprintf("%s day %d..%d Kc=%.3f", w.Stage, w.FirstDay, w.LastDay, w.Kc)
}

// DescribeWindows renders a whole curve.
func DescribeWindows(windows []Window) string {
	parts := make([]string, 0, len(windows))
	for _, w := range windows {
		parts = append(parts, w.String())
	}
	return strings.Join(parts, "; ")
}

// SeasonLength is the last growth day covered by the curve.
func SeasonLength(windows []Window) int {
	last := 0
	for _, w := range windows {
		if w.LastDay > last {
			last = w.LastDay
		}
	}
	return last
}

// WindowForDay finds the segment that owns a growth day.
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
