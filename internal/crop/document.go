package crop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"et0-fao56/internal/penman"
)

const ReferenceAgreementTolerance = 1e-6

type Plan struct {
	Name              string   `json:"name,omitempty"`
	Kc                Kc       `json:"kc"`
	GrowthDay         *int     `json:"growthDay,omitempty"`
	StressCoefficient *float64 `json:"stressCoefficient,omitempty"`
}

type Document struct {
	Site      string           `json:"site,omitempty"`
	TimeScale penman.TimeScale `json:"timeScale,omitempty"`
	Weather   *penman.Input    `json:"weather,omitempty"`
	ET0       *float64         `json:"et0,omitempty"`
	WindSweep []float64        `json:"windSweep,omitempty"`
	Crop      *Plan            `json:"crop,omitempty"`
}

type Outcome struct {
	Site      string             `json:"site,omitempty"`
	Scale     penman.TimeScale   `json:"scale"`
	Reference *penman.Result     `json:"reference,omitempty"`
	ET0       float64            `json:"et0"`
	ET0Source string             `json:"et0Source"`
	Crop      *Result            `json:"crop,omitempty"`
	WindSweep []penman.WindPoint `json:"windSweep,omitempty"`
}

func LoadBytes(data []byte) (*Document, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, ErrDocumentEmpty
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var doc Document
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("crop: parse document: %w", err)
	}
	if err := doc.Validate(); err != nil {
		return nil, err
	}
	return &doc, nil
}

func LoadFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("crop: read %s: %w", path, err)
	}
	doc, err := LoadBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return doc, nil
}

func (d *Document) Validate() error {
	if d == nil {
		return ErrDocumentEmpty
	}
	if d.Weather == nil && d.ET0 == nil {
		return ErrReferenceMissing
	}
	if _, err := d.Scale(); err != nil {
		return err
	}
	if d.Weather != nil {
		if err := d.Weather.Validate(); err != nil {
			return err
		}
	}
	if d.ET0 != nil {
		if err := requireFinite("et0", *d.ET0); err != nil {
			return err
		}
	}
	for _, speed := range d.WindSweep {
		if err := penman.ValidateWindSpeed(speed); err != nil {
			return err
		}
	}
	if d.Crop != nil {
		if err := d.Crop.Kc.Validate(); err != nil {
			return err
		}
		if d.Crop.StressCoefficient != nil {
			if err := ValidateStress(*d.Crop.StressCoefficient); err != nil {
				return err
			}
		}
		if d.Crop.GrowthDay != nil && *d.Crop.GrowthDay < 1 {
			return fmt.Errorf("%w: got %d", ErrGrowthDayInvalid, *d.Crop.GrowthDay)
		}
	}
	return nil
}

func (d *Document) Scale() (penman.TimeScale, error) {
	coeffs, err := penman.CoefficientsFor(d.TimeScale)
	if err != nil {
		return "", err
	}
	return coeffs.Scale, nil
}

func (d *Document) Reference() (float64, string, *penman.Result, error) {
	if err := d.Validate(); err != nil {
		return 0, "", nil, err
	}
	scale, err := d.Scale()
	if err != nil {
		return 0, "", nil, err
	}
	if d.Weather == nil {
		return *d.ET0, "input", nil, nil
	}
	res, err := penman.Compute(*d.Weather, scale)
	if err != nil {
		return 0, "", nil, err
	}
	if d.ET0 != nil && math.Abs(*d.ET0-res.ET0) > ReferenceAgreementTolerance {
		return 0, "", nil, fmt.Errorf("%w: document gives ET0=%g but the weather block computes %g",
			ErrReferenceConflict, *d.ET0, res.ET0)
	}
	return res.ET0, "weather", res, nil
}

func (d *Document) Evaluate() (*Outcome, error) {
	et0, source, reference, err := d.Reference()
	if err != nil {
		return nil, err
	}
	scale, err := d.Scale()
	if err != nil {
		return nil, err
	}
	out := &Outcome{
		Site:      d.Site,
		Scale:     scale,
		Reference: reference,
		ET0:       et0,
		ET0Source: source,
	}
	if len(d.WindSweep) > 0 {
		if d.Weather == nil {
			return nil, fmt.Errorf("crop: a wind sweep needs the weather block")
		}
		sweep, err := penman.WindSweep(*d.Weather, scale, d.WindSweep)
		if err != nil {
			return nil, err
		}
		out.WindSweep = sweep
	}
	if d.Crop != nil {
		res, err := Evaluate(et0, d.Crop.Kc, d.Crop.GrowthDay, d.Crop.StressCoefficient)
		if err != nil {
			return nil, err
		}
		out.Crop = res
	}
	return out, nil
}

func (d *Document) EvaluateCrop() (*Outcome, error) {
	if d == nil {
		return nil, ErrDocumentEmpty
	}
	if d.Crop == nil {
		return nil, ErrPlanMissing
	}
	return d.Evaluate()
}
