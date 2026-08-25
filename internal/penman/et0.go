package penman

import (
	"fmt"

	"et0-fao56/internal/meteo"
)

type Result struct {
	Scale TimeScale `json:"scale"`

	ET0             float64 `json:"et0"`
	RadiationTerm   float64 `json:"radiationTerm"`
	AerodynamicTerm float64 `json:"aerodynamicTerm"`

	Delta       float64 `json:"delta"`
	Gamma       float64 `json:"gamma"`
	Denominator float64 `json:"denominator"`

	SaturationVaporPressure float64 `json:"es"`
	ActualVaporPressure     float64 `json:"ea"`
	Deficit                 float64 `json:"deficit"`

	AvailableEnergy float64 `json:"availableEnergy"`
	WindSpeed       float64 `json:"windSpeed"`
	WindSpeedInput  float64 `json:"windSpeedInput"`
	WindHeight      float64 `json:"windHeight"`

	DepthUnit  string `json:"depthUnit"`
	EnergyUnit string `json:"energyUnit"`

	Air    *meteo.Air `json:"air"`
	Terms  Terms      `json:"terms"`
	Energy Energy     `json:"energy"`
	Checks []Check    `json:"checks"`
}

func Compute(in Input, scale TimeScale) (*Result, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	coeffs, err := CoefficientsFor(scale)
	if err != nil {
		return nil, err
	}
	air, err := meteo.Derive(in.AirInput())
	if err != nil {
		return nil, err
	}
	wind, err := WindAtTwoMetres(in.WindSpeed, in.WindHeight)
	if err != nil {
		return nil, err
	}
	terms, err := BuildTerms(coeffs, air, in.AvailableEnergy(), wind)
	if err != nil {
		return nil, err
	}
	energy, err := Reconcile(terms, air.LatentHeat)
	if err != nil {
		return nil, err
	}
	height := ReferenceWindHeight
	if in.WindHeight != nil {
		height = *in.WindHeight
	}
	res := &Result{
		Scale:                   coeffs.Scale,
		ET0:                     terms.ET0,
		RadiationTerm:           terms.RadiationTerm,
		AerodynamicTerm:         terms.AerodynamicTerm,
		Delta:                   terms.Delta,
		Gamma:                   terms.Gamma,
		Denominator:             terms.Denominator,
		SaturationVaporPressure: air.SaturationVaporPressure,
		ActualVaporPressure:     air.ActualVaporPressure,
		Deficit:                 air.Deficit,
		AvailableEnergy:         terms.AvailableEnergy,
		WindSpeed:               wind,
		WindSpeedInput:          in.WindSpeed,
		WindHeight:              height,
		DepthUnit:               coeffs.DepthUnit,
		EnergyUnit:              coeffs.EnergyUnit,
		Air:                     air,
		Terms:                   terms,
		Energy:                  energy,
	}
	checks, err := Audit(res, in)
	if err != nil {
		return nil, err
	}
	res.Checks = checks
	return res, nil
}

func ComputeDaily(in Input) (*Result, error) {
	return Compute(in, ScaleDaily)
}

func (r *Result) Rebuild() (float64, error) {
	if r == nil {
		return 0, fmt.Errorf("penman: nil result")
	}
	coeffs := r.Terms.Coefficients
	if err := coeffs.Validate(); err != nil {
		return 0, err
	}
	weight, err := coeffs.WindWeight(meteo.Celsius(r.Air.TemperatureCelsius))
	if err != nil {
		return 0, err
	}
	numerator := coeffs.DepthPerEnergy*r.Delta*r.AvailableEnergy +
		r.Gamma*weight*r.WindSpeed*r.Deficit
	denominator := Denominator(coeffs, r.Delta, r.Gamma, r.WindSpeed)
	if denominator <= 0 {
		return 0, fmt.Errorf("%w: denominator=%g", ErrDenominatorNonPositive, denominator)
	}
	return numerator / denominator, nil
}

func (r *Result) LatentHeatFlux() float64 {
	return r.Energy.LatentHeatFlux
}
