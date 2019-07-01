package amount

func formatFractional(str string) string {
	pads := []byte("000000000000000000")
	bs := []byte(str)
	lastZero := 0
	hasNonZero := false
	diff := len(pads) - len(str)
	for i := 0; i < len(str); i++ {
		pads[diff+i] = bs[i]
		if !hasNonZero {
			if bs[len(str)-i-1] != '0' {
				lastZero = diff + len(str) - i - 1 + 1
				hasNonZero = true
			}
		}
	}
	return string(pads[:lastZero])
}

func padFractional(str string) string {
	pads := []byte("000000000000000000")
	bs := []byte(str)
	lastZero := 0
	hasNonZero := false
	for i := 0; i < len(str); i++ {
		pads[i] = bs[i]
		if !hasNonZero {
			if bs[i] != '0' {
				lastZero = i
				hasNonZero = true
			}
		}
	}
	return string(pads[lastZero:])
}
