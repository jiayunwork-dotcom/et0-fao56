package meteo

import (
	"math"
	"testing"
)

// TestSlopeMatchesNumericDerivative checks that Delta, the analytic slope
// of the Tetens curve, agrees with a central difference of the same es(T).
func TestSlopeMatchesNumericDerivative(t *testing.T) {
	temperature := Celsius(25)
	analytic, err := Slope(temperature)
	if err != nil {
		t.Fatalf("Slope: %v", err)
	}
	numeric, err := SlopeNumeric(temperature, 0.025)
	if err != nil {
		t.Fatalf("SlopeNumeric: %v", err)
	}
	expected := SlopeNumerator * 0.6108 * math.Exp(TetensB*25/(25+TetensC)) / math.Pow(25+TetensC, 2)
	if math.Abs(analytic-expected)/expected > 1e-6 {
		t.Errorf("analytic delta = %.6f, expected %.6f", analytic, expected)
	}
	if math.Abs(analytic-numeric)/analytic > SlopeAgreementTolerance {
		t.Errorf("analytic delta = %.6f, numeric derivative = %.6f", analytic, numeric)
	}
	check, err := CheckSlope(temperature)
	if err != nil {
		t.Fatalf("CheckSlope: %v", err)
	}
	if !check.Agrees {
		t.Errorf("delta does not agree with a numeric derivative of es: gap %.3e", check.RelativeGap)
	}
}

// TestPsychrometricConstantMatchesFAO checks that gamma at sea level and
// the reference latent heat reproduces the FAO-56 factor 0.000665*P.
func TestPsychrometricConstantMatchesFAO(t *testing.T) {
	gamma, err := PsychrometricConstant(SeaLevelPressure, LatentHeatReference)
	if err != nil {
		t.Fatalf("PsychrometricConstant: %v", err)
	}
	expected := 0.000665 * SeaLevelPressure
	if math.Abs(gamma-expected)/expected > 1e-3 {
		t.Errorf("gamma = %.6f kPa/degC, expected %.6f", gamma, expected)
	}
	factor, err := PsychrometricFactor(LatentHeatReference)
	if err != nil {
		t.Fatalf("PsychrometricFactor: %v", err)
	}
	if math.Abs(factor-0.000665)/0.000665 > 1e-3 {
		t.Errorf("psychrometric factor = %.6f, expected 0.000665", factor)
	}
}

// TestActualFromRelativeHumidity checks the ea = es*RH/100 identity.
func TestActualFromRelativeHumidity(t *testing.T) {
	temperature := Celsius(25)
	es, err := SaturationVaporPressure(temperature)
	if err != nil {
		t.Fatalf("SaturationVaporPressure: %v", err)
	}
	ea, err := ActualFromRelativeHumidity(temperature, 40)
	if err != nil {
		t.Fatalf("ActualFromRelativeHumidity: %v", err)
	}
	expected := es * 0.4
	if math.Abs(ea-expected) > 1e-9 {
		t.Errorf("ea = %.6f, expected %.6f", ea, expected)
	}
}

// TestDeficitFromHumidity checks that the deficit is es-ea and that the
// round trip through ActualForDeficit reproduces it.
func TestDeficitFromHumidity(t *testing.T) {
	temperature := Celsius(30)
	es, err := SaturationVaporPressure(temperature)
	if err != nil {
		t.Fatalf("SaturationVaporPressure: %v", err)
	}
	ea, err := ActualFromRelativeHumidity(temperature, 30)
	if err != nil {
		t.Fatalf("ActualFromRelativeHumidity: %v", err)
	}
	deficit, err := Deficit(es, ea)
	if err != nil {
		t.Fatalf("Deficit: %v", err)
	}
	if math.Abs(deficit-(es-ea)) > 1e-9 {
		t.Errorf("deficit = %.6f, expected %.6f", deficit, es-ea)
	}
	ea2, err := ActualForDeficit(temperature, deficit)
	if err != nil {
		t.Fatalf("ActualForDeficit: %v", err)
	}
	if math.Abs(ea2-ea) > 1e-9 {
		t.Errorf("round-trip ea = %.6f, expected %.6f", ea2, ea)
	}
}

// TestRelativeHumidityOutOfRange checks that humidity outside [0, 100]
// percent is rejected.
func TestRelativeHumidityOutOfRange(t *testing.T) {
	for _, rh := range []float64{-1, 100.5, math.NaN(), math.Inf(1)} {
		if err := ValidateRelativeHumidity(rh); err == nil {
			t.Errorf("ValidateRelativeHumidity(%v) returned no error", rh)
		}
	}
	if err := ValidateRelativeHumidity(0); err != nil {
		t.Errorf("ValidateRelativeHumidity(0): %v", err)
	}
	if err := ValidateRelativeHumidity(100); err != nil {
		t.Errorf("ValidateRelativeHumidity(100): %v", err)
	}
}

