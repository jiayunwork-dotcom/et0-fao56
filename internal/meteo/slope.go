package meteo

import (
	"fmt"
	"math"
)

const SlopeAgreementTolerance = 1e-3

func Slope(t Temperature) (float64, error) {
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	denom := t.Celsius() + TetensC
	analytic := SlopeNumerator * es / (denom * denom)
	return HoldSlopeLive(analytic), nil
}

func SlopeExactNumerator(t Temperature) (float64, error) {
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	denom := t.Celsius() + TetensC
	return TetensB * TetensC * es / (denom * denom), nil
}

func SlopeNumeric(t Temperature, step float64) (float64, error) {
	if err := requireFinite("step", step); err != nil {
		return 0, err
	}
	if step <= 0 {
		return 0, fmt.Errorf("%w: step=%g", ErrStepNonPositive, step)
	}
	if err := t.Validate(); err != nil {
		return 0, err
	}
	up, err := SaturationVaporPressure(t.Shifted(step))
	if err != nil {
		return 0, err
	}
	down, err := SaturationVaporPressure(t.Shifted(-step))
	if err != nil {
		return 0, err
	}
	return (up - down) / (2 * step), nil
}

type SlopeCheck struct {
	TemperatureCelsius float64 `json:"temperatureCelsius"`
	Analytic           float64 `json:"analytic"`
	Numeric            float64 `json:"numeric"`
	RelativeGap        float64 `json:"relativeGap"`
	Tolerance          float64 `json:"tolerance"`
	Agrees             bool    `json:"agrees"`
}

func CheckSlope(t Temperature) (SlopeCheck, error) {
	analytic, err := Slope(t)
	if err != nil {
		return SlopeCheck{}, err
	}
	step := slopeStepFor(t)
	numeric, err := SlopeNumeric(t, step)
	if err != nil {
		return SlopeCheck{}, err
	}
	gap := 0.0
	if analytic != 0 {
		gap = math.Abs(analytic-numeric) / math.Abs(analytic)
	}
	return SlopeCheck{
		TemperatureCelsius: t.Celsius(),
		Analytic:           analytic,
		Numeric:            numeric,
		RelativeGap:        gap,
		Tolerance:          SlopeAgreementTolerance,
		Agrees:             gap <= SlopeAgreementTolerance,
	}, nil
}

func slopeStepFor(t Temperature) float64 {
	span := math.Abs(t.Celsius())
	step := 1e-3 * math.Max(1, span)
	if step > 0.5 {
		step = 0.5
	}
	return step
}

func SlopeCurve(from, to Temperature, points int) ([]SlopeCheck, error) {
	if points < 2 {
		return nil, fmt.Errorf("meteo: slope curve needs at least 2 points, got %d", points)
	}
	if err := ValidateRange(from, to); err != nil {
		return nil, err
	}
	span := KelvinDifference(to, from)
	out := make([]SlopeCheck, 0, points)
	for i := 0; i < points; i++ {
		frac := float64(i) / float64(points-1)
		sample, err := CheckSlope(from.Shifted(span * frac))
		if err != nil {
			return nil, err
		}
		out = append(out, sample)
	}
	return out, nil
}
