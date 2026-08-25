package meteo

import (
	"fmt"
	"sort"
	"strings"
)

const (
	supersaturationTolerance = 1e-9

	humidityAgreementTolerance = 0.02
)

type Humidity struct {
	RelativeHumidity    *float64
	ActualVaporPressure *float64
	DewpointCelsius     *float64
}

func (h Humidity) Sources() []string {
	var out []string
	if h.ActualVaporPressure != nil {
		out = append(out, "actualVaporPressure")
	}
	if h.DewpointCelsius != nil {
		out = append(out, "dewpointTemperature")
	}
	if h.RelativeHumidity != nil {
		out = append(out, "relativeHumidity")
	}
	sort.Strings(out)
	return out
}

func (h Humidity) Resolve(t Temperature) (float64, string, error) {
	if err := t.Validate(); err != nil {
		return 0, "", err
	}
	type candidate struct {
		label string
		value float64
	}
	var candidates []candidate
	if h.ActualVaporPressure != nil {
		ea := *h.ActualVaporPressure
		if err := requireFinite("actualVaporPressure", ea); err != nil {
			return 0, "", err
		}
		if ea < 0 {
			return 0, "", fmt.Errorf("%w: ea=%g kPa", ErrNonPositiveVaporPressure, ea)
		}
		candidates = append(candidates, candidate{label: "actualVaporPressure", value: ea})
	}
	if h.DewpointCelsius != nil {
		dew := Celsius(*h.DewpointCelsius)
		if err := dew.Validate(); err != nil {
			return 0, "", fmt.Errorf("dewpoint temperature: %w", err)
		}
		if Warmer(dew, t) {
			return 0, "", fmt.Errorf("%w: dewpoint %s above air temperature %s", ErrSupersaturated, dew, t)
		}
		ea, err := ActualFromDewpoint(dew)
		if err != nil {
			return 0, "", err
		}
		candidates = append(candidates, candidate{label: "dewpointTemperature", value: ea})
	}
	if h.RelativeHumidity != nil {
		rh := *h.RelativeHumidity
		ea, err := ActualFromRelativeHumidity(t, rh)
		if err != nil {
			return 0, "", err
		}
		candidates = append(candidates, candidate{label: "relativeHumidity", value: ea})
	}
	if len(candidates) == 0 {
		return 0, "", ErrHumidityMissing
	}
	primary := candidates[0]
	for _, other := range candidates[1:] {
		if absDiff(primary.value, other.value) > humidityAgreementTolerance {
			return 0, "", fmt.Errorf("%w: %s gives ea=%.3f kPa but %s gives ea=%.3f kPa",
				ErrHumidityConflict, primary.label, primary.value, other.label, other.value)
		}
	}
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return 0, "", err
	}
	if primary.value > es+supersaturationTolerance {
		return 0, "", fmt.Errorf("%w: ea=%.4f kPa from %s above es=%.4f kPa at %s",
			ErrSupersaturated, primary.value, primary.label, es, t)
	}
	return primary.value, primary.label, nil
}

func (h Humidity) Describe() string {
	src := h.Sources()
	if len(src) == 0 {
		return "none"
	}
	return strings.Join(src, "+")
}

func ValidateRelativeHumidity(rh float64) error {
	if err := requireFinite("relativeHumidity", rh); err != nil {
		return err
	}
	if rh < 0 || rh > 100 {
		return fmt.Errorf("%w: got %g%%", ErrHumidityOutOfRange, rh)
	}
	return nil
}

func MeanRelativeHumidity(min, max float64) (float64, error) {
	if err := ValidateRelativeHumidity(min); err != nil {
		return 0, fmt.Errorf("minimum relative humidity: %w", err)
	}
	if err := ValidateRelativeHumidity(max); err != nil {
		return 0, fmt.Errorf("maximum relative humidity: %w", err)
	}
	if max < min {
		return 0, fmt.Errorf("meteo: maximum relative humidity %g%% below minimum %g%%", max, min)
	}
	return 0.5 * (min + max), nil
}
