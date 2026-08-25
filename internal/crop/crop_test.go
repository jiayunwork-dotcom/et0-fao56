package crop

import (
	"errors"
	"math"
	"testing"

	"et0-fao56/internal/penman"
)

func intPtr(v int) *int { return &v }

func dailyWeather() penman.Input {
	rh := 20.0
	elevation := 100.0
	return penman.Input{
		NetRadiation:     8,
		AirTemperature:   25,
		WindSpeed:        5,
		RelativeHumidity: &rh,
		Elevation:        &elevation,
	}
}

func TestSingleKcMultiplies(t *testing.T) {
	res, err := Evaluate(5.0, SingleKc(1.2), nil, nil)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if math.Abs(res.ETc-6.0) > 1e-9 {
		t.Errorf("ETc = %.4f, expected 6.0", res.ETc)
	}
	if math.Abs(res.ETcPotential-6.0) > 1e-9 {
		t.Errorf("ETc potential = %.4f, expected 6.0", res.ETcPotential)
	}
	if res.Stage != StageSingle {
		t.Errorf("stage = %q, expected single", res.Stage)
	}
}

func TestStagedKcSelectsStage(t *testing.T) {
	spec := StagedKc(0.4, 1.2, 0.6, 20, 60, 20)
	mid, err := Evaluate(5.0, spec, intPtr(45), nil)
	if err != nil {
		t.Fatalf("Evaluate(mid): %v", err)
	}
	if mid.Stage != StageMid {
		t.Errorf("stage = %q, expected mid", mid.Stage)
	}
	if math.Abs(mid.Kc-1.2) > 1e-9 {
		t.Errorf("Kc = %v, expected 1.2", mid.Kc)
	}
	if math.Abs(mid.ETc-6.0) > 1e-9 {
		t.Errorf("ETc = %.4f, expected 6.0", mid.ETc)
	}
	initial, err := Evaluate(5.0, spec, intPtr(10), nil)
	if err != nil {
		t.Fatalf("Evaluate(initial): %v", err)
	}
	if initial.Stage != StageInitial {
		t.Errorf("stage = %q, expected initial", initial.Stage)
	}
	if math.Abs(initial.ETc-2.0) > 1e-9 {
		t.Errorf("ETc = %.4f, expected 2.0", initial.ETc)
	}
}

func TestKcNonPositiveRejected(t *testing.T) {
	for _, value := range []float64{0, -1.5} {
		if _, err := Evaluate(5.0, SingleKc(value), nil, nil); err == nil {
			t.Errorf("Evaluate accepted Kc=%v", value)
		}
	}
	spec := StagedKc(0.4, 0, 0.6, 20, 60, 20)
	if _, err := Evaluate(5.0, spec, intPtr(45), nil); err == nil {
		t.Error("staged Kc accepted mid=0")
	}
	conflict := StagedKc(0.4, 1.2, 0.6, 20, 60, 20)
	conflict.Single = &conflictMid
	if _, err := Evaluate(5.0, conflict, intPtr(45), nil); err == nil {
		t.Error("Kc with both single and staged values was accepted")
	}
}

var conflictMid = 1.0

func TestStressScalesETc(t *testing.T) {
	unstressed, err := Evaluate(5.0, SingleKc(1.2), nil, nil)
	if err != nil {
		t.Fatalf("Evaluate(unstressed): %v", err)
	}
	ksOne := 1.0
	withOne, err := Evaluate(5.0, SingleKc(1.2), nil, &ksOne)
	if err != nil {
		t.Fatalf("Evaluate(Ks=1): %v", err)
	}
	if math.Abs(withOne.ETc-unstressed.ETc) > 1e-12 {
		t.Errorf("Ks=1 ETc = %.4f should equal unstressed %.4f", withOne.ETc, unstressed.ETc)
	}
	if unstressed.Stressed {
		t.Error("no stress coefficient supplied should leave the unstressed default flag")
	}
	if !withOne.Stressed {
		t.Error("an explicit Ks=1 should mark the result as stress-considered")
	}
	ksHalf := 0.5
	half, err := Evaluate(5.0, SingleKc(1.2), nil, &ksHalf)
	if err != nil {
		t.Fatalf("Evaluate(Ks=0.5): %v", err)
	}
	if math.Abs(half.ETc-3.0) > 1e-9 {
		t.Errorf("ETc = %.4f, expected 3.0", half.ETc)
	}
	if !half.Stressed {
		t.Error("Ks=0.5 should be flagged as stressed")
	}
	for _, invalid := range []float64{0, -1, 2} {
		if _, err := Evaluate(5.0, SingleKc(1.2), nil, fptrForTest(invalid)); err == nil {
			t.Errorf("Evaluate accepted Ks=%v", invalid)
		}
	}
}

