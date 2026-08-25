package meteo

import (
	"fmt"
	"math"
)

func SaturationVaporPressure(t Temperature) (float64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	c := t.Celsius()
	return TetensA * math.Exp(TetensB*c/(c+TetensC)), nil
}

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

func ActualFromDewpoint(dew Temperature) (float64, error) {
	return SaturationVaporPressure(dew)
}

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
