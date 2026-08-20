package meteo

import (
	"fmt"
	"math"
)

// PsychrometricConstant returns gamma in kPa/degC from the atmospheric
// pressure in kPa and the latent heat of vaporisation in MJ/kg:
// gamma = cp*P/(epsilon*lambda).
func PsychrometricConstant(pressure, latentHeat float64) (float64, error) {
	if err := ValidatePressure(pressure); err != nil {
		return 0, err
	}
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	return SpecificHeatMoistAir * pressure / (MolecularWeightRatio * latentHeat), nil
}

// PsychrometricFactor returns the cp/(epsilon*lambda) factor that multiplies
// the pressure; with the FAO-56 reference latent heat it is 0.000665.
func PsychrometricFactor(latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	return SpecificHeatMoistAir / (MolecularWeightRatio * latentHeat), nil
}

// ValidateLatentHeat rejects non-finite and non-positive latent heat.
func ValidateLatentHeat(latentHeat float64) error {
	if err := requireFinite("latentHeat", latentHeat); err != nil {
		return err
	}
	if latentHeat <= 0 {
		return fmt.Errorf("%w: got %g MJ/kg", ErrNonPositiveLatentHeat, latentHeat)
	}
	return nil
}

// LatentHeatAt returns the temperature dependent latent heat of
// vaporisation in MJ/kg. FAO-56 pins 2.45 MJ/kg for daily work; this form
// is available for reporting how far a day sits from that reference.
func LatentHeatAt(t Temperature) (float64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	lambda := LatentHeatIntercept - LatentHeatSlope*t.Celsius()
	if lambda <= 0 {
		return 0, fmt.Errorf("%w: %g MJ/kg at %s", ErrNonPositiveLatentHeat, lambda, t)
	}
	return lambda, nil
}

// ResolveLatentHeat picks the latent heat to use: the explicit value when
// given, otherwise the FAO-56 reference 2.45 MJ/kg.
func ResolveLatentHeat(explicit *float64) (float64, string, error) {
	if explicit == nil {
		return LatentHeatReference, "fao56-reference", nil
	}
	if err := ValidateLatentHeat(*explicit); err != nil {
		return 0, "", err
	}
	return *explicit, "input", nil
}

// LatentHeatDeviation returns the relative gap between a latent heat and
// the FAO-56 reference value.
func LatentHeatDeviation(latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	return math.Abs(latentHeat-LatentHeatReference) / LatentHeatReference, nil
}

// EnergyToDepth converts a latent heat flux in MJ/(m2 d) into an
// evaporation depth in mm/d. One millimetre of water is one kg/m2.
func EnergyToDepth(energy, latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	if err := requireFinite("energy", energy); err != nil {
		return 0, err
	}
	return energy / latentHeat, nil
}

// DepthToEnergy converts an evaporation depth in mm/d into the equivalent
// latent heat flux in MJ/(m2 d).
func DepthToEnergy(depth, latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	if err := requireFinite("depth", depth); err != nil {
		return 0, err
	}
	return depth * latentHeat, nil
}
