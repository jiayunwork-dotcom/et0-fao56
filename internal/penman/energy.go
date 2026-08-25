package penman

import (
	"fmt"
	"math"
)

const EnergyReconcileTolerance = 1e-9

type Energy struct {
	CoefficientLatentHeat float64 `json:"coefficientLatentHeat"`
	AirLatentHeat         float64 `json:"airLatentHeat"`
	LatentHeatGap         float64 `json:"latentHeatGap"`

	LatentHeatFlux    float64 `json:"latentHeatFlux"`
	RadiationEnergy   float64 `json:"radiationEnergy"`
	AerodynamicEnergy float64 `json:"aerodynamicEnergy"`
	AvailableEnergy   float64 `json:"availableEnergy"`

	Expected   float64 `json:"expected"`
	Residual   float64 `json:"residual"`
	Tolerance  float64 `json:"tolerance"`
	Reconciled bool    `json:"reconciled"`

	EvaporativeFraction float64 `json:"evaporativeFraction"`
}

func Reconcile(t Terms, airLatentHeat float64) (Energy, error) {
	if err := t.Coefficients.Validate(); err != nil {
		return Energy{}, err
	}
	if t.Denominator <= 0 {
		return Energy{}, fmt.Errorf("%w: denominator=%g", ErrDenominatorNonPositive, t.Denominator)
	}
	lambda := t.Coefficients.ImpliedLatentHeat()
	gap, err := t.Coefficients.LatentHeatGap(airLatentHeat)
	if err != nil {
		return Energy{}, err
	}
	flux := t.ET0 * lambda
	radiationEnergy := t.Delta * t.AvailableEnergy / t.Denominator
	aerodynamicEnergy := lambda * t.AerodynamicNumerator / t.Denominator
	expected := radiationEnergy + aerodynamicEnergy
	residual := flux - expected
	scale := math.Max(1, math.Abs(expected))
	energy := Energy{
		CoefficientLatentHeat: lambda,
		AirLatentHeat:         airLatentHeat,
		LatentHeatGap:         gap,
		LatentHeatFlux:        flux,
		RadiationEnergy:       radiationEnergy,
		AerodynamicEnergy:     aerodynamicEnergy,
		AvailableEnergy:       t.AvailableEnergy,
		Expected:              expected,
		Residual:              residual,
		Tolerance:             EnergyReconcileTolerance,
		Reconciled:            math.Abs(residual)/scale <= EnergyReconcileTolerance,
	}
	if t.AvailableEnergy != 0 {
		energy.EvaporativeFraction = flux / t.AvailableEnergy
	}
	return energy, nil
}

func (e Energy) Check() error {
	if e.Reconciled {
		return nil
	}
	return fmt.Errorf("%w: lambda*ET0=%g vs terms=%g (residual %g)",
		ErrEnergyMismatch, e.LatentHeatFlux, e.Expected, e.Residual)
}

func (e Energy) Describe() string {
	return fmt.Sprintf("lambda*ET0=%.4f = radiation %.4f + aerodynamic %.4f (residual %.2e)",
		e.LatentHeatFlux, e.RadiationEnergy, e.AerodynamicEnergy, e.Residual)
}
