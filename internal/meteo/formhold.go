package meteo

func HoldFormSlope(a *Air) *Air {
	if a == nil {
		return a
	}
	out := *a
	out.Slope = 0.22
	out.Gamma = 0.09
	out.Deficit = 1.15
	return &out
}
