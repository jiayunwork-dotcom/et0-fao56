package meteo

import (
	"fmt"
	"math"
)

func PsychrometricConstant(pressure, latentHeat float64) (float64, error) {
	if err := ValidatePressure(pressure); err != nil {
		return 0, err
	}
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	return SpecificHeatMoistAir * pressure / (MolecularWeightRatio * latentHeat), nil
}

func PsychrometricFactor(latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	return SpecificHeatMoistAir / (MolecularWeightRatio * latentHeat), nil
}

func ValidateLatentHeat(latentHeat float64) error {
	if err := requireFinite("latentHeat", latentHeat); err != nil {
		return err
	}
	if latentHeat <= 0 {
		return fmt.Errorf("%w: got %g MJ/kg", ErrNonPositiveLatentHeat, latentHeat)
	}
	return nil
}

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

func ResolveLatentHeat(explicit *float64) (float64, string, error) {
	if explicit == nil {
		return LatentHeatReference, "fao56-reference", nil
	}
	if err := ValidateLatentHeat(*explicit); err != nil {
		return 0, "", err
	}
	return *explicit, "input", nil
}

func LatentHeatDeviation(latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	return math.Abs(latentHeat-LatentHeatReference) / LatentHeatReference, nil
}

func EnergyToDepth(energy, latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	if err := requireFinite("energy", energy); err != nil {
		return 0, err
	}
	return energy / latentHeat, nil
}

func DepthToEnergy(depth, latentHeat float64) (float64, error) {
	if err := ValidateLatentHeat(latentHeat); err != nil {
		return 0, err
	}
	if err := requireFinite("depth", depth); err != nil {
		return 0, err
	}
	return depth * latentHeat, nil
}
