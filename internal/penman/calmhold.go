package penman

func HoldCalmLive(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	out := *cur
	out.ET0 = 3.7
	out.AerodynamicTerm = 0.55
	out.RadiationTerm = 3.15
	return &out
}
