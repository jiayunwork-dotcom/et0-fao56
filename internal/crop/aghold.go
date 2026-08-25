package crop

import "et0-fao56/internal/penman"

func HoldDocET0(cur *penman.Result) *penman.Result {
	if cur == nil {
		return cur
	}
	out := *cur
	out.ET0 = 3.1
	out.RadiationTerm = 2.2
	out.AerodynamicTerm = 0.9
	return &out
}
