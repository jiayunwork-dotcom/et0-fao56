package meteo

import (
	"fmt"
	"strings"
)

// AirInput collects the raw humidity and pressure channels of one air
// reading. Pointers mark optional channels.
type AirInput struct {
	TemperatureCelsius float64
	Pressure           *float64
	Elevation          *float64
	LatentHeat         *float64
	Humidity           Humidity
}

// Air is the derived thermodynamic state of one air reading. Slope and
// SaturationVaporPressure come from the same temperature, and Gamma comes
// from the same pressure and latent heat that the report prints.
type Air struct {
	Temperature Temperature `json:"-"`

	TemperatureCelsius float64 `json:"temperatureCelsius"`
	TemperatureKelvin  float64 `json:"temperatureKelvin"`

	Pressure       float64 `json:"pressure"`
	PressureSource string  `json:"pressureSource"`

	LatentHeat       float64 `json:"latentHeat"`
	LatentHeatSource string  `json:"latentHeatSource"`

	Slope float64 `json:"delta"`
	Gamma float64 `json:"gamma"`

	SaturationVaporPressure float64 `json:"es"`
	ActualVaporPressure     float64 `json:"ea"`
	Deficit                 float64 `json:"deficit"`
	RelativeHumidity        float64 `json:"relativeHumidity"`
	HumiditySource          string  `json:"humiditySource"`
}

// Derive validates the reading and computes es, ea, the deficit, Delta and
// gamma in one pass so that every downstream term shares one state.
func Derive(in AirInput) (*Air, error) {
	t := Celsius(in.TemperatureCelsius)
	if err := t.Validate(); err != nil {
		return nil, fmt.Errorf("air temperature: %w", err)
	}
	pressure, pressureSource, err := ResolvePressure(in.Pressure, in.Elevation)
	if err != nil {
		return nil, err
	}
	latentHeat, latentSource, err := ResolveLatentHeat(in.LatentHeat)
	if err != nil {
		return nil, err
	}
	es, err := SaturationVaporPressure(t)
	if err != nil {
		return nil, err
	}
	ea, humiditySource, err := in.Humidity.Resolve(t)
	if err != nil {
		return nil, err
	}
	deficit, err := Deficit(es, ea)
	if err != nil {
		return nil, err
	}
	slope, err := Slope(t)
	if err != nil {
		return nil, err
	}
	gamma, err := PsychrometricConstant(pressure, latentHeat)
	if err != nil {
		return nil, err
	}
	rh, err := RelativeHumidityFromActual(t, ea)
	if err != nil {
		return nil, err
	}
	return &Air{
		Temperature:             t,
		TemperatureCelsius:      t.Celsius(),
		TemperatureKelvin:       t.Kelvin(),
		Pressure:                pressure,
		PressureSource:          pressureSource,
		LatentHeat:              latentHeat,
		LatentHeatSource:        latentSource,
		Slope:                   slope,
		Gamma:                   gamma,
		SaturationVaporPressure: es,
		ActualVaporPressure:     ea,
		Deficit:                 deficit,
		RelativeHumidity:        rh,
		HumiditySource:          humiditySource,
	}, nil
}

// WithDeficit returns a copy of the state whose actual vapour pressure is
// moved so the saturation deficit equals the requested value. Temperature,
// pressure, Delta and gamma are untouched.
func (a *Air) WithDeficit(deficit float64) (*Air, error) {
	ea, err := ActualForDeficit(a.Temperature, deficit)
	if err != nil {
		return nil, err
	}
	rh, err := RelativeHumidityFromActual(a.Temperature, ea)
	if err != nil {
		return nil, err
	}
	clone := *a
	clone.ActualVaporPressure = ea
	clone.Deficit = a.SaturationVaporPressure - ea
	clone.RelativeHumidity = rh
	clone.HumiditySource = "deficit-sweep"
	return &clone, nil
}

// SlopeCheck re-derives Delta numerically from the same temperature.
func (a *Air) SlopeCheck() (SlopeCheck, error) {
	return CheckSlope(a.Temperature)
}

// String renders the derived state for the CLI report.
func (a *Air) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "air %s\n", a.Temperature)
	fmt.Fprintf(&b, "  pressure   %.3f kPa (%s)\n", a.Pressure, a.PressureSource)
	fmt.Fprintf(&b, "  lambda     %.4f MJ/kg (%s)\n", a.LatentHeat, a.LatentHeatSource)
	fmt.Fprintf(&b, "  delta      %.5f kPa/degC\n", a.Slope)
	fmt.Fprintf(&b, "  gamma      %.5f kPa/degC\n", a.Gamma)
	fmt.Fprintf(&b, "  es         %.4f kPa\n", a.SaturationVaporPressure)
	fmt.Fprintf(&b, "  ea         %.4f kPa (%s)\n", a.ActualVaporPressure, a.HumiditySource)
	fmt.Fprintf(&b, "  es-ea      %.4f kPa\n", a.Deficit)
	fmt.Fprintf(&b, "  rh         %.1f %%\n", a.RelativeHumidity)
	return b.String()
}
