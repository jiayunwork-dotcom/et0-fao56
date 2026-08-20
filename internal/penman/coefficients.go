package penman

import (
	"fmt"
	"math"
	"sort"

	"et0-fao56/internal/meteo"
)

// TimeScale selects one row of the coefficient table.
type TimeScale string

const (
	ScaleDaily       TimeScale = "daily"
	ScaleHourlyDay   TimeScale = "hourly-day"
	ScaleHourlyNight TimeScale = "hourly-night"
	DefaultTimeScale           = ScaleDaily
)

// Coefficients is one row of the FAO-56 Penman-Monteith coefficient table.
// DepthPerEnergy is the 0.408 factor that turns MJ/m2 into mm, WindNumerator
// is the 900 (daily) or 37 (hourly) numerator of the aerodynamic term and
// WindDrag is the 0.34 (daily), 0.24 (hourly daytime) or 0.96 (hourly
// nighttime) factor in the denominator.
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

// CoefficientsFor returns the single table row for a time scale. An empty
// scale falls back to the daily row.
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

// KnownTimeScales lists the table keys in a stable order.
func KnownTimeScales() []TimeScale {
	out := make([]TimeScale, 0, len(coefficientTable))
	for scale := range coefficientTable {
		out = append(out, scale)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Validate checks that a table row is usable.
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

// ImpliedLatentHeat is the latent heat of vaporisation that the 0.408
// factor stands for, that is 1/0.408 MJ/kg.
func (c Coefficients) ImpliedLatentHeat() float64 {
	return 1.0 / c.DepthPerEnergy
}

// LatentHeatGap returns the relative gap between the latent heat used for
// gamma and the one implied by DepthPerEnergy.
func (c Coefficients) LatentHeatGap(latentHeat float64) (float64, error) {
	if err := meteo.ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	implied := c.ImpliedLatentHeat()
	return math.Abs(latentHeat-implied) / implied, nil
}

// WindWeight is the gamma*(Cn/(T+offset)) weight of the aerodynamic term.
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

// Describe renders the pinned constants for reports.
func (c Coefficients) Describe() string {
	return fmt.Sprintf("%s: %g / %g / %g, T+%g, %s -> %s",
		c.Scale, c.DepthPerEnergy, c.WindNumerator, c.WindDrag, c.KelvinOffset, c.EnergyUnit, c.DepthUnit)
}
