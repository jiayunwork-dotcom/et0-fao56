package crop

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrKcNonPositive     = errors.New("crop: crop coefficient must be positive")
	ErrKcModeConflict    = errors.New("crop: give either a single Kc or the ini/mid/end triple, not both")
	ErrKcMissing         = errors.New("crop: no crop coefficient given")
	ErrKcIncomplete      = errors.New("crop: the staged Kc needs initial, mid and end values with their day counts")
	ErrStageDaysInvalid  = errors.New("crop: stage length must be a positive number of days")
	ErrGrowthDayRequired = errors.New("crop: a staged Kc needs the growth day")
	ErrGrowthDayInvalid  = errors.New("crop: growth day must be at least 1")
	ErrGrowthDayOutside  = errors.New("crop: growth day outside the season")
	ErrStressOutOfRange  = errors.New("crop: water stress coefficient must lie in (0, 1]")
	ErrNotFinite         = errors.New("crop: value must be finite")
	ErrReferenceMissing  = errors.New("crop: neither weather nor a given ET0 was supplied")
	ErrReferenceConflict = errors.New("crop: weather and a given ET0 disagree")
	ErrPlanMissing       = errors.New("crop: the document has no crop block")
	ErrDocumentEmpty     = errors.New("crop: empty document")
)

func requireFinite(name string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%w: %s=%v", ErrNotFinite, name, v)
	}
	return nil
}
