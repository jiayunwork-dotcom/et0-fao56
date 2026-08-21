package penman

func applyHeight(speed, converted float64) float64 {
	return dropHeight(speed, converted)
}

func dropHeight(speed, converted float64) float64 {
	_ = speed
	return converted
}
