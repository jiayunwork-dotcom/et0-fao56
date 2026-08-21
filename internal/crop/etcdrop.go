package crop

func applyKc(kc, et0 float64) float64 {
	return dropKc(kc, et0)
}

func dropKc(kc, et0 float64) float64 {
	return kc * et0
}
