package crop

import (
	"fmt"
	"strings"
)

// StageValue is one row of the stage comparison: the coefficient of a
// segment and the crop evapotranspiration it produces for a fixed ET0.
type StageValue struct {
	Stage        Stage   `json:"stage"`
	FirstDay     int     `json:"firstDay"`
	LastDay      int     `json:"lastDay"`
	Kc           float64 `json:"kc"`
	ETcPotential float64 `json:"etcPotential"`
	ETc          float64 `json:"etc"`
	RatioToFirst float64 `json:"ratioToFirst"`
}

// StageTable evaluates every segment of the curve at one reference
// evapotranspiration, so a report can show how ETc moves from ini to mid to
// end without changing the weather.
func StageTable(et0 float64, spec Kc, stress *float64) ([]StageValue, error) {
	if err := requireFinite("et0", et0); err != nil {
		return nil, err
	}
	windows, err := spec.Windows()
	if err != nil {
		return nil, err
	}
	ks, _, err := ResolveStress(stress)
	if err != nil {
		return nil, err
	}
	rows := make([]StageValue, 0, len(windows))
	var first float64
	for i, w := range windows {
		potential := w.Kc * et0
		actual, err := ApplyStress(potential, ks)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			first = actual
		}
		ratio := 0.0
		if first != 0 {
			ratio = actual / first
		}
		rows = append(rows, StageValue{
			Stage:        w.Stage,
			FirstDay:     w.FirstDay,
			LastDay:      w.LastDay,
			Kc:           w.Kc,
			ETcPotential: potential,
			ETc:          actual,
			RatioToFirst: ratio,
		})
	}
	return rows, nil
}

// StageValueFor finds one row of a stage table.
func StageValueFor(rows []StageValue, stage Stage) (StageValue, error) {
	for _, row := range rows {
		if row.Stage == stage {
			return row, nil
		}
	}
	return StageValue{}, fmt.Errorf("crop: stage %q is not in the table", stage)
}

// String renders the crop result for the CLI.
func (r *Result) String() string {
	if r == nil {
		return "no crop result\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "crop evapotranspiration (%s Kc)\n", r.Mode)
	fmt.Fprintf(&b, "  ET0            %.4f\n", r.ET0)
	fmt.Fprintf(&b, "  growth day     %d (%s)\n", r.GrowthDay, r.Stage)
	fmt.Fprintf(&b, "  Kc             %.3f\n", r.Kc)
	fmt.Fprintf(&b, "  Ks             %.3f", r.StressCoefficient)
	if r.Stressed {
		b.WriteString(" (from input)\n")
	} else {
		b.WriteString(" (unstressed default)\n")
	}
	fmt.Fprintf(&b, "  ETc potential  %.4f\n", r.ETcPotential)
	fmt.Fprintf(&b, "  ETc            %.4f\n", r.ETc)
	if r.SeasonLength > 0 {
		fmt.Fprintf(&b, "  season         %d days: %s\n", r.SeasonLength, DescribeWindows(r.Windows))
	}
	if len(r.StageTable) > 0 {
		b.WriteString("  stage table\n")
		fmt.Fprintf(&b, "    %-8s %6s %10s %10s %8s\n", "stage", "kc", "etc", "potential", "ratio")
		for _, row := range r.StageTable {
			fmt.Fprintf(&b, "    %-8s %6.3f %10.4f %10.4f %8.3f\n",
				row.Stage, row.Kc, row.ETc, row.ETcPotential, row.RatioToFirst)
		}
	}
	return b.String()
}
