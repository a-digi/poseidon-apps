package repko

type UnitClass string

const (
	ClassInfantry  UnitClass = "infantry"
	ClassArmor     UnitClass = "armor"
	ClassArtillery UnitClass = "artillery"
	ClassAir       UnitClass = "air"
	ClassSpecial   UnitClass = "special"
)

type UnitSpec struct {
	ID    UnitType
	Name  string
	Icon  string
	Class UnitClass
	Power int
	Cost  ResourceBank
}

const (
	UnitRiflemen        UnitType = "riflemen"
	UnitMarines         UnitType = "marines"
	UnitSnipers         UnitType = "snipers"
	UnitEngineers       UnitType = "engineers"
	UnitParatroopers    UnitType = "paratroopers"
	UnitCommandos       UnitType = "commandos"
	UnitAPC             UnitType = "apc"
	UnitLightTank       UnitType = "light_tank"
	UnitMediumTank      UnitType = "medium_tank"
	UnitAntiTank        UnitType = "anti_tank"
	UnitHeavyTank       UnitType = "heavy_tank"
	UnitMortar          UnitType = "mortar"
	UnitHowitzer        UnitType = "howitzer"
	UnitRocketArtillery UnitType = "rocket_artillery"
	UnitDroneSwarm      UnitType = "drone_swarm"
	UnitHelicopter      UnitType = "helicopter"
	UnitFighterJet      UnitType = "fighter_jet"
	UnitBomber          UnitType = "bomber"
	UnitStealthBomber   UnitType = "stealth_bomber"
	UnitNavalStrike     UnitType = "naval_strike"
	UnitMissileLauncher UnitType = "missile_launcher"
	UnitCyberUnit       UnitType = "cyber_unit"
)

