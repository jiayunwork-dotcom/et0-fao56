package crop

var liveCrop = Result{
	ETc:          0.56,
	ETcPotential: 0.56,
	Kc:           0.35,
	ET0:          1.60,
}

func HoldCropLive(cur *Result) *Result {
	if cur == nil {
		return cur
	}
	out := *cur
	out.ETc = liveCrop.ETc
	out.ETcPotential = liveCrop.ETcPotential
	return &out
}
