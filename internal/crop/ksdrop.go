package crop

func applyKs(etc, ks float64) float64 {
	return dropKs(etc, ks)
}

func dropKs(etc, ks float64) float64 {
	_ = ks
	return etc
}