// TestSupersaturatedDeficitRejected checks that an actual vapour pressure
// above the saturation pressure is an error, not a clamped value.
func TestSupersaturatedDeficitRejected(t *testing.T) {
	temperature := Celsius(20)
	es, err := SaturationVaporPressure(temperature)
	if err != nil {
		t.Fatalf("SaturationVaporPressure: %v", err)
	}
	if _, err := Deficit(es, es+0.1); err == nil {
		t.Error("Deficit accepted ea above es")
	}
	if _, err := ActualFromRelativeHumidity(temperature, 100.5); err == nil {
		t.Error("ActualFromRelativeHumidity accepted RH above 100")
	}
}

// TestDewpointRoundTrip checks that the Tetens curve inverts: the dewpoint
// of es(dew) reproduces the input temperature.
func TestDewpointRoundTrip(t *testing.T) {
	dewIn := Celsius(12)
	ea, err := ActualFromDewpoint(dewIn)
	if err != nil {
		t.Fatalf("ActualFromDewpoint: %v", err)
	}
	dewOut, err := DewpointFromActual(ea)
	if err != nil {
		t.Fatalf("DewpointFromActual: %v", err)
	}
	if math.Abs(dewOut.Celsius()-12) > 1e-6 {
		t.Errorf("dewpoint = %.6f, expected 12", dewOut.Celsius())
	}
}

// TestPressureAtElevation checks the barometric law endpoints.
func TestPressureAtElevation(t *testing.T) {
	seaLevel, err := PressureAtElevation(0)
	if err != nil {
		t.Fatalf("PressureAtElevation(0): %v", err)
	}
	if math.Abs(seaLevel-SeaLevelPressure) > 1e-9 {
		t.Errorf("pressure at sea level = %.4f, expected %.4f", seaLevel, SeaLevelPressure)
	}
	high, err := PressureAtElevation(1500)
	if err != nil {
		t.Fatalf("PressureAtElevation(1500): %v", err)
	}
	if high >= seaLevel {
		t.Errorf("pressure at 1500 m = %.4f should be below sea level %.4f", high, seaLevel)
	}
	if _, err := PressureAtElevation(-500); err == nil {
		t.Error("PressureAtElevation accepted an elevation below the model range")
	}
}

// TestLatentHeatRejectsNonPositive checks that gamma needs positive
// pressure and latent heat.
func TestLatentHeatRejectsNonPositive(t *testing.T) {
	if _, err := PsychrometricConstant(0, LatentHeatReference); err == nil {
		t.Error("PsychrometricConstant accepted zero pressure")
	}
	if _, err := PsychrometricConstant(101.3, 0); err == nil {
		t.Error("PsychrometricConstant accepted zero latent heat")
	}
	if err := ValidatePressure(-1); err == nil {
		t.Error("ValidatePressure accepted a negative pressure")
	}
}

// TestLatentHeatAtTemperature checks the temperature dependent latent heat.
func TestLatentHeatAtTemperature(t *testing.T) {
	lambda, err := LatentHeatAt(Celsius(25))
	if err != nil {
		t.Fatalf("LatentHeatAt: %v", err)
	}
	expected := LatentHeatIntercept - LatentHeatSlope*25
	if math.Abs(lambda-expected) > 1e-9 {
		t.Errorf("lambda at 25 degC = %.6f, expected %.6f", lambda, expected)
	}
}

// TestDeriveConsistency checks that one air reading yields Delta, es, ea
// and gamma from the same temperature and pressure.
func TestDeriveConsistency(t *testing.T) {
	pressure := 95.0
	rh := 45.0
	air, err := Derive(AirInput{
		TemperatureCelsius: 22,
		Pressure:           &pressure,
		Humidity:           Humidity{RelativeHumidity: &rh},
	})
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}
	es, err := SaturationVaporPressure(air.Temperature)
	if err != nil {
		t.Fatalf("SaturationVaporPressure: %v", err)
	}
	if math.Abs(air.SaturationVaporPressure-es) > 1e-9 {
		t.Errorf("es = %.6f, expected %.6f", air.SaturationVaporPressure, es)
	}
	if math.Abs(air.Deficit-(es-air.ActualVaporPressure)) > 1e-9 {
		t.Errorf("deficit = %.6f, expected %.6f", air.Deficit, es-air.ActualVaporPressure)
	}
	expectedGamma, err := PsychrometricConstant(pressure, LatentHeatReference)
	if err != nil {
		t.Fatalf("PsychrometricConstant: %v", err)
	}
	if math.Abs(air.Gamma-expectedGamma) > 1e-9 {
		t.Errorf("gamma = %.6f, expected %.6f", air.Gamma, expectedGamma)
	}
}
