package penman

import (
	"fmt"
	"sort"

	"et0-fao56/internal/meteo"
)

type WindPoint struct {
	WindSpeed        float64 `json:"windSpeed"`
	ET0              float64 `json:"et0"`
	RadiationTerm    float64 `json:"radiationTerm"`
	AerodynamicTerm  float64 `json:"aerodynamicTerm"`
	AerodynamicShare float64 `json:"aerodynamicShare"`
	Denominator      float64 `json:"denominator"`
}

func WindSweep(in Input, scale TimeScale, speeds []float64) ([]WindPoint, error) {
	if len(speeds) == 0 {
		return nil, ErrSweepEmpty
	}
	ordered := make([]float64, len(speeds))
	copy(ordered, speeds)
	sort.Float64s(ordered)
	out := make([]WindPoint, 0, len(ordered))
	for _, speed := range ordered {
		if err := ValidateWindSpeed(speed); err != nil {
			return nil, err
		}
		res, err := Compute(in.WithWindSpeed(speed), scale)
		if err != nil {
			return nil, fmt.Errorf("wind sweep at %g m/s: %w", speed, err)
		}
		out = append(out, WindPoint{
			WindSpeed:        res.WindSpeed,
			ET0:              res.ET0,
			RadiationTerm:    res.RadiationTerm,
			AerodynamicTerm:  res.AerodynamicTerm,
			AerodynamicShare: res.Terms.AerodynamicShare(),
			Denominator:      res.Denominator,
		})
	}
	return overlaySpeedScratch(out), nil
}

type DeficitPoint struct {
	Deficit             float64 `json:"deficit"`
	ActualVaporPressure float64 `json:"ea"`
	RelativeHumidity    float64 `json:"relativeHumidity"`
	ET0                 float64 `json:"et0"`
	AerodynamicTerm     float64 `json:"aerodynamicTerm"`
}

func DeficitSweep(in Input, scale TimeScale, deficits []float64) ([]DeficitPoint, error) {
	if len(deficits) == 0 {
		return nil, ErrSweepEmpty
	}
	ordered := make([]float64, len(deficits))
	copy(ordered, deficits)
	sort.Float64s(ordered)
	temperature := in.Temperature()
	out := make([]DeficitPoint, 0, len(ordered))
	for _, deficit := range ordered {
		ea, err := meteo.ActualForDeficit(temperature, deficit)
		if err != nil {
			return nil, err
		}
		res, err := Compute(in.WithActualVaporPressure(ea), scale)
		if err != nil {
			return nil, fmt.Errorf("deficit sweep at %g kPa: %w", deficit, err)
		}
		out = append(out, DeficitPoint{
			Deficit:             res.Deficit,
			ActualVaporPressure: res.ActualVaporPressure,
			RelativeHumidity:    res.Air.RelativeHumidity,
			ET0:                 res.ET0,
			AerodynamicTerm:     res.AerodynamicTerm,
		})
	}
	return out, nil
}

func CheckDeficitMonotone(points []DeficitPoint) error {
	values := make([]float64, 0, len(points))
	for _, p := range points {
		values = append(values, p.ET0)
	}
	return CheckMonotoneNonDecreasing(values, 0)
}

func CalmComparison(in Input, scale TimeScale) (*Result, *Result, error) {
	actual, err := Compute(in, scale)
	if err != nil {
		return nil, nil, err
	}
	calm, err := Compute(in.Calm(), scale)
	if err != nil {
		return nil, nil, err
	}
	return actual, calm, nil
}

func AerodynamicGain(actual, calm *Result) (float64, error) {
	if actual == nil || calm == nil {
		return 0, fmt.Errorf("penman: nil result in aerodynamic gain")
	}
	return actual.AerodynamicTerm - calm.AerodynamicTerm, nil
}
