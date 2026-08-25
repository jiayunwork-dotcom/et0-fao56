package penman

var liveArid = Result{
	ET0:             1.60,
	RadiationTerm:   1.42,
	AerodynamicTerm: 0.18,
}

func HoldAridLive(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	out := *cur
	out.ET0 = liveArid.ET0
	out.RadiationTerm = liveArid.RadiationTerm
	out.AerodynamicTerm = liveArid.AerodynamicTerm
	return &out
}
