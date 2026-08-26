package penman

import (
	"errors"
	"math"
	"testing"
)

func fptr(v float64) *float64 { return &v }

func dailyInput() Input {
	return Input{
		NetRadiation:     8,
		SoilHeatFlux:     0,
		AirTemperature:   25,
		WindSpeed:        5,
		RelativeHumidity: fptr(20),
		Elevation:        fptr(100),
	}
}

func TestET0MatchesDocumentedFormula(t *testing.T) {
	res, err := Compute(dailyInput(), ScaleDaily)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	weight := 900.0 / (25.0 + 273.0)
	denom := res.Delta + res.Gamma*(1+0.34*res.WindSpeed)
	if math.Abs(denom-res.Denominator) > 1e-9 {
		t.Errorf("denominator = %.6f, expected %.6f", res.Denominator, denom)
	}
	numerator := 0.408*res.Delta*8.0 + res.Gamma*weight*res.WindSpeed*res.Deficit
	expected := numerator / denom
	if math.Abs(res.ET0-expected)/expected > 1e-9 {
		t.Errorf("ET0 = %.6f, expected %.6f", res.ET0, expected)
	}
	expectedRadiation := 0.408 * res.Delta * 8.0 / denom
	expectedAerodynamic := res.Gamma * weight * res.WindSpeed * res.Deficit / denom
	if math.Abs(res.RadiationTerm-expectedRadiation) > 1e-9 {
		t.Errorf("radiation term = %.6f, expected %.6f", res.RadiationTerm, expectedRadiation)
	}
	if math.Abs(res.AerodynamicTerm-expectedAerodynamic) > 1e-9 {
		t.Errorf("aerodynamic term = %.6f, expected %.6f", res.AerodynamicTerm, expectedAerodynamic)
	}
	rebuild, err := res.Rebuild()
	if err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if math.Abs(rebuild-res.ET0) > 1e-9 {
		t.Errorf("Rebuild = %.6f, expected %.6f", rebuild, res.ET0)
	}
}

func TestCalmDegeneration(t *testing.T) {
	calm, err := Compute(dailyInput().WithWindSpeed(0), ScaleDaily)
	if err != nil {
		t.Fatalf("Compute(calm): %v", err)
	}
	expectedDenom := calm.Delta + calm.Gamma
	if math.Abs(calm.Denominator-expectedDenom) > 1e-9 {
		t.Errorf("calm denominator = %.6f, expected delta+gamma=%.6f", calm.Denominator, expectedDenom)
	}
	if calm.AerodynamicTerm != 0 {
		t.Errorf("calm aerodynamic term = %v, expected 0", calm.AerodynamicTerm)
	}
	expected := 0.408 * calm.Delta * calm.AvailableEnergy / expectedDenom
	if math.Abs(calm.ET0-expected)/expected > 1e-9 {
		t.Errorf("calm ET0 = %.6f, expected radiation-only %.6f", calm.ET0, expected)
	}
	if calm.RadiationTerm <= 0 {
		t.Errorf("calm radiation term = %v, expected positive", calm.RadiationTerm)
	}
	windy, err := Compute(dailyInput(), ScaleDaily)
	if err != nil {
		t.Fatalf("Compute(windy): %v", err)
	}
	if windy.ET0 <= calm.ET0 {
		t.Errorf("windy ET0 = %.4f should exceed calm ET0 = %.4f", windy.ET0, calm.ET0)
	}
}

func TestDeficitMonotone(t *testing.T) {
	points, err := DeficitSweep(dailyInput(), ScaleDaily, []float64{0.2, 0.5, 1.0, 2.0})
	if err != nil {
		t.Fatalf("DeficitSweep: %v", err)
	}
	if err := CheckDeficitMonotone(points); err != nil {
		t.Errorf("ET0 fell while the deficit grew: %v", err)
	}
	for i := 1; i < len(points); i++ {
		if points[i].ET0 < points[i-1].ET0 {
			t.Errorf("sample %d ET0 = %.4f below previous %.4f", i, points[i].ET0, points[i-1].ET0)
		}
	}
}

func TestEnergyReconciliation(t *testing.T) {
	res, err := Compute(dailyInput(), ScaleDaily)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	lambda := 1.0 / 0.408
	flux := res.ET0 * lambda
	expected := res.Delta*res.AvailableEnergy/res.Denominator +
		lambda*res.Terms.AerodynamicNumerator/res.Denominator
	if math.Abs(flux-expected)/math.Max(1, math.Abs(expected)) > 1e-9 {
		t.Errorf("lambda*ET0 = %.6f, expected %.6f", flux, expected)
	}
	if !res.Energy.Reconciled {
		t.Errorf("energy audit reports residual %.3e", res.Energy.Residual)
	}
}

func TestNegativeWindRejected(t *testing.T) {
	in := dailyInput()
	in.WindSpeed = -0.5
	if _, err := Compute(in, ScaleDaily); err == nil {
		t.Error("Compute accepted a negative wind speed")
	}
	if err := ValidateWindSpeed(-0.1); err == nil {
		t.Error("ValidateWindSpeed accepted a negative speed")
	}
}

func TestWindHeightConversion(t *testing.T) {
	converted, err := WindAtTwoMetres(3, fptr(10))
	if err != nil {
		t.Fatalf("WindAtTwoMetres: %v", err)
	}
	if converted >= 3 {
		t.Errorf("u2 from 10 m = %.4f should be below the measured 3 m/s", converted)
	}
	if converted <= 0 {
		t.Errorf("u2 from 10 m = %.4f should stay positive", converted)
	}
	atTwo, err := WindAtTwoMetres(3, fptr(2))
	if err != nil {
		t.Fatalf("WindAtTwoMetres: %v", err)
	}
	if math.Abs(atTwo-3) > 1e-12 {
		t.Errorf("u2 at 2 m = %.4f, expected unchanged 3", atTwo)
	}
}

func TestClaimTranspiredContradiction(t *testing.T) {
	in := dailyInput()
	in.NetRadiation = -8
	in.WindSpeed = 0
	in.ClaimTranspired = true
	if _, err := Compute(in, ScaleDaily); !errors.Is(err, ErrSignContradiction) {
		t.Errorf("want ErrSignContradiction, got %v", err)
	}
	unclaimed := in
	unclaimed.ClaimTranspired = false
	if _, err := Compute(unclaimed, ScaleDaily); err != nil {
		t.Errorf("negative day without the claim should still compute: %v", err)
	}
}

func TestWindSweepSorts(t *testing.T) {
	points, err := WindSweep(dailyInput(), ScaleDaily, []float64{5, 1, 3})
	if err != nil {
		t.Fatalf("WindSweep: %v", err)
	}
	want := []float64{1, 3, 5}
	for i, speed := range want {
		if math.Abs(points[i].WindSpeed-speed) > 1e-12 {
			t.Errorf("sample %d wind speed = %.2f, expected %.2f", i, points[i].WindSpeed, speed)
		}
	}
}
