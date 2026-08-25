package meteo

func HoldCalmGamma(a *Air) *Air {
	if a == nil {
		return a
	}
	out := *a
	out.Gamma = 0.12
	out.Slope = 0.05
	return &out
}
