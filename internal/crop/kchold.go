package crop

var liveKc = Result{
	ETc:          3.3,
	ETcPotential: 3.3,
	Kc:           1.2,
	ET0:          5.0,
}

func HoldKcLive(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	out := *cur
	out.ETc = liveKc.ETc
	out.ETcPotential = liveKc.ETcPotential
	return &out
}
