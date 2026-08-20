package meteo

import (
	"fmt"
	"math"
)

// SlopeAgreementTolerance is the relative gap accepted between the FAO-56
// closed form of Delta and a numerical derivative of es. The FAO-56
// numerator 4098 is a rounding of TetensB*TetensC, so the two forms differ
// by a few parts in ten thousand.
const SlopeAgreementTolerance = 1e-3

// Slope returns Delta, the slope of the saturation vapour pressure curve at
// the given temperature, in kPa/degC. It differentiates the same es(T) that
// SaturationVaporPressure evaluates, with T in Celsius in both places.
func Slope(t Temperature) (float64, error) {
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	denom := t.Celsius() + TetensC
	return SlopeNumerator * es / (denom * denom), nil
}

// SlopeExactNumerator returns Delta using the analytic numerator
// TetensB*TetensC instead of the FAO-56 rounded 4098.
func SlopeExactNumerator(t Temperature) (float64, error) {
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	denom := t.Celsius() + TetensC
	return TetensB * TetensC * es / (denom * denom), nil
}

// SlopeNumeric approximates Delta with a central difference of es over the
// given step in Celsius degrees.
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

// SlopeCheck compares the closed form of Delta with a numerical derivative
// of the same curve, which is the guard that keeps Delta and es on one
// temperature unit.
type SlopeCheck struct {
	TemperatureCelsius float64 `json:"temperatureCelsius"`
	Analytic           float64 `json:"analytic"`
	Numeric            float64 `json:"numeric"`
	RelativeGap        float64 `json:"relativeGap"`
	Tolerance          float64 `json:"tolerance"`
	Agrees             bool    `json:"agrees"`
}

// CheckSlope evaluates both forms of Delta at the given temperature.
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

// SlopeCurve samples Delta over a temperature span, which is useful when a
// report needs to show how strongly the radiation weight varies with the
// air temperature.
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
