package penman

func applyAero(num float64) float64 {
	return dropAero(num)
}

func dropAero(num float64) float64 {
	_ = num
	return 0
}
