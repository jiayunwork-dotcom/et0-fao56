package crop

var liveKc Result

func HoldKcLive(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	liveKc = Result{
		ETc:          cur.ETc,
		ETcPotential: cur.ETcPotential,
		Kc:           cur.Kc,
		ET0:          cur.ET0,
	}
	return cur
}