func fptrForTest(v float64) *float64 { return &v }

func TestStageRatioProportional(t *testing.T) {
	spec := StagedKc(0.4, 1.2, 0.6, 20, 60, 20)
	rows, err := StageTable(5.0, spec, nil)
	if err != nil {
		t.Fatalf("StageTable: %v", err)
	}
	initial, err := StageValueFor(rows, StageInitial)
	if err != nil {
		t.Fatalf("StageValueFor(initial): %v", err)
	}
	mid, err := StageValueFor(rows, StageMid)
	if err != nil {
		t.Fatalf("StageValueFor(mid): %v", err)
	}
	ratio := 1.2 / 0.4
	if math.Abs(mid.RatioToFirst-ratio) > 1e-9 {
		t.Errorf("mid ratio = %.4f, expected %.4f", mid.RatioToFirst, ratio)
	}
	if math.Abs(mid.ETc-ratio*initial.ETc) > 1e-9 {
		t.Errorf("mid ETc = %.4f, expected %.4f", mid.ETc, ratio*initial.ETc)
	}
}

func TestGrowthDayErrors(t *testing.T) {
	spec := StagedKc(0.4, 1.2, 0.6, 20, 60, 20)
	if _, err := Evaluate(5.0, spec, nil, nil); !errors.Is(err, ErrGrowthDayRequired) {
		t.Errorf("want ErrGrowthDayRequired, got %v", err)
	}
	if _, err := Evaluate(5.0, spec, intPtr(200), nil); !errors.Is(err, ErrGrowthDayOutside) {
		t.Errorf("want ErrGrowthDayOutside, got %v", err)
	}
	if _, err := Evaluate(5.0, spec, intPtr(0), nil); !errors.Is(err, ErrGrowthDayInvalid) {
		t.Errorf("want ErrGrowthDayInvalid, got %v", err)
	}
}

func TestDocumentReferenceAgreement(t *testing.T) {
	weather := dailyWeather()
	res, err := penman.Compute(weather, penman.ScaleDaily)
	if err != nil {
		t.Fatalf("penman.Compute: %v", err)
	}
	et0 := res.ET0
	doc := &Document{Weather: &weather, ET0: &et0}
	if _, _, _, err := doc.Reference(); err != nil {
		t.Errorf("matching ET0 should agree with the weather: %v", err)
	}
	wrong := et0 + 0.5
	doc.ET0 = &wrong
	if _, _, _, err := doc.Reference(); !errors.Is(err, ErrReferenceConflict) {
		t.Errorf("want ErrReferenceConflict, got %v", err)
	}
}

func TestAridDayExampleAerodynamicDominates(t *testing.T) {
	doc, err := LoadFile("../../example/arid-day.json")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	out, err := doc.Evaluate()
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	ref := out.Reference
	if ref == nil {
		t.Fatal("the example produced no reference result")
	}
	if ref.ET0 <= 0 {
		t.Errorf("arid day ET0 = %.4f must be positive", ref.ET0)
	}
	if ref.AerodynamicTerm <= ref.RadiationTerm {
		t.Errorf("arid windy day aerodynamic term = %.4f should exceed the radiation term = %.4f",
			ref.AerodynamicTerm, ref.RadiationTerm)
	}
	actual, calm, err := penman.CalmComparison(*doc.Weather, penman.ScaleDaily)
	if err != nil {
		t.Fatalf("CalmComparison: %v", err)
	}
	if actual.ET0 <= calm.ET0 {
		t.Errorf("windy ET0 = %.4f should exceed the no-wind ET0 = %.4f", actual.ET0, calm.ET0)
	}
	if actual.AerodynamicTerm <= calm.AerodynamicTerm {
		t.Errorf("windy aerodynamic term should exceed the no-wind one")
	}
	if out.Crop == nil {
		t.Fatal("the example carried no crop block")
	}
	if math.Abs(out.Crop.ETc-out.Crop.Kc*out.ET0) > 1e-9 {
		t.Errorf("ETc = %.4f should equal Kc*ET0 = %.4f", out.Crop.ETc, out.Crop.Kc*out.ET0)
	}
}
