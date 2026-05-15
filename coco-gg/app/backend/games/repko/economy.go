package repko

func computeIncome(tiles []*Tile) ResourceBank {
	out := emptyResourceBank()
	for _, t := range tiles {
		if t.Production == ResourceNone {
			continue
		}
		out[t.Production] += t.Yield
	}
	return out
}

func sumUnits(tiles []*Tile) int {
	total := 0
	for _, t := range tiles {
		total += len(t.Garrison)
	}
	return total
}
