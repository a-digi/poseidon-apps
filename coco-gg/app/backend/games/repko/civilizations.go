package repko

func loadout(infantry, cavalry, artillery int) map[UnitType]int {
	return map[UnitType]int{
		UnitInfantry:  infantry,
		UnitCavalry:   cavalry,
		UnitArtillery: artillery,
	}
}

// asserts: every StartingLoadout sums to exactly 50.
var Civilizations = []Civilization{
	{ID: "rome", Name: "Rome", Color: "#b91c1c", Flag: "🏛️", StartingLoadout: loadout(40, 5, 5)},
	{ID: "egypt", Name: "Egypt", Color: "#eab308", Flag: "🇪🇬", StartingLoadout: loadout(30, 15, 5)},
	{ID: "persia", Name: "Persia", Color: "#7c3aed", Flag: "🇮🇷", StartingLoadout: loadout(25, 15, 10)},
	{ID: "greece", Name: "Greece", Color: "#2563eb", Flag: "🇬🇷", StartingLoadout: loadout(35, 5, 10)},
	{ID: "china", Name: "China", Color: "#dc2626", Flag: "🇨🇳", StartingLoadout: loadout(30, 10, 10)},
	{ID: "india", Name: "India", Color: "#f97316", Flag: "🇮🇳", StartingLoadout: loadout(25, 15, 10)},
	{ID: "japan", Name: "Japan", Color: "#f43f5e", Flag: "🇯🇵", StartingLoadout: loadout(35, 10, 5)},
	{ID: "mongols", Name: "Mongols", Color: "#78350f", Flag: "🐎", StartingLoadout: loadout(10, 35, 5)},
	{ID: "vikings", Name: "Vikings", Color: "#0f766e", Flag: "⚔️", StartingLoadout: loadout(30, 15, 5)},
	{ID: "aztec", Name: "Aztec", Color: "#15803d", Flag: "🦅", StartingLoadout: loadout(35, 5, 10)},
	{ID: "maya", Name: "Maya", Color: "#65a30d", Flag: "🐍", StartingLoadout: loadout(30, 5, 15)},
	{ID: "britain", Name: "Britain", Color: "#1d4ed8", Flag: "🇬🇧", StartingLoadout: loadout(25, 10, 15)},
	{ID: "france", Name: "France", Color: "#3b82f6", Flag: "🇫🇷", StartingLoadout: loadout(25, 15, 10)},
	{ID: "spain", Name: "Spain", Color: "#facc15", Flag: "🇪🇸", StartingLoadout: loadout(25, 15, 10)},
	{ID: "germany", Name: "Germany", Color: "#374151", Flag: "🇩🇪", StartingLoadout: loadout(30, 10, 10)},
	{ID: "russia", Name: "Russia", Color: "#0ea5e9", Flag: "🇷🇺", StartingLoadout: loadout(35, 10, 5)},
	{ID: "ottoman", Name: "Ottoman", Color: "#16a34a", Flag: "☪️", StartingLoadout: loadout(15, 15, 20)},
	{ID: "ethiopia", Name: "Ethiopia", Color: "#854d0e", Flag: "🇪🇹", StartingLoadout: loadout(30, 15, 5)},
	{ID: "mali", Name: "Mali", Color: "#d97706", Flag: "🦁", StartingLoadout: loadout(25, 20, 5)},
	{ID: "inuit", Name: "Inuit", Color: "#94a3b8", Flag: "❄️", StartingLoadout: loadout(40, 5, 5)},
}

func init() {
	for _, c := range Civilizations {
		sum := 0
		for _, n := range c.StartingLoadout {
			sum += n
		}
		if sum != 50 {
			panic("repko: civilization " + c.ID + " StartingLoadout does not sum to 50")
		}
	}
}
