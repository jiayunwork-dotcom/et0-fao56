package penman

import (
	"fmt"
	"math"
)

const (
	StatusOK   = "ok"
	StatusWarn = "warn"

	latentHeatWarnGap = 0.02
)

// Check is one audit line about the result.
type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// Audit reviews a finished result. Contradictions that make the answer
// meaningless are returned as errors; everything else is reported as a
// check line.
func Audit(r *Result, in Input) ([]Check, error) {
	if r == nil {
		return nil, fmt.Errorf("penman: nil result")
	}
	checks := make([]Check, 0, 6)

	slope, err := r.Air.SlopeCheck()
	if err != nil {
		return nil, err
	}
	slopeStatus := StatusOK
	if !slope.Agrees {
		slopeStatus = StatusWarn
	}
	checks = append(checks, Check{
		Name:   "delta-shares-es-temperature",
		Status: slopeStatus,
		Detail: fmt.Sprintf("delta=%.6f kPa/degC, numeric des/dT=%.6f at %.2f degC (relative gap %.2e)",
			slope.Analytic, slope.Numeric, slope.TemperatureCelsius, slope.RelativeGap),
	})

	if err := r.Energy.Check(); err != nil {
		return nil, err
	}
	checks = append(checks, Check{
		Name:   "energy-reconcile",
		Status: StatusOK,
		Detail: r.Energy.Describe(),
	})

	latentStatus := StatusOK
	if r.Energy.LatentHeatGap > latentHeatWarnGap {
		latentStatus = StatusWarn
	}
	checks = append(checks, Check{
		Name:   "latent-heat-consistency",
		Status: latentStatus,
		Detail: fmt.Sprintf("gamma uses lambda=%.4f MJ/kg, the %.3f depth factor implies %.4f MJ/kg (relative gap %.2e)",
			r.Energy.AirLatentHeat, r.Terms.Coefficients.DepthPerEnergy, r.Energy.CoefficientLatentHeat, r.Energy.LatentHeatGap),
	})

	if IsCalm(r.WindSpeed) {
		checks = append(checks, Check{
			Name:   "calm-degeneration",
			Status: StatusOK,
			Detail: fmt.Sprintf("u2=0 removes the aerodynamic term, denominator collapses to delta+gamma=%.6f", CalmDenominator(r.Delta, r.Gamma)),
		})
	} else {
		checks = append(checks, Check{
			Name:   "aerodynamic-share",
			Status: StatusOK,
			Detail: fmt.Sprintf("u2=%.3f m/s carries %.1f%% of ET0", r.WindSpeed, 100*r.Terms.AerodynamicShare()),
		})
	}

	energyStatus := StatusOK
	if r.AvailableEnergy < 0 {
		energyStatus = StatusWarn
	}
	checks = append(checks, Check{
		Name:   "available-energy-sign",
		Status: energyStatus,
		Detail: fmt.Sprintf("Rn-G=%.4f %s", r.AvailableEnergy, r.Terms.Coefficients.EnergyUnit),
	})

	if in.ClaimTranspired && r.ET0 <= 0 {
		return nil, fmt.Errorf("%w: input claims transpiration but Rn-G=%.4f %s gives ET0=%.4f %s",
			ErrSignContradiction, r.AvailableEnergy, r.Terms.Coefficients.EnergyUnit, r.ET0, r.Terms.Coefficients.DepthUnit)
	}
	if r.ET0 < 0 {
		checks = append(checks, Check{
			Name:   "et0-sign",
			Status: StatusWarn,
			Detail: fmt.Sprintf("ET0=%.4f %s is negative: the reading describes condensation, not transpiration",
				r.ET0, r.Terms.Coefficients.DepthUnit),
		})
	}
	return checks, nil
}

// Warnings filters the audit lines that are not ok.
func Warnings(checks []Check) []Check {
	var out []Check
	for _, c := range checks {
		if c.Status != StatusOK {
			out = append(out, c)
		}
	}
	return out
}

// CheckMonotoneNonDecreasing verifies that a series never falls by more
// than the given absolute slack.
func CheckMonotoneNonDecreasing(values []float64, slack float64) error {
	if len(values) == 0 {
		return ErrSweepEmpty
	}
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1]-math.Abs(slack) {
			return fmt.Errorf("%w: sample %d fell from %g to %g", ErrDeficitNotMonotone, i, values[i-1], values[i])
		}
	}
	return nil
}
