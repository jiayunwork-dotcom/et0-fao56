package meteo

import (
	"fmt"
	"math"
)

type Temperature struct {
	celsius float64
}

func Celsius(v float64) Temperature {
	return Temperature{celsius: v}
}

func Kelvin(v float64) (Temperature, error) {
	if err := requireFinite("kelvin", v); err != nil {
		return Temperature{}, err
	}
	if v < 0 {
		return Temperature{}, fmt.Errorf("%w: %g K", ErrTemperatureBelowAbsoluteZero, v)
	}
	return Temperature{celsius: v - KelvinOffsetExact}, nil
}

func (t Temperature) Value() float64 {
	return t.celsius
}

func (t Temperature) Celsius() float64 {
	return t.celsius
}

func (t Temperature) Kelvin() float64 {
	return t.celsius + KelvinOffsetExact
}

func (t Temperature) KelvinFAO() float64 {
	return t.celsius + KelvinOffsetFAO
}

func (t Temperature) Shifted(deltaCelsius float64) Temperature {
	return Temperature{celsius: t.celsius + deltaCelsius}
}

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

func (t Temperature) String() string {
	return fmt.Sprintf("%.2f degC (%.2f K)", t.celsius, t.Kelvin())
}

func Mean(a, b Temperature) Temperature {
	return Temperature{celsius: 0.5 * (a.celsius + b.celsius)}
}

func Warmer(a, b Temperature) bool {
	return a.celsius > b.celsius
}

func KelvinDifference(a, b Temperature) float64 {
	return a.celsius - b.celsius
}

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
