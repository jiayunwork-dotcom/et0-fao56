package penman

import (
	"fmt"

	"et0-fao56/internal/meteo"
)

// Input is one day (or hour) of weather. Energy fluxes are MJ/(m2 period),
// temperatures Celsius, wind m/s, vapour pressures kPa.
type Input struct {
	NetRadiation   float64  `json:"netRadiation"`
	SoilHeatFlux   float64  `json:"soilHeatFlux,omitempty"`
	AirTemperature float64  `json:"airTemperature"`
	WindSpeed      float64  `json:"windSpeed"`
	WindHeight     *float64 `json:"windHeight,omitempty"`

	RelativeHumidity    *float64 `json:"relativeHumidity,omitempty"`
	ActualVaporPressure *float64 `json:"actualVaporPressure,omitempty"`
	DewpointTemperature *float64 `json:"dewpointTemperature,omitempty"`

	Pressure   *float64 `json:"pressure,omitempty"`
	Elevation  *float64 `json:"elevation,omitempty"`
	LatentHeat *float64 `json:"latentHeat,omitempty"`

	ClaimTranspired bool `json:"claimTranspired,omitempty"`
}

// Validate checks the raw channels before any thermodynamics is derived.
func (in Input) Validate() error {
	if err := requireFinite("netRadiation", in.NetRadiation); err != nil {
		return err
	}
	if err := requireFinite("soilHeatFlux", in.SoilHeatFlux); err != nil {
		return err
	}
	if err := requireFinite("airTemperature", in.AirTemperature); err != nil {
		return err
	}
	if err := ValidateWindSpeed(in.WindSpeed); err != nil {
		return err
	}
	if in.WindHeight != nil {
		if err := ValidateWindHeight(*in.WindHeight); err != nil {
			return err
		}
	}
	if in.RelativeHumidity != nil {
		if err := meteo.ValidateRelativeHumidity(*in.RelativeHumidity); err != nil {
			return err
		}
	}
	if in.Pressure != nil {
		if err := meteo.ValidatePressure(*in.Pressure); err != nil {
			return err
		}
	}
	if in.Elevation != nil {
		if err := meteo.ValidateElevation(*in.Elevation); err != nil {
			return err
		}
	}
	if in.LatentHeat != nil {
		if err := meteo.ValidateLatentHeat(*in.LatentHeat); err != nil {
			return err
		}
	}
	if in.Humidity().Sources() == nil {
		return fmt.Errorf("weather: %w", meteo.ErrHumidityMissing)
	}
	return nil
}

// Humidity groups the alternative humidity channels of the reading.
func (in Input) Humidity() meteo.Humidity {
	return meteo.Humidity{
		RelativeHumidity:    in.RelativeHumidity,
		ActualVaporPressure: in.ActualVaporPressure,
		DewpointCelsius:     in.DewpointTemperature,
	}
}

// AirInput maps the reading onto the thermodynamic input of meteo.
func (in Input) AirInput() meteo.AirInput {
	return meteo.AirInput{
		TemperatureCelsius: in.AirTemperature,
		Pressure:           in.Pressure,
		Elevation:          in.Elevation,
		LatentHeat:         in.LatentHeat,
		Humidity:           in.Humidity(),
	}
}

// AvailableEnergy returns Rn-G in MJ/(m2 period).
func (in Input) AvailableEnergy() float64 {
	return in.NetRadiation - in.SoilHeatFlux
}

// WithWindSpeed returns a copy of the reading at another 2 m wind speed.
func (in Input) WithWindSpeed(speed float64) Input {
	clone := in
	clone.WindSpeed = speed
	clone.WindHeight = nil
	return clone
}

// Calm returns the no-wind counterpart of the reading.
func (in Input) Calm() Input {
	return in.WithWindSpeed(calmWindSpeed)
}

// WithActualVaporPressure returns a copy whose humidity is pinned to an
// actual vapour pressure, dropping the other humidity channels.
func (in Input) WithActualVaporPressure(ea float64) Input {
	clone := in
	value := ea
	clone.ActualVaporPressure = &value
	clone.RelativeHumidity = nil
	clone.DewpointTemperature = nil
	return clone
}

// Temperature returns the air temperature as a unit-tagged reading.
func (in Input) Temperature() meteo.Temperature {
	return meteo.Celsius(in.AirTemperature)
}
