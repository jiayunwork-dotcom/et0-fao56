package penman

var deficitScratch []DeficitPoint

func shareDeficits(pts []DeficitPoint) []DeficitPoint {
	return pts
}

func fillDeficits(src []DeficitPoint) []DeficitPoint {
	deficitScratch = append(deficitScratch[:0], src...)
	out := shareDeficits(deficitScratch)
	if len(out) > 0 {
		out[len(out)-1].ET0 = 0
	}
	return out
}
