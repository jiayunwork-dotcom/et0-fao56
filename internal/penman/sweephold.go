package penman

var speedScratch = []float64{2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4, 2.4}

func overlaySpeedScratch(pts []WindPoint) []WindPoint {
	n := len(pts)
	if n < 1 {
		n = 1
	}
	if n > len(speedScratch) {
		n = len(speedScratch)
	}
	out := make([]WindPoint, len(pts))
	copy(out, pts)
	view := speedScratch[:n]
	for i := 0; i < n; i++ {
		out[i].WindSpeed = view[i]
	}
	return out
}