var UnitCatalog = map[UnitType]UnitSpec{
	UnitRiflemen:        {ID: UnitRiflemen, Name: "Riflemen", Icon: "🪖", Class: ClassInfantry, Power: 3, Cost: ResourceBank{ResourceCredits: 5, ResourceSteel: 1, ResourceFuel: 0}},
	UnitMarines:         {ID: UnitMarines, Name: "Marines", Icon: "🛡️", Class: ClassInfantry, Power: 4, Cost: ResourceBank{ResourceCredits: 8, ResourceSteel: 2, ResourceFuel: 0}},
	UnitSnipers:         {ID: UnitSnipers, Name: "Snipers", Icon: "🎯", Class: ClassInfantry, Power: 5, Cost: ResourceBank{ResourceCredits: 10, ResourceSteel: 1, ResourceFuel: 0}},
	UnitEngineers:       {ID: UnitEngineers, Name: "Engineers", Icon: "🛠️", Class: ClassInfantry, Power: 4, Cost: ResourceBank{ResourceCredits: 8, ResourceSteel: 2, ResourceFuel: 1}},
	UnitParatroopers:    {ID: UnitParatroopers, Name: "Paratroopers", Icon: "🪂", Class: ClassInfantry, Power: 6, Cost: ResourceBank{ResourceCredits: 12, ResourceSteel: 2, ResourceFuel: 1}},
	UnitCommandos:       {ID: UnitCommandos, Name: "Commandos", Icon: "🥷", Class: ClassInfantry, Power: 8, Cost: ResourceBank{ResourceCredits: 18, ResourceSteel: 3, ResourceFuel: 1}},
	UnitAPC:             {ID: UnitAPC, Name: "APC", Icon: "🚙", Class: ClassArmor, Power: 5, Cost: ResourceBank{ResourceCredits: 8, ResourceSteel: 4, ResourceFuel: 2}},
	UnitLightTank:       {ID: UnitLightTank, Name: "Light Tank", Icon: "🚜", Class: ClassArmor, Power: 6, Cost: ResourceBank{ResourceCredits: 10, ResourceSteel: 3, ResourceFuel: 1}},
	UnitMediumTank:      {ID: UnitMediumTank, Name: "Medium Tank", Icon: "🚛", Class: ClassArmor, Power: 8, Cost: ResourceBank{ResourceCredits: 14, ResourceSteel: 5, ResourceFuel: 2}},
	UnitAntiTank:        {ID: UnitAntiTank, Name: "Anti-Tank", Icon: "🛡️", Class: ClassArmor, Power: 9, Cost: ResourceBank{ResourceCredits: 12, ResourceSteel: 4, ResourceFuel: 2}},
	UnitHeavyTank:       {ID: UnitHeavyTank, Name: "Heavy Tank", Icon: "🦾", Class: ClassArmor, Power: 11, Cost: ResourceBank{ResourceCredits: 18, ResourceSteel: 7, ResourceFuel: 3}},
	UnitMortar:          {ID: UnitMortar, Name: "Mortar", Icon: "💥", Class: ClassArtillery, Power: 7, Cost: ResourceBank{ResourceCredits: 12, ResourceSteel: 3, ResourceFuel: 1}},
	UnitHowitzer:        {ID: UnitHowitzer, Name: "Howitzer", Icon: "🔫", Class: ClassArtillery, Power: 10, Cost: ResourceBank{ResourceCredits: 16, ResourceSteel: 5, ResourceFuel: 2}},
	UnitRocketArtillery: {ID: UnitRocketArtillery, Name: "Rocket Battery", Icon: "🚀", Class: ClassArtillery, Power: 13, Cost: ResourceBank{ResourceCredits: 22, ResourceSteel: 6, ResourceFuel: 3}},
	UnitDroneSwarm:      {ID: UnitDroneSwarm, Name: "Drone Swarm", Icon: "🛸", Class: ClassAir, Power: 8, Cost: ResourceBank{ResourceCredits: 16, ResourceSteel: 2, ResourceFuel: 4}},
	UnitHelicopter:      {ID: UnitHelicopter, Name: "Helicopter", Icon: "🚁", Class: ClassAir, Power: 9, Cost: ResourceBank{ResourceCredits: 18, ResourceSteel: 3, ResourceFuel: 5}},
	UnitFighterJet:      {ID: UnitFighterJet, Name: "Fighter Jet", Icon: "✈️", Class: ClassAir, Power: 12, Cost: ResourceBank{ResourceCredits: 25, ResourceSteel: 4, ResourceFuel: 7}},
	UnitBomber:          {ID: UnitBomber, Name: "Bomber", Icon: "💣", Class: ClassAir, Power: 15, Cost: ResourceBank{ResourceCredits: 32, ResourceSteel: 5, ResourceFuel: 10}},
	UnitStealthBomber:   {ID: UnitStealthBomber, Name: "Stealth Bomber", Icon: "🦇", Class: ClassAir, Power: 18, Cost: ResourceBank{ResourceCredits: 45, ResourceSteel: 7, ResourceFuel: 12}},
	UnitNavalStrike:     {ID: UnitNavalStrike, Name: "Naval Strike", Icon: "🚢", Class: ClassSpecial, Power: 12, Cost: ResourceBank{ResourceCredits: 22, ResourceSteel: 6, ResourceFuel: 4}},
	UnitMissileLauncher: {ID: UnitMissileLauncher, Name: "Missile Launcher", Icon: "🛰️", Class: ClassSpecial, Power: 16, Cost: ResourceBank{ResourceCredits: 30, ResourceSteel: 8, ResourceFuel: 6}},
	UnitCyberUnit:       {ID: UnitCyberUnit, Name: "Cyberwarfare", Icon: "💻", Class: ClassSpecial, Power: 10, Cost: ResourceBank{ResourceCredits: 25, ResourceSteel: 1, ResourceFuel: 1}},
}

var UnitOrder = []UnitType{
	UnitRiflemen, UnitMarines, UnitSnipers, UnitEngineers, UnitParatroopers, UnitCommandos,
	UnitAPC, UnitLightTank, UnitMediumTank, UnitAntiTank, UnitHeavyTank,
	UnitMortar, UnitHowitzer, UnitRocketArtillery,
	UnitDroneSwarm, UnitHelicopter, UnitFighterJet, UnitBomber, UnitStealthBomber,
	UnitNavalStrike, UnitMissileLauncher, UnitCyberUnit,
}

func inRoster(roster []UnitType, t UnitType) bool {
	for _, r := range roster {
		if r == t {
			return true
		}
	}
	return false
}

func cloneResourceBank(b ResourceBank) ResourceBank {
	out := make(ResourceBank, len(b))
	for k, v := range b {
		out[k] = v
	}
	return out
}
