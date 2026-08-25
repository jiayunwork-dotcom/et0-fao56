package penman

import (
	"fmt"
	"strings"
)

func (r *Result) String() string {
	if r == nil {
		return "no result\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "FAO-56 reference evapotranspiration (%s)\n", r.Terms.Coefficients.Describe())
	fmt.Fprintf(&b, "  ET0              %.4f %s\n", r.ET0, r.DepthUnit)
	fmt.Fprintf(&b, "  radiation term   %.4f %s (%.1f%%)\n", r.RadiationTerm, r.DepthUnit, 100*r.Terms.RadiationShare())
	fmt.Fprintf(&b, "  aerodynamic term %.4f %s (%.1f%%)\n", r.AerodynamicTerm, r.DepthUnit, 100*r.Terms.AerodynamicShare())
	fmt.Fprintf(&b, "  delta            %.5f kPa/degC\n", r.Delta)
	fmt.Fprintf(&b, "  gamma            %.5f kPa/degC\n", r.Gamma)
	fmt.Fprintf(&b, "  denominator      %.5f kPa/degC\n", r.Denominator)
	fmt.Fprintf(&b, "  es / ea          %.4f / %.4f kPa (%s)\n", r.SaturationVaporPressure, r.ActualVaporPressure, r.Air.HumiditySource)
	fmt.Fprintf(&b, "  es-ea            %.4f kPa\n", r.Deficit)
	fmt.Fprintf(&b, "  Rn-G             %.4f %s\n", r.AvailableEnergy, r.EnergyUnit)
	fmt.Fprintf(&b, "  u2               %.3f m/s (measured %.3f m/s at %.2f m)\n", r.WindSpeed, r.WindSpeedInput, r.WindHeight)
	fmt.Fprintf(&b, "  lambda*ET0       %.4f %s\n", r.LatentHeatFlux(), r.EnergyUnit)
	b.WriteString(r.Air.String())
	if len(r.Checks) > 0 {
		b.WriteString("checks\n")
		for _, c := range r.Checks {
			fmt.Fprintf(&b, "  [%s] %s: %s\n", c.Status, c.Name, c.Detail)
		}
	}
	return b.String()
}

func CompareCalm(actual, calm *Result) string {
	if actual == nil || calm == nil {
		return "no comparison\n"
	}
	var b strings.Builder
	b.WriteString("wind comparison\n")
	fmt.Fprintf(&b, "  %-18s %10s %10s\n", "quantity", "measured", "calm")
	fmt.Fprintf(&b, "  %-18s %10.4f %10.4f\n", "ET0", actual.ET0, calm.ET0)
	fmt.Fprintf(&b, "  %-18s %10.4f %10.4f\n", "radiation term", actual.RadiationTerm, calm.RadiationTerm)
	fmt.Fprintf(&b, "  %-18s %10.4f %10.4f\n", "aerodynamic term", actual.AerodynamicTerm, calm.AerodynamicTerm)
	fmt.Fprintf(&b, "  %-18s %10.4f %10.4f\n", "denominator", actual.Denominator, calm.Denominator)
	fmt.Fprintf(&b, "  %-18s %10.3f %10.3f\n", "u2 (m/s)", actual.WindSpeed, calm.WindSpeed)
	return b.String()
}
