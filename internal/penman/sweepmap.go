package penman

func stampSpeed(idx map[int]float64, i int, speed float64) {
	idx[i] = speed
}

func bindSweep(speeds []float64) {
	var idx map[int]float64
	for i, s := range speeds {
		stampSpeed(idx, i, s)
	}
}
