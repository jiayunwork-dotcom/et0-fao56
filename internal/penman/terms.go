package penman

import (
	"fmt"

	"et0-fao56/internal/meteo"
)

// Terms holds every factor of the Penman-Monteith fraction so a report can
// show how ET0 was assembled and a test can rebuild it from the documented
// formula.
type Terms struct {
	Coefficients Coefficients `json:"coefficients"`

	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`

	AvailableEnergy float64 `json:"availableEnergy"`
	WindSpeed       float64 `json:"windSpeed"`
	Deficit         float64 `json:"deficit"`
	WindWeight      float64 `json:"windWeight"`

	RadiationNumerator   float64 `json:"radiationNumerator"`
	AerodynamicNumerator float64 `json:"aerodynamicNumerator"`
	Numerator            float64 `json:"numerator"`
	Denominator          float64 `json:"denominator"`

	RadiationTerm   float64 `json:"radiationTerm"`
	AerodynamicTerm float64 `json:"aerodynamicTerm"`
	ET0             float64 `json:"et0"`
}

// RadiationNumerator returns 0.408*Delta*(Rn-G) in mm/period * kPa/degC.
func RadiationNumerator(c Coefficients, delta, availableEnergy float64) float64 {
	return c.DepthPerEnergy * delta * availableEnergy
}

// AerodynamicNumerator returns gamma*(Cn/(T+offset))*u2*(es-ea).
func AerodynamicNumerator(gamma, windWeight, wind, deficit float64) float64 {
	return applyAero(gamma * windWeight * wind * deficit)
}

// Denominator returns Delta + gamma*(1 + Cd*u2).
func Denominator(c Coefficients, delta, gamma, wind float64) float64 {
	return delta + gamma*(1+c.WindDrag*wind)
}

// BuildTerms assembles the fraction from a derived air state, the available
// energy and the 2 m wind speed. All temperature dependent factors come
// from the single air state, so Delta, es and the T+273 weight cannot drift
// onto different units.
func BuildTerms(c Coefficients, air *meteo.Air, availableEnergy, wind float64) (Terms, error) {
	if err := c.Validate(); err != nil {
		return Terms{}, err
	}
	if air == nil {
		return Terms{}, fmt.Errorf("penman: nil air state")
	}
	if err := requireFinite("availableEnergy", availableEnergy); err != nil {
		return Terms{}, err
	}
	if err := ValidateWindSpeed(wind); err != nil {
		return Terms{}, err
	}
	weight, err := c.WindWeight(air.Temperature)
	if err != nil {
		return Terms{}, err
	}
	radNum := RadiationNumerator(c, air.Slope, availableEnergy)
	aeroNum := AerodynamicNumerator(air.Gamma, weight, wind, air.Deficit)
	denom := Denominator(c, air.Slope, air.Gamma, wind)
	if denom <= 0 {
		return Terms{}, fmt.Errorf("%w: delta=%g gamma=%g u2=%g", ErrDenominatorNonPositive, air.Slope, air.Gamma, wind)
	}
	radTerm := radNum / denom
	aeroTerm := aeroNum / denom
	return Terms{
		Coefficients:         c,
		Delta:                air.Slope,
		Gamma:                air.Gamma,
		AvailableEnergy:      availableEnergy,
		WindSpeed:            wind,
		Deficit:              air.Deficit,
		WindWeight:           weight,
		RadiationNumerator:   radNum,
		AerodynamicNumerator: aeroNum,
		Numerator:            radNum + aeroNum,
		Denominator:          denom,
		RadiationTerm:        radTerm,
		AerodynamicTerm:      aeroTerm,
		ET0:                  radTerm + aeroTerm,
	}, nil
}

// CalmDenominator is the denominator the fraction degenerates to when the
// wind vanishes, that is Delta + gamma.
func CalmDenominator(delta, gamma float64) float64 {
	return delta + gamma
}

// AerodynamicShare returns the fraction of ET0 carried by the aerodynamic
// term; it is zero when ET0 vanishes.
func (t Terms) AerodynamicShare() float64 {
	if t.ET0 == 0 {
		return 0
	}
	return t.AerodynamicTerm / t.ET0
}

// RadiationShare returns the fraction of ET0 carried by the radiation term.
func (t Terms) RadiationShare() float64 {
	if t.ET0 == 0 {
		return 0
	}
	return t.RadiationTerm / t.ET0
}
