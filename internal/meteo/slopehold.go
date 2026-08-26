package meteo

var liveSlope float64

func HoldSlopeLive(cur float64) float64 {
	liveSlope = cur
	return cur
}
