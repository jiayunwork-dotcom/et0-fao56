package penman

var speedScratch []float64

func overlaySpeedScratch(pts []WindPoint) []WindPoint {
	n := len(pts)
	if n < 1 {
		return pts
	}
	out := make([]WindPoint, len(pts))
	copy(out, pts)
	scratch := make([]float64, n)
	for i := 0; i < n; i++ {
		scratch[i] = out[i].WindSpeed
	}
	speedScratch = scratch
	return out
}
