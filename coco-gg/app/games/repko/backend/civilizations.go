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
	{ID: "rome", Name: "Rome", Color: "#b91c1c", BackgroundColor: "#fecaca", Flag: "🏛️", CoatOfArms: "🦅", StartingLoadout: loadout(40, 5, 5)},
	{ID: "egypt", Name: "Egypt", Color: "#eab308", BackgroundColor: "#fef08a", Flag: "🇪🇬", CoatOfArms: "☥", StartingLoadout: loadout(30, 15, 5)},
	{ID: "persia", Name: "Persia", Color: "#7c3aed", BackgroundColor: "#ddd6fe", Flag: "🇮🇷", CoatOfArms: "🦁", StartingLoadout: loadout(25, 15, 10)},
	{ID: "greece", Name: "Greece", Color: "#2563eb", BackgroundColor: "#bfdbfe", Flag: "🇬🇷", CoatOfArms: "🛡️", StartingLoadout: loadout(35, 5, 10)},
	{ID: "china", Name: "China", Color: "#dc2626", BackgroundColor: "#fecaca", Flag: "🇨🇳", CoatOfArms: "🐉", StartingLoadout: loadout(30, 10, 10)},
	{ID: "india", Name: "India", Color: "#f97316", BackgroundColor: "#fed7aa", Flag: "🇮🇳", CoatOfArms: "🐘", StartingLoadout: loadout(25, 15, 10)},
	{ID: "japan", Name: "Japan", Color: "#f43f5e", BackgroundColor: "#fecdd3", Flag: "🇯🇵", CoatOfArms: "⛩️", StartingLoadout: loadout(35, 10, 5)},
	{ID: "mongols", Name: "Mongols", Color: "#78350f", BackgroundColor: "#fde68a", Flag: "🐎", CoatOfArms: "🏹", StartingLoadout: loadout(10, 35, 5)},
	{ID: "vikings", Name: "Vikings", Color: "#0f766e", BackgroundColor: "#99f6e4", Flag: "⚔️", CoatOfArms: "⚒️", StartingLoadout: loadout(30, 15, 5)},
	{ID: "aztec", Name: "Aztec", Color: "#15803d", BackgroundColor: "#bbf7d0", Flag: "🦅", CoatOfArms: "🦅", StartingLoadout: loadout(35, 5, 10)},
	{ID: "maya", Name: "Maya", Color: "#65a30d", BackgroundColor: "#d9f99d", Flag: "🐍", CoatOfArms: "🐍", StartingLoadout: loadout(30, 5, 15)},
	{ID: "britain", Name: "Britain", Color: "#1d4ed8", BackgroundColor: "#bfdbfe", Flag: "🇬🇧", CoatOfArms: "🌹", StartingLoadout: loadout(25, 10, 15)},
	{ID: "france", Name: "France", Color: "#3b82f6", BackgroundColor: "#dbeafe", Flag: "🇫🇷", CoatOfArms: "⚜️", StartingLoadout: loadout(25, 15, 10)},
	{ID: "spain", Name: "Spain", Color: "#facc15", BackgroundColor: "#fef9c3", Flag: "🇪🇸", CoatOfArms: "🐂", StartingLoadout: loadout(25, 15, 10)},
	{ID: "germany", Name: "Germany", Color: "#374151", BackgroundColor: "#e5e7eb", Flag: "🇩🇪", CoatOfArms: "🐺", StartingLoadout: loadout(30, 10, 10)},
	{ID: "russia", Name: "Russia", Color: "#0ea5e9", BackgroundColor: "#e0f2fe", Flag: "🇷🇺", CoatOfArms: "🐻", StartingLoadout: loadout(35, 10, 5)},
	{ID: "ottoman", Name: "Ottoman", Color: "#16a34a", BackgroundColor: "#bbf7d0", Flag: "☪️", CoatOfArms: "🌙", StartingLoadout: loadout(15, 15, 20)},
	{ID: "ethiopia", Name: "Ethiopia", Color: "#854d0e", BackgroundColor: "#fde68a", Flag: "🇪🇹", CoatOfArms: "⛰️", StartingLoadout: loadout(30, 15, 5)},
	{ID: "mali", Name: "Mali", Color: "#d97706", BackgroundColor: "#fed7aa", Flag: "🦁", CoatOfArms: "🐪", StartingLoadout: loadout(25, 20, 5)},
	{ID: "inuit", Name: "Inuit", Color: "#94a3b8", BackgroundColor: "#e2e8f0", Flag: "❄️", CoatOfArms: "🐻‍❄️", StartingLoadout: loadout(40, 5, 5)},
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
