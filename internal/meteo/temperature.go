package meteo

import (
	"fmt"
	"math"
)

// Temperature holds an air temperature in a single canonical unit so the
// saturation vapour pressure curve, its slope and the Penman-Monteith
// fraction can never be fed different units of the same reading.
type Temperature struct {
	celsius float64
}

// Celsius builds a Temperature from a Celsius reading.
func Celsius(v float64) Temperature {
	return Temperature{celsius: v}
}

// Kelvin builds a Temperature from a Kelvin reading.
func Kelvin(v float64) (Temperature, error) {
	if err := requireFinite("kelvin", v); err != nil {
		return Temperature{}, err
	}
	if v < 0 {
		return Temperature{}, fmt.Errorf("%w: %g K", ErrTemperatureBelowAbsoluteZero, v)
	}
	return Temperature{celsius: v - KelvinOffsetExact}, nil
}

// Value returns the Celsius reading.
func (t Temperature) Value() float64 {
	return t.celsius
}

// Celsius returns the Celsius reading.
func (t Temperature) Celsius() float64 {
	return t.celsius
}

// Kelvin returns the thermodynamic temperature using the exact offset.
func (t Temperature) Kelvin() float64 {
	return t.celsius + KelvinOffsetExact
}

// KelvinFAO returns the temperature with the 273 offset that FAO-56 uses
// inside the aerodynamic term of the daily equation.
func (t Temperature) KelvinFAO() float64 {
	return t.celsius + KelvinOffsetFAO
}

// Shifted returns a temperature moved by a Celsius increment; the unit of
// the increment is the unit of the receiver by construction.
func (t Temperature) Shifted(deltaCelsius float64) Temperature {
	return Temperature{celsius: t.celsius + deltaCelsius}
}

// Validate rejects non-finite readings, readings below absolute zero and
// readings outside the domain where the Tetens curve is defined.
func (t Temperature) Validate() error {
	if err := requireFinite("temperature", t.celsius); err != nil {
		return err
	}
	if t.celsius <= AbsoluteZeroCelsius {
		return fmt.Errorf("%w: %g degC", ErrTemperatureBelowAbsoluteZero, t.celsius)
	}
	if t.celsius <= -TetensC {
		return fmt.Errorf("%w: %g degC breaks the Tetens denominator", ErrTemperatureOutsideCurve, t.celsius)
	}
	if t.celsius < CurveMinCelsius || t.celsius > CurveMaxCelsius {
		return fmt.Errorf("%w: %g degC outside [%g, %g]", ErrTemperatureOutsideCurve,
			t.celsius, CurveMinCelsius, CurveMaxCelsius)
	}
	return nil
}

// String renders the reading with both units so reports cannot be read
// with the wrong scale.
func (t Temperature) String() string {
	return fmt.Sprintf("%.2f degC (%.2f K)", t.celsius, t.Kelvin())
}

// Mean averages two readings in the canonical unit.
func Mean(a, b Temperature) Temperature {
	return Temperature{celsius: 0.5 * (a.celsius + b.celsius)}
}

// Warmer reports whether a is above b.
func Warmer(a, b Temperature) bool {
	return a.celsius > b.celsius
}

// KelvinDifference returns the difference between two readings; the value
// is identical in Celsius and Kelvin because both scales share a degree.
func KelvinDifference(a, b Temperature) float64 {
	return a.celsius - b.celsius
}

// ValidateRange checks an ordered pair of daily extremes.
func ValidateRange(min, max Temperature) error {
	if err := min.Validate(); err != nil {
		return fmt.Errorf("minimum temperature: %w", err)
	}
	if err := max.Validate(); err != nil {
		return fmt.Errorf("maximum temperature: %w", err)
	}
	if max.celsius < min.celsius {
		return fmt.Errorf("meteo: maximum temperature %g degC below minimum %g degC", max.celsius, min.celsius)
	}
	return nil
}

func absDiff(a, b float64) float64 {
	return math.Abs(a - b)
}
