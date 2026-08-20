package crop

import (
	"fmt"
)

// Mode tells how a crop coefficient was given.
type Mode string

const (
	ModeSingle Mode = "single"
	ModeStaged Mode = "staged"
)

// Kc is a crop coefficient: either one value for the whole season, or the
// ini/mid/end triple with the length of each stage in days.
type Kc struct {
	Single *float64 `json:"single,omitempty"`

	Initial *float64 `json:"initial,omitempty"`
	Mid     *float64 `json:"mid,omitempty"`
	End     *float64 `json:"end,omitempty"`

	InitialDays *int `json:"initialDays,omitempty"`
	MidDays     *int `json:"midDays,omitempty"`
	EndDays     *int `json:"endDays,omitempty"`
}

// Mode reports whether the coefficient is a single value or staged.
func (k Kc) Mode() (Mode, error) {
	staged := k.Initial != nil || k.Mid != nil || k.End != nil ||
		k.InitialDays != nil || k.MidDays != nil || k.EndDays != nil
	if k.Single != nil && staged {
		return "", ErrKcModeConflict
	}
	if k.Single != nil {
		return ModeSingle, nil
	}
	if !staged {
		return "", ErrKcMissing
	}
	return ModeStaged, nil
}

// Validate checks the coefficient values and the stage lengths.
func (k Kc) Validate() error {
	mode, err := k.Mode()
	if err != nil {
		return err
	}
	if mode == ModeSingle {
		return validateValue("kc", *k.Single)
	}
	values := []struct {
		name  string
		value *float64
	}{
		{"kc.initial", k.Initial},
		{"kc.mid", k.Mid},
		{"kc.end", k.End},
	}
	for _, v := range values {
		if v.value == nil {
			return fmt.Errorf("%w: %s is missing", ErrKcIncomplete, v.name)
		}
		if err := validateValue(v.name, *v.value); err != nil {
			return err
		}
	}
	days := []struct {
		name  string
		value *int
	}{
		{"kc.initialDays", k.InitialDays},
		{"kc.midDays", k.MidDays},
		{"kc.endDays", k.EndDays},
	}
	for _, d := range days {
		if d.value == nil {
			return fmt.Errorf("%w: %s is missing", ErrKcIncomplete, d.name)
		}
		if *d.value < 1 {
			return fmt.Errorf("%w: %s=%d", ErrStageDaysInvalid, d.name, *d.value)
		}
	}
	return nil
}

// Windows expands the coefficient into its segments over the growth days.
func (k Kc) Windows() ([]Window, error) {
	if err := k.Validate(); err != nil {
		return nil, err
	}
	mode, err := k.Mode()
	if err != nil {
		return nil, err
	}
	if mode == ModeSingle {
		return []Window{{
			Stage:    StageSingle,
			FirstDay: 1,
			LastDay:  maxSeasonDay,
			Days:     maxSeasonDay,
			Kc:       *k.Single,
		}}, nil
	}
	segments := []struct {
		stage Stage
		days  int
		value float64
	}{
		{StageInitial, *k.InitialDays, *k.Initial},
		{StageMid, *k.MidDays, *k.Mid},
		{StageEnd, *k.EndDays, *k.End},
	}
	windows := make([]Window, 0, len(segments))
	cursor := 1
	for _, s := range segments {
		last := cursor + s.days - 1
		windows = append(windows, Window{
			Stage:    s.stage,
			FirstDay: cursor,
			LastDay:  last,
			Days:     s.days,
			Kc:       s.value,
		})
		cursor = last + 1
	}
	return windows, nil
}

// At returns the coefficient in force on a growth day, with the segment it
// came from.
func (k Kc) At(day int) (float64, Window, error) {
	windows, err := k.Windows()
	if err != nil {
		return 0, Window{}, err
	}
	if day < 1 {
		return 0, Window{}, fmt.Errorf("%w: got day %d", ErrGrowthDayInvalid, day)
	}
	window, err := WindowForDay(windows, day)
	if err != nil {
		return 0, Window{}, err
	}
	return window.Kc, window, nil
}

// SeasonLength is the number of growth days the curve covers; a single
// coefficient covers the whole season and reports zero.
func (k Kc) SeasonLength() (int, error) {
	mode, err := k.Mode()
	if err != nil {
		return 0, err
	}
	if mode == ModeSingle {
		return 0, nil
	}
	if err := k.Validate(); err != nil {
		return 0, err
	}
	return *k.InitialDays + *k.MidDays + *k.EndDays, nil
}

// SingleKc builds a single valued coefficient.
func SingleKc(value float64) Kc {
	v := value
	return Kc{Single: &v}
}

// StagedKc builds an ini/mid/end coefficient with its stage lengths.
func StagedKc(initial, mid, end float64, initialDays, midDays, endDays int) Kc {
	i, m, e := initial, mid, end
	id, md, ed := initialDays, midDays, endDays
	return Kc{
		Initial:     &i,
		Mid:         &m,
		End:         &e,
		InitialDays: &id,
		MidDays:     &md,
		EndDays:     &ed,
	}
}

const maxSeasonDay = 366

func validateValue(name string, value float64) error {
	if err := requireFinite(name, value); err != nil {
		return err
	}
	if value <= 0 {
		return commitKc(fmt.Errorf("%w: %s=%g", ErrKcNonPositive, name, value))
	}
	return nil
}
