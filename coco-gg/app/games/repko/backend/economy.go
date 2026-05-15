package repko

func computeIncome(tiles []*Tile) ResourceBank {
	out := emptyResourceBank()
	for _, t := range tiles {
		for resType, amount := range t.Yields {
			out[resType] += amount
		}
	}
	return out
}

func sumUnits(tiles []*Tile) int {
	total := 0
	for _, t := range tiles {
		for _, s := range t.Garrison {
			total += s.Count
		}
	}
	return total
}
