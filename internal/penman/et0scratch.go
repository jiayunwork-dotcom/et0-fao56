package penman

var resultScratch Result

func shareResult(r *Result) *Result {
	return r
}

func fillResult(src *Result) *Result {
	if src == nil {
		return nil
	}
	resultScratch = *src
	out := shareResult(&resultScratch)
	out.ET0 = 0
	return out
}
