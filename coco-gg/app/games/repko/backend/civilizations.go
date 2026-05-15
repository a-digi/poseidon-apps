package repko

// asserts: every StartingLoadout sums to exactly 50 and every key appears in UnitRoster.
var Civilizations = []Civilization{
	{
		ID: "rome", Name: "Rome", Color: "#b91c1c", BackgroundColor: "#fecaca", Flag: "🏛️", CoatOfArms: "🦅",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:   25,
			UnitMarines:    10,
			UnitLightTank:  8,
			UnitMediumTank: 5,
			UnitHowitzer:   2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitMarines, UnitLightTank, UnitMediumTank, UnitHowitzer, UnitHelicopter},
		IncomePercent: 100,
	},
	{
		ID: "egypt", Name: "Egypt", Color: "#eab308", BackgroundColor: "#fef08a", Flag: "🇪🇬", CoatOfArms: "☥",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:  30,
			UnitSnipers:   10,
			UnitEngineers: 5,
			UnitMortar:    3,
			UnitLightTank: 2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitSnipers, UnitEngineers, UnitMortar, UnitLightTank},
		IncomePercent: 125,
	},
	{
		ID: "persia", Name: "Persia", Color: "#7c3aed", BackgroundColor: "#ddd6fe", Flag: "🇮🇷", CoatOfArms: "🦁",
		StartingLoadout: map[UnitType]int{
			UnitMarines:    20,
			UnitSnipers:    15,
			UnitLightTank:  8,
			UnitMediumTank: 5,
			UnitHowitzer:   2,
		},
		UnitRoster:    []UnitType{UnitMarines, UnitSnipers, UnitLightTank, UnitMediumTank, UnitHowitzer, UnitHelicopter},
		IncomePercent: 100,
	},
	{
		ID: "greece", Name: "Greece", Color: "#2563eb", BackgroundColor: "#bfdbfe", Flag: "🇬🇷", CoatOfArms: "🛡️",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:   25,
			UnitSnipers:    12,
			UnitMarines:    7,
			UnitMediumTank: 4,
			UnitAntiTank:   2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitSnipers, UnitMarines, UnitMediumTank, UnitHowitzer, UnitAntiTank},
		IncomePercent: 105,
	},
	{
		ID: "china", Name: "China", Color: "#dc2626", BackgroundColor: "#fecaca", Flag: "🇨🇳", CoatOfArms: "🐉",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:  30,
			UnitMarines:   10,
			UnitEngineers: 5,
			UnitMortar:    3,
			UnitAPC:       2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitMarines, UnitEngineers, UnitMortar, UnitAPC},
		IncomePercent: 125,
	},
	{
		ID: "india", Name: "India", Color: "#f97316", BackgroundColor: "#fed7aa", Flag: "🇮🇳", CoatOfArms: "🐘",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:  25,
			UnitMarines:   12,
			UnitSnipers:   8,
			UnitEngineers: 3,
			UnitLightTank: 2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitMarines, UnitSnipers, UnitEngineers, UnitLightTank},
		IncomePercent: 120,
	},
	{
		ID: "japan", Name: "Japan", Color: "#f43f5e", BackgroundColor: "#fecdd3", Flag: "🇯🇵", CoatOfArms: "⛩️",
		StartingLoadout: map[UnitType]int{
			UnitMarines:      25,
			UnitParatroopers: 10,
			UnitSnipers:      8,
			UnitMediumTank:   4,
			UnitAntiTank:     3,
		},
		UnitRoster:    []UnitType{UnitMarines, UnitParatroopers, UnitSnipers, UnitMediumTank, UnitAntiTank, UnitHelicopter},
		IncomePercent: 100,
	},
	{
		ID: "mongols", Name: "Mongols", Color: "#78350f", BackgroundColor: "#fde68a", Flag: "🐎", CoatOfArms: "🏹",
		StartingLoadout: map[UnitType]int{
			UnitParatroopers:    25,
			UnitCommandos:       15,
			UnitHeavyTank:       6,
			UnitRocketArtillery: 3,
			UnitMissileLauncher: 1,
		},
		UnitRoster:    []UnitType{UnitParatroopers, UnitCommandos, UnitHeavyTank, UnitRocketArtillery, UnitMissileLauncher},
		IncomePercent: 85,
	},
	{
		ID: "vikings", Name: "Vikings", Color: "#0f766e", BackgroundColor: "#99f6e4", Flag: "⚔️", CoatOfArms: "⚒️",
		StartingLoadout: map[UnitType]int{
			UnitMarines:      25,
			UnitParatroopers: 12,
			UnitMediumTank:   6,
			UnitMortar:       4,
			UnitNavalStrike:  3,
		},
		UnitRoster:    []UnitType{UnitMarines, UnitParatroopers, UnitMediumTank, UnitMortar, UnitAntiTank, UnitNavalStrike},
		IncomePercent: 100,
	},
	{
		ID: "aztec", Name: "Aztec", Color: "#15803d", BackgroundColor: "#bbf7d0", Flag: "🦅", CoatOfArms: "🦅",
		StartingLoadout: map[UnitType]int{
			UnitParatroopers:    25,
			UnitCommandos:       12,
			UnitHeavyTank:       7,
			UnitRocketArtillery: 4,
			UnitHelicopter:      2,
		},
		UnitRoster:    []UnitType{UnitParatroopers, UnitCommandos, UnitHeavyTank, UnitRocketArtillery, UnitHelicopter},
		IncomePercent: 90,
	},
	{
		ID: "maya", Name: "Maya", Color: "#65a30d", BackgroundColor: "#d9f99d", Flag: "🐍", CoatOfArms: "🐍",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:   30,
			UnitSnipers:    10,
			UnitEngineers:  5,
			UnitLightTank:  3,
			UnitDroneSwarm: 2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitSnipers, UnitEngineers, UnitLightTank, UnitDroneSwarm},
		IncomePercent: 120,
	},
	{
		ID: "britain", Name: "Britain", Color: "#1d4ed8", BackgroundColor: "#bfdbfe", Flag: "🇬🇧", CoatOfArms: "🌹",
		StartingLoadout: map[UnitType]int{
			UnitMarines:      20,
			UnitParatroopers: 12,
			UnitMediumTank:   8,
			UnitHowitzer:     5,
			UnitHelicopter:   3,
			UnitNavalStrike:  2,
		},
		UnitRoster:    []UnitType{UnitMarines, UnitParatroopers, UnitMediumTank, UnitHowitzer, UnitHelicopter, UnitNavalStrike},
		IncomePercent: 100,
	},
	{
		ID: "france", Name: "France", Color: "#3b82f6", BackgroundColor: "#dbeafe", Flag: "🇫🇷", CoatOfArms: "⚜️",
		StartingLoadout: map[UnitType]int{
			UnitCommandos:  20,
			UnitMediumTank: 12,
			UnitHeavyTank:  10,
			UnitFighterJet: 5,
			UnitBomber:     3,
		},
		UnitRoster:    []UnitType{UnitCommandos, UnitMediumTank, UnitHeavyTank, UnitFighterJet, UnitBomber},
		IncomePercent: 90,
	},
	{
		ID: "spain", Name: "Spain", Color: "#facc15", BackgroundColor: "#fef9c3", Flag: "🇪🇸", CoatOfArms: "🐂",
		StartingLoadout: map[UnitType]int{
			UnitMarines:     20,
			UnitLightTank:   12,
			UnitMediumTank:  8,
			UnitHowitzer:    5,
			UnitHelicopter:  3,
			UnitNavalStrike: 2,
		},
		UnitRoster:    []UnitType{UnitMarines, UnitLightTank, UnitMediumTank, UnitHowitzer, UnitHelicopter, UnitNavalStrike},
		IncomePercent: 100,
	},
	{
		ID: "germany", Name: "Germany", Color: "#374151", BackgroundColor: "#e5e7eb", Flag: "🇩🇪", CoatOfArms: "🐺",
		StartingLoadout: map[UnitType]int{
			UnitCommandos:       20,
			UnitHeavyTank:       15,
			UnitRocketArtillery: 8,
			UnitFighterJet:      5,
			UnitMissileLauncher: 2,
		},
		UnitRoster:    []UnitType{UnitCommandos, UnitHeavyTank, UnitRocketArtillery, UnitFighterJet, UnitMissileLauncher},
		IncomePercent: 85,
	},
	{
		ID: "russia", Name: "Russia", Color: "#0ea5e9", BackgroundColor: "#e0f2fe", Flag: "🇷🇺", CoatOfArms: "🐻",
		StartingLoadout: map[UnitType]int{
			UnitParatroopers:    20,
			UnitHeavyTank:       15,
			UnitRocketArtillery: 7,
			UnitFighterJet:      5,
			UnitBomber:          3,
		},
		UnitRoster:    []UnitType{UnitParatroopers, UnitHeavyTank, UnitRocketArtillery, UnitFighterJet, UnitBomber},
		IncomePercent: 85,
	},
	{
		ID: "ottoman", Name: "Ottoman", Color: "#16a34a", BackgroundColor: "#bbf7d0", Flag: "☪️", CoatOfArms: "🌙",
		StartingLoadout: map[UnitType]int{
			UnitMarines:      20,
			UnitParatroopers: 12,
			UnitMediumTank:   8,
			UnitMortar:       5,
			UnitHowitzer:     3,
			UnitHelicopter:   2,
		},
		UnitRoster:    []UnitType{UnitMarines, UnitParatroopers, UnitMediumTank, UnitMortar, UnitHowitzer, UnitHelicopter},
		IncomePercent: 100,
	},
	{
		ID: "ethiopia", Name: "Ethiopia", Color: "#854d0e", BackgroundColor: "#fde68a", Flag: "🇪🇹", CoatOfArms: "⛰️",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:  25,
			UnitMarines:   12,
			UnitEngineers: 8,
			UnitAPC:       3,
			UnitLightTank: 2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitMarines, UnitEngineers, UnitAPC, UnitLightTank},
		IncomePercent: 125,
	},
	{
		ID: "mali", Name: "Mali", Color: "#d97706", BackgroundColor: "#fed7aa", Flag: "🦁", CoatOfArms: "🐪",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:  25,
			UnitSnipers:   12,
			UnitEngineers: 8,
			UnitMortar:    3,
			UnitAPC:       2,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitSnipers, UnitEngineers, UnitMortar, UnitAPC},
		IncomePercent: 125,
	},
	{
		ID: "inuit", Name: "Inuit", Color: "#94a3b8", BackgroundColor: "#e2e8f0", Flag: "❄️", CoatOfArms: "🐻‍❄️",
		StartingLoadout: map[UnitType]int{
			UnitRiflemen:   35,
			UnitSnipers:    10,
			UnitDroneSwarm: 4,
			UnitHelicopter: 1,
		},
		UnitRoster:    []UnitType{UnitRiflemen, UnitSnipers, UnitDroneSwarm, UnitHelicopter},
		IncomePercent: 130,
	},
}

func init() {
	for _, c := range Civilizations {
		sum := 0
		for ut, n := range c.StartingLoadout {
			sum += n
			if !inRoster(c.UnitRoster, ut) {
				panic("repko: civilization " + c.ID + " StartingLoadout key " + string(ut) + " not in UnitRoster")
			}
		}
		if sum != 50 {
			panic("repko: civilization " + c.ID + " StartingLoadout does not sum to 50")
		}
	}
}
