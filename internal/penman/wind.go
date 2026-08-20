package penman

import (
	"fmt"
	"math"
)

const (
	// ReferenceWindHeight is the 2 m height the equation expects.
	ReferenceWindHeight = 2.0

	windProfileNumerator = 4.87
	windProfileScale     = 67.8
	windProfileOffset    = 5.42

	minWindHeight = 0.1
	maxWindHeight = 50.0

	calmWindSpeed = 0.0
)

// ValidateWindSpeed rejects non-finite and negative wind speeds.
func ValidateWindSpeed(speed float64) error {
	if err := requireFinite("windSpeed", speed); err != nil {
		return err
	}
	if speed < 0 {
		return fmt.Errorf("%w: got %g m/s", ErrNegativeWindSpeed, speed)
	}
	return nil
}

// ValidateWindHeight keeps the measurement height inside the range where
// the logarithmic profile of FAO-56 eq. 47 is defined.
func ValidateWindHeight(height float64) error {
	if err := requireFinite("windHeight", height); err != nil {
		return err
	}
	if height < minWindHeight || height > maxWindHeight {
		return fmt.Errorf("%w: %g m outside [%g, %g]", ErrWindHeightOutOfRange, height, minWindHeight, maxWindHeight)
	}
	if windProfileScale*height-windProfileOffset <= 1 {
		return fmt.Errorf("%w: %g m collapses the log profile", ErrWindHeightOutOfRange, height)
	}
	return nil
}

// WindAtTwoMetres converts a wind speed measured at an arbitrary height to
// the 2 m reference height. A nil height means the reading is already at
// 2 m.
func WindAtTwoMetres(speed float64, height *float64) (float64, error) {
	if err := ValidateWindSpeed(speed); err != nil {
		return 0, err
	}
	if height == nil {
		return speed, nil
	}
	if err := ValidateWindHeight(*height); err != nil {
		return 0, err
	}
	if *height == ReferenceWindHeight {
		return speed, nil
	}
	factor := windProfileNumerator / math.Log(windProfileScale**height-windProfileOffset)
	return speed * factor, nil
}

// WindProfileFactor exposes the multiplier that WindAtTwoMetres applies.
func WindProfileFactor(height float64) (float64, error) {
	if err := ValidateWindHeight(height); err != nil {
		return 0, err
	}
	return windProfileNumerator / math.Log(windProfileScale*height-windProfileOffset), nil
}

// IsCalm reports whether a wind speed switches the aerodynamic term off.
func IsCalm(speed float64) bool {
	return speed == calmWindSpeed
}
