package crop

import (
	"fmt"
)

const UnstressedKs = 1.0

func ValidateStress(ks float64) error {
	if err := requireFinite("stressCoefficient", ks); err != nil {
		return err
	}
	if ks <= 0 || ks > 1 {
		return fmt.Errorf("%w: got %g", ErrStressOutOfRange, ks)
	}
	return nil
}

func ResolveStress(ks *float64) (float64, bool, error) {
	if ks == nil {
		return UnstressedKs, false, nil
	}
	if err := ValidateStress(*ks); err != nil {
		return 0, false, err
	}
	return *ks, true, nil
}

func ApplyStress(etc, ks float64) (float64, error) {
	if err := requireFinite("etc", etc); err != nil {
		return 0, err
	}
	if err := ValidateStress(ks); err != nil {
		return 0, err
	}
	if ks == UnstressedKs {
		return etc, nil
	}
	return ks * etc, nil
}

func StressReduction(potential, actual float64) float64 {
	return potential - actual
}
