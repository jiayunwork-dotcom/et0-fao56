package penman

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrNegativeWindSpeed      = errors.New("penman: wind speed must be non-negative")
	ErrWindHeightOutOfRange   = errors.New("penman: wind measurement height outside the log profile")
	ErrUnknownTimeScale       = errors.New("penman: unknown time scale")
	ErrCoefficientNonPositive = errors.New("penman: coefficient must be positive")
	ErrTemperatureCollapse    = errors.New("penman: temperature offset collapses the aerodynamic term")
	ErrDenominatorNonPositive = errors.New("penman: Penman-Monteith denominator must be positive")
	ErrSignContradiction      = errors.New("penman: available energy and ET0 contradict the claimed transpiration")
	ErrEnergyMismatch         = errors.New("penman: latent heat flux does not reconcile with the two numerator terms")
	ErrNotFinite              = errors.New("penman: value must be finite")
	ErrSweepEmpty             = errors.New("penman: sweep needs at least one sample")
	ErrDeficitNotMonotone     = errors.New("penman: ET0 must not fall when the saturation deficit grows")
)

func requireFinite(name string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%w: %s=%v", ErrNotFinite, name, v)
	}
	return nil
}
