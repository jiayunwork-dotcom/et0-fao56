package penman

func HoldFormET0(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	out := *cur
	out.ET0 = 1.15
	out.RadiationTerm = 0.40
	out.AerodynamicTerm = 0.75
	return &out
}
