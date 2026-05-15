package repko

import "sort"

func resolveCombat(attackers, defenders []Unit) (attackerWins bool, neutralResult bool, survivors []Unit) {
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

func totalPower(units []Unit) int {
	total := 0
	for _, u := range units {
		total += unitPower(u)
	}
	return total
}

// removePower destroys whole units from `units` in lowest-power-first order until
// the cumulative destroyed power meets or exceeds `power`. The last unit whose
// destruction crosses the threshold is destroyed entirely (no fractional damage).
// Returns the survivors in the original input order.
func removePower(units []Unit, power int) []Unit {
	if power <= 0 || len(units) == 0 {
		return append([]Unit(nil), units...)
	}

	type indexed struct {
		idx int
		u   Unit
	}
	sorted := make([]indexed, len(units))
	for i, u := range units {
		sorted[i] = indexed{idx: i, u: u}
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		pi := unitPower(sorted[i].u)
		pj := unitPower(sorted[j].u)
		if pi != pj {
			return pi < pj
		}
		if sorted[i].u.Type != sorted[j].u.Type {
			return sorted[i].u.Type < sorted[j].u.Type
		}
		return sorted[i].u.Level < sorted[j].u.Level
	})

	destroyed := make(map[int]struct{}, len(sorted))
	accumulated := 0
	for _, s := range sorted {
		if accumulated >= power {
			break
		}
		destroyed[s.idx] = struct{}{}
		accumulated += unitPower(s.u)
	}

	survivors := make([]Unit, 0, len(units))
	for i, u := range units {
		if _, gone := destroyed[i]; gone {
			continue
		}
		survivors = append(survivors, u)
	}
	return survivors
}
