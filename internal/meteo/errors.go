package meteo

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrTemperatureBelowAbsoluteZero = errors.New("meteo: temperature below absolute zero")
	ErrTemperatureOutsideCurve      = errors.New("meteo: temperature outside the saturation curve domain")
	ErrHumidityOutOfRange           = errors.New("meteo: relative humidity must lie in [0, 100] percent")
	ErrHumidityMissing              = errors.New("meteo: no humidity input given (relative humidity, dewpoint or actual vapour pressure)")
	ErrHumidityConflict             = errors.New("meteo: humidity inputs disagree")
	ErrNonPositivePressure          = errors.New("meteo: atmospheric pressure must be positive")
	ErrNonPositiveLatentHeat        = errors.New("meteo: latent heat of vaporisation must be positive")
	ErrNonPositiveVaporPressure     = errors.New("meteo: vapour pressure must be positive")
	ErrSupersaturated               = errors.New("meteo: actual vapour pressure exceeds saturation vapour pressure")
	ErrElevationOutOfRange          = errors.New("meteo: elevation outside the barometric model")
	ErrNotFinite                    = errors.New("meteo: value must be finite")
	ErrStepNonPositive              = errors.New("meteo: differentiation step must be positive")
)

func requireFinite(name string, v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%w: %s=%v", ErrNotFinite, name, v)
	}
	return nil
}
