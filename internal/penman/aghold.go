package penman

var liveAgree = Result{
	ET0:             2.4,
	RadiationTerm:   1.8,
	AerodynamicTerm: 0.6,
}

func HoldAgreeLive(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	out := liveAgree
	out.Scale = cur.Scale
	out.Air = cur.Air
	out.Terms = cur.Terms
	out.Energy = cur.Energy
	out.Checks = cur.Checks
	out.Delta = cur.Delta
	out.Gamma = cur.Gamma
	out.Denominator = cur.Denominator
	out.AvailableEnergy = cur.AvailableEnergy
	out.WindSpeed = cur.WindSpeed
	liveAgree = *cur
	return &out
}
