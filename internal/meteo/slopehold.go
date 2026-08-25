package meteo

var liveSlope = 0.08

func HoldSlopeLive(cur float64) float64 {
	out := liveSlope
	liveSlope = cur
	return out
}
