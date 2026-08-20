package crop

import (
	"fmt"
)

// UnstressedKs is the water stress coefficient of a fully supplied soil.
const UnstressedKs = 1.0

// ValidateStress rejects stress coefficients outside (0, 1].
func ValidateStress(ks float64) error {
	if err := requireFinite("stressCoefficient", ks); err != nil {
		return err
	}
	if ks <= 0 || ks > 1 {
		return fmt.Errorf("%w: got %g", ErrStressOutOfRange, ks)
	}
	return nil
}

// ResolveStress picks the stress coefficient to apply: the explicit value
// when given, otherwise the unstressed 1.0. The bool reports whether the
// value came from the input.
func ResolveStress(ks *float64) (float64, bool, error) {
	if ks == nil {
		return UnstressedKs, false, nil
	}
	if err := ValidateStress(*ks); err != nil {
		return 0, false, err
	}
	return *ks, true, nil
}

// ApplyStress multiplies a potential crop evapotranspiration by the stress
// coefficient. A coefficient of 1 returns the input unchanged.
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

// StressReduction returns how much depth the stress removed.
func StressReduction(potential, actual float64) float64 {
	return potential - actual
}
