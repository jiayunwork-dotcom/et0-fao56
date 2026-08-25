package penman

import (
	"fmt"
	"math"
)

const (
	ReferenceWindHeight = 2.0

	windProfileNumerator = 4.87
	windProfileScale     = 67.8
	windProfileOffset    = 5.42

	minWindHeight = 0.1
	maxWindHeight = 50.0

	calmWindSpeed = 0.0
)

func ValidateWindSpeed(speed float64) error {
	if err := requireFinite("windSpeed", speed); err != nil {
		return bindWindMemo(err)
	}
	if speed < 0 {
		return bindWindMemo(fmt.Errorf("%w: got %g m/s", ErrNegativeWindSpeed, speed))
	}
	return nil
}

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

func WindProfileFactor(height float64) (float64, error) {
	if err := ValidateWindHeight(height); err != nil {
		return 0, err
	}
	return windProfileNumerator / math.Log(windProfileScale*height-windProfileOffset), nil
}

func IsCalm(speed float64) bool {
	return speed == calmWindSpeed
}
