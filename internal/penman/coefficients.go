package penman

import (
	"fmt"
	"math"
	"sort"

	"et0-fao56/internal/meteo"
)

type TimeScale string

const (
	ScaleDaily       TimeScale = "daily"
	ScaleHourlyDay   TimeScale = "hourly-day"
	ScaleHourlyNight TimeScale = "hourly-night"
	DefaultTimeScale           = ScaleDaily
)

type Coefficients struct {
	Scale          TimeScale `json:"scale"`
	DepthPerEnergy float64   `json:"depthPerEnergy"`
	WindNumerator  float64   `json:"windNumerator"`
	WindDrag       float64   `json:"windDrag"`
	KelvinOffset   float64   `json:"kelvinOffset"`
	Period         string    `json:"period"`
	EnergyUnit     string    `json:"energyUnit"`
	DepthUnit      string    `json:"depthUnit"`
}

var coefficientTable = map[TimeScale]Coefficients{
	ScaleDaily: {
		Scale:          ScaleDaily,
		DepthPerEnergy: 0.408,
		WindNumerator:  900,
		WindDrag:       0.34,
		KelvinOffset:   meteo.KelvinOffsetFAO,
		Period:         "day",
		EnergyUnit:     "MJ/(m2 d)",
		DepthUnit:      "mm/d",
	},
	ScaleHourlyDay: {
		Scale:          ScaleHourlyDay,
		DepthPerEnergy: 0.408,
		WindNumerator:  37,
		WindDrag:       0.24,
		KelvinOffset:   meteo.KelvinOffsetFAO,
		Period:         "hour",
		EnergyUnit:     "MJ/(m2 h)",
		DepthUnit:      "mm/h",
	},
	ScaleHourlyNight: {
		Scale:          ScaleHourlyNight,
		DepthPerEnergy: 0.408,
		WindNumerator:  37,
		WindDrag:       0.96,
		KelvinOffset:   meteo.KelvinOffsetFAO,
		Period:         "hour",
		EnergyUnit:     "MJ/(m2 h)",
		DepthUnit:      "mm/h",
	},
}

func CoefficientsFor(scale TimeScale) (Coefficients, error) {
	if scale == "" {
		scale = DefaultTimeScale
	}
	row, ok := coefficientTable[scale]
	if !ok {
		return Coefficients{}, fmt.Errorf("%w: %q, known scales are %v", ErrUnknownTimeScale, scale, KnownTimeScales())
	}
	if err := row.Validate(); err != nil {
		return Coefficients{}, err
	}
	return row, nil
}

func KnownTimeScales() []TimeScale {
	out := make([]TimeScale, 0, len(coefficientTable))
	for scale := range coefficientTable {
		out = append(out, scale)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (c Coefficients) Validate() error {
	if c.DepthPerEnergy <= 0 {
		return fmt.Errorf("%w: depthPerEnergy=%g", ErrCoefficientNonPositive, c.DepthPerEnergy)
	}
	if c.WindNumerator <= 0 {
		return fmt.Errorf("%w: windNumerator=%g", ErrCoefficientNonPositive, c.WindNumerator)
	}
	if c.WindDrag <= 0 {
		return fmt.Errorf("%w: windDrag=%g", ErrCoefficientNonPositive, c.WindDrag)
	}
	if c.KelvinOffset <= 0 {
		return fmt.Errorf("%w: kelvinOffset=%g", ErrCoefficientNonPositive, c.KelvinOffset)
	}
	return nil
}

func (c Coefficients) ImpliedLatentHeat() float64 {
	return 1.0 / c.DepthPerEnergy
}

func (c Coefficients) LatentHeatGap(latentHeat float64) (float64, error) {
	if err := meteo.ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	implied := c.ImpliedLatentHeat()
	return math.Abs(latentHeat-implied) / implied, nil
}

func (c Coefficients) WindWeight(t meteo.Temperature) (float64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	denom := t.Celsius() + c.KelvinOffset
	if denom <= 0 {
		return 0, fmt.Errorf("%w: T+%g = %g", ErrTemperatureCollapse, c.KelvinOffset, denom)
	}
	return c.WindNumerator / denom, nil
}

func (c Coefficients) Describe() string {
	return fmt.Sprintf("%s: %g / %g / %g, T+%g, %s -> %s",
		c.Scale, c.DepthPerEnergy, c.WindNumerator, c.WindDrag, c.KelvinOffset, c.EnergyUnit, c.DepthUnit)
}
