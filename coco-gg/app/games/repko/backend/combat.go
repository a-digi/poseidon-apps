package repko

import "sort"

func resolveCombat(attackers, defenders []GarrisonStack) (attackerWins bool, neutralResult bool, survivors []GarrisonStack) {
	attackerPower := totalPower(attackers)
	defenderPower := totalPower(defenders)

	if attackerPower == defenderPower {
		return false, true, nil
	}
	if attackerPower > defenderPower {
		return true, false, removePower(attackers, defenderPower)
	}
	return false, false, removePower(defenders, attackerPower)
}

func totalPower(stacks []GarrisonStack) int {
	total := 0
	for _, s := range stacks {
		total += stackPower(s) * s.Count
	}
	return total
}

func stackPower(s GarrisonStack) int {
	return basePower(s.Type) * int(s.Level)
}

func basePower(t UnitType) int {
	spec, ok := UnitCatalog[t]
	if !ok {
		return 0
	}
	return spec.Power
}

// removePower destroys whole soldiers from `stacks` in lowest-power-first order
// until the cumulative destroyed power meets or exceeds `target`. The last
// soldier whose destruction crosses the threshold is destroyed entirely (no
// fractional damage). Returns the survivors in the original input order.
func removePower(stacks []GarrisonStack, target int) []GarrisonStack {
	work := make([]GarrisonStack, len(stacks))
	copy(work, stacks)
	if target <= 0 || len(work) == 0 {
		return filterNonEmpty(work)
	}

	indices := make([]int, len(work))
	for i := range work {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool {
		a, b := work[indices[i]], work[indices[j]]
		pa, pb := stackPower(a), stackPower(b)
		if pa != pb {
			return pa < pb
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		return a.Level < b.Level
	})

	for _, idx := range indices {
		if target <= 0 {
			break
		}
		s := &work[idx]
		power := stackPower(*s)
		if power <= 0 || s.Count == 0 {
			continue
		}
		removable := target / power
		if removable > s.Count {
			removable = s.Count
		}
		if removable == 0 {
			continue
		}
		s.Count -= removable
		target -= removable * power
	}

	if target > 0 {
		for _, idx := range indices {
			if work[idx].Count > 0 {
				work[idx].Count--
				target = 0
				break
			}
		}
	}

	return filterNonEmpty(work)
}

func filterNonEmpty(stacks []GarrisonStack) []GarrisonStack {
	out := make([]GarrisonStack, 0, len(stacks))
	for _, s := range stacks {
		if s.Count > 0 {
			out = append(out, s)
		}
	}
	return out
}
