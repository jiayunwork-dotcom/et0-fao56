package meteo

import (
	"fmt"
	"math"
)

// SaturationVaporPressure evaluates the Tetens form of the saturation
// vapour pressure curve, es(T), in kPa. The temperature enters in Celsius,
// the same unit that Slope differentiates with respect to.
func SaturationVaporPressure(t Temperature) (float64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	c := t.Celsius()
	return TetensA * math.Exp(TetensB*c/(c+TetensC)), nil
}

// MeanSaturationVaporPressure averages es(Tmin) and es(Tmax), which is the
// FAO-56 recommendation when daily extremes are available.
func MeanSaturationVaporPressure(min, max Temperature) (float64, error) {
	if err := ValidateRange(min, max); err != nil {
		return 0, err
	}
	esMin, err := SaturationVaporPressure(min)
	if err != nil {
		return 0, err
	}
	esMax, err := SaturationVaporPressure(max)
	if err != nil {
		return 0, err
	}
	return 0.5 * (esMin + esMax), nil
}

// ActualFromRelativeHumidity converts a relative humidity in percent into
// an actual vapour pressure using es at the same air temperature.
func ActualFromRelativeHumidity(t Temperature, rh float64) (float64, error) {
	if err := ValidateRelativeHumidity(rh); err != nil {
		return 0, err
	}
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	return es * rh / 100.0, nil
}

// ActualFromDewpoint returns the actual vapour pressure implied by a
// dewpoint temperature, that is es evaluated at the dewpoint.
func ActualFromDewpoint(dew Temperature) (float64, error) {
	return SaturationVaporPressure(dew)
}

// DewpointFromActual inverts the Tetens curve and returns the dewpoint of
// a given actual vapour pressure.
func DewpointFromActual(ea float64) (Temperature, error) {
	if err := requireFinite("actualVaporPressure", ea); err != nil {
		return Temperature{}, err
	}
	if ea <= 0 {
		return Temperature{}, fmt.Errorf("%w: ea=%g kPa", ErrNonPositiveVaporPressure, ea)
	}
	ratio := math.Log(ea / TetensA)
	denom := TetensB - ratio
	if denom <= 0 {
		return Temperature{}, fmt.Errorf("%w: ea=%g kPa is outside the invertible range", ErrTemperatureOutsideCurve, ea)
	}
	dew := Celsius(TetensC * ratio / denom)
	if err := dew.Validate(); err != nil {
		return Temperature{}, err
	}
	return dew, nil
}

// RelativeHumidityFromActual returns the relative humidity in percent that
// corresponds to an actual vapour pressure at the given air temperature.
func RelativeHumidityFromActual(t Temperature, ea float64) (float64, error) {
	if err := requireFinite("actualVaporPressure", ea); err != nil {
		return 0, err
	}
	if ea < 0 {
		return 0, fmt.Errorf("%w: ea=%g kPa", ErrNonPositiveVaporPressure, ea)
	}
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	if ea > es+supersaturationTolerance {
		return 0, fmt.Errorf("%w: ea=%g kPa above es=%g kPa at %s", ErrSupersaturated, ea, es, t)
	}
	return 100.0 * math.Min(ea, es) / es, nil
}

// Deficit returns the saturation deficit es-ea in kPa. A tiny excess of ea
// over es is treated as measurement noise and clamped to zero; anything
// larger is a supersaturated input and is rejected.
func Deficit(es, ea float64) (float64, error) {
	if err := requireFinite("saturationVaporPressure", es); err != nil {
		return 0, err
	}
	if err := requireFinite("actualVaporPressure", ea); err != nil {
		return 0, err
	}
	if es <= 0 {
		return 0, fmt.Errorf("%w: es=%g kPa", ErrNonPositiveVaporPressure, es)
	}
	if ea < 0 {
		return 0, fmt.Errorf("%w: ea=%g kPa", ErrNonPositiveVaporPressure, ea)
	}
	if ea > es+supersaturationTolerance {
		return 0, fmt.Errorf("%w: ea=%g kPa above es=%g kPa", ErrSupersaturated, ea, es)
	}
	if ea > es {
		return 0, nil
	}
	return es - ea, nil
}

// ActualForDeficit returns the actual vapour pressure that produces a
// requested saturation deficit at the given temperature.
func ActualForDeficit(t Temperature, deficit float64) (float64, error) {
	if err := requireFinite("deficit", deficit); err != nil {
		return 0, err
	}
	if deficit < 0 {
		return 0, fmt.Errorf("meteo: saturation deficit must be non-negative, got %g kPa", deficit)
	}
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, err
	}
	if deficit > es {
		return 0, fmt.Errorf("meteo: deficit %g kPa exceeds es=%g kPa at %s", deficit, es, t)
	}
	return es - deficit, nil
}
