package crop

import (
	"fmt"
)

// Result is the crop evapotranspiration of one day together with the
// coefficient that produced it.
type Result struct {
	ET0  float64 `json:"et0"`
	Kc   float64 `json:"kc"`
	Mode Mode    `json:"mode"`

	Stage     Stage  `json:"stage"`
	GrowthDay int    `json:"growthDay"`
	Window    Window `json:"window"`

	StressCoefficient float64 `json:"stressCoefficient"`
	Stressed          bool    `json:"stressed"`

	ETcPotential float64 `json:"etcPotential"`
	ETc          float64 `json:"etc"`

	SeasonLength int          `json:"seasonLength"`
	Windows      []Window     `json:"windows"`
	StageTable   []StageValue `json:"stageTable"`
}

// Evaluate multiplies a reference evapotranspiration by the crop
// coefficient in force on the growth day and applies the optional water
// stress coefficient.
func Evaluate(et0 float64, spec Kc, growthDay *int, stress *float64) (*Result, error) {
	if err := requireFinite("et0", et0); err != nil {
		return nil, err
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	mode, err := spec.Mode()
	if err != nil {
		return nil, err
	}
	day, err := resolveGrowthDay(mode, growthDay)
	if err != nil {
		return nil, err
	}
	kc, window, err := spec.At(day)
	if err != nil {
		return nil, err
	}
	ks, stressed, err := ResolveStress(stress)
	if err != nil {
		return nil, err
	}
	potential := kc * et0
	actual, err := ApplyStress(potential, ks)
	if err != nil {
		return nil, err
	}
	windows, err := spec.Windows()
	if err != nil {
		return nil, err
	}
	season, err := spec.SeasonLength()
	if err != nil {
		return nil, err
	}
	table, err := StageTable(et0, spec, stress)
	if err != nil {
		return nil, err
	}
	return &Result{
		ET0:               et0,
		Kc:                kc,
		Mode:              mode,
		Stage:             window.Stage,
		GrowthDay:         day,
		Window:            window,
		StressCoefficient: ks,
		Stressed:          stressed,
		ETcPotential:      potential,
		ETc:               actual,
		SeasonLength:      season,
		Windows:           windows,
		StageTable:        table,
	}, nil
}

func resolveGrowthDay(mode Mode, growthDay *int) (int, error) {
	if growthDay == nil {
		if mode == ModeStaged {
			return 0, ErrGrowthDayRequired
		}
		return 1, nil
	}
	if *growthDay < 1 {
		return 0, fmt.Errorf("%w: got %d", ErrGrowthDayInvalid, *growthDay)
	}
	return *growthDay, nil
}

// PotentialRatio returns ETc divided by ET0, that is Kc times Ks.
func (r *Result) PotentialRatio() float64 {
	if r == nil || r.ET0 == 0 {
		return 0
	}
	return r.ETc / r.ET0
}

// StressLoss returns the depth removed by the water stress coefficient.
func (r *Result) StressLoss() float64 {
	if r == nil {
		return 0
	}
	return StressReduction(r.ETcPotential, r.ETc)
}
