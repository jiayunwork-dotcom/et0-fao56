package meteo

import (
	"fmt"
	"math"
)

// PressureAtElevation returns the atmospheric pressure in kPa at a station
// elevation in metres, using the simplified barometric law of FAO-56.
func PressureAtElevation(elevation float64) (float64, error) {
	if err := ValidateElevation(elevation); err != nil {
		return 0, err
	}
	ratio := (ReferenceTemperature - ElevationLapse*elevation) / ReferenceTemperature
	if ratio <= 0 {
		return 0, fmt.Errorf("%w: elevation %g m collapses the barometric ratio", ErrElevationOutOfRange, elevation)
	}
	return SeaLevelPressure * math.Pow(ratio, PressureExponent), nil
}

// ElevationForPressure inverts the barometric law, which lets a report
// state the elevation implied by a pressure reading.
func ElevationForPressure(pressure float64) (float64, error) {
	if err := ValidatePressure(pressure); err != nil {
		return 0, err
	}
	ratio := math.Pow(pressure/SeaLevelPressure, 1.0/PressureExponent)
	elevation := ReferenceTemperature * (1 - ratio) / ElevationLapse
	if err := ValidateElevation(elevation); err != nil {
		return 0, fmt.Errorf("pressure %g kPa implies an out-of-model elevation: %w", pressure, err)
	}
	return elevation, nil
}

// ValidateElevation keeps the station inside the range where the
// barometric law of FAO-56 is meaningful.
func ValidateElevation(elevation float64) error {
	if err := requireFinite("elevation", elevation); err != nil {
		return err
	}
	if elevation < MinElevation || elevation > MaxElevation {
		return fmt.Errorf("%w: %g m outside [%g, %g]", ErrElevationOutOfRange, elevation, MinElevation, MaxElevation)
	}
	return nil
}

// ValidatePressure rejects non-finite and non-positive pressures.
func ValidatePressure(pressure float64) error {
	if err := requireFinite("pressure", pressure); err != nil {
		return err
	}
	if pressure <= 0 {
		return fmt.Errorf("%w: got %g kPa", ErrNonPositivePressure, pressure)
	}
	return nil
}

// ResolvePressure picks the pressure to use: the explicit reading when
// given, otherwise the barometric estimate from the elevation, otherwise
// the sea level standard. The label names the source that was used.
func ResolvePressure(pressure, elevation *float64) (float64, string, error) {
	if pressure != nil {
		if err := ValidatePressure(*pressure); err != nil {
			return 0, "", err
		}
		if elevation != nil {
			if err := ValidateElevation(*elevation); err != nil {
				return 0, "", err
			}
		}
		return *pressure, "input", nil
	}
	if elevation != nil {
		p, err := PressureAtElevation(*elevation)
		if err != nil {
			return 0, "", err
		}
		return p, "elevation", nil
	}
	return SeaLevelPressure, "sea-level-default", nil
}
