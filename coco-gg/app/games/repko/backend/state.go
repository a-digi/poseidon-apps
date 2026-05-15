package repko

type Phase string

const (
	PhaseLobby    Phase = "lobby"
	PhaseCivPick  Phase = "civ_pick"
	PhaseTilePick Phase = "tile_pick"
	PhasePlaying  Phase = "playing"
	PhaseGameOver Phase = "game_over"
)

type ResourceType string

const (
	ResourceCredits ResourceType = "credits"
	ResourceSteel   ResourceType = "steel"
	ResourceFuel    ResourceType = "fuel"
	ResourceNone    ResourceType = "none"
)

type UnitType string

type UnitLevel int

type GarrisonStack struct {
	Type  UnitType  `json:"type"`
	Level UnitLevel `json:"level"`
	Count int       `json:"count"`
}

type Civilization struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Color           string           `json:"color"`
	BackgroundColor string           `json:"backgroundColor"`
	Flag            string           `json:"flag"`
	CoatOfArms      string           `json:"coatOfArms"`
	StartingLoadout map[UnitType]int `json:"startingLoadout"`
	UnitRoster      []UnitType       `json:"unitRoster"`
	IncomePercent   int              `json:"incomePercent"`
}

type DiplomacyOffer struct {
	Q          int    `json:"q"`
	R          int    `json:"r"`
	AttackerID string `json:"attackerId"`
	DefenderID string `json:"defenderId"`
}

type Tile struct {
	Q          int                  `json:"q"`
	R          int                  `json:"r"`
	Production ResourceType         `json:"production"`
	Yields     map[ResourceType]int `json:"yields"`
	OwnerID    string               `json:"ownerId"`
	FoundedBy  string               `json:"foundedBy,omitempty"`
	Name       string               `json:"name,omitempty"`
	Garrison   []GarrisonStack      `json:"garrison"`
}

type PlayerState struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Color          string       `json:"color"`
	CivilizationID string       `json:"civilizationId"`
	TileCount      int          `json:"tileCount"`
	UnitCount      int          `json:"unitCount"`
	Eliminated     bool         `json:"eliminated"`
	LeftGame       bool         `json:"leftGame"`
	Resources      ResourceBank `json:"-"`
	Disconnected   bool         `json:"-"`
}

type CurrentTurn struct {
	PlayerID   string `json:"playerId"`
	DeadlineMs int64  `json:"deadlineMs,omitempty"`
}

type YouState struct {
	Resources ResourceBank `json:"resources"`
}

type ResourceBank map[ResourceType]int

type GameState struct {
	Phase                 Phase
	Board                 *Board
	Players               []*PlayerState
	Current               *CurrentTurn
	Civilizations         []Civilization
	PickedCivs            map[string]string
	UsedActionsThisTurn   map[string]int
	MaxRounds             int             `json:"maxRounds"`
	RoundNumber           int             `json:"roundNumber"`
	PlayersActedThisRound map[string]bool `json:"-"`
	PendingDiplomacy      []DiplomacyOffer
	WinnerID              string
}

func (g *GameState) tile(q, r int) *Tile {
	if g == nil || g.Board == nil {
		return nil
	}
	return g.Board.tile(Hex{Q: q, R: r})
}

func (g *GameState) playerByID(id string) *PlayerState {
	for _, p := range g.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *GameState) playerIndex(id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func (g *GameState) ownedTiles(playerID string) []*Tile {
	out := make([]*Tile, 0)
	if g == nil || g.Board == nil {
		return out
	}
	for _, t := range g.Board.Tiles {
		if t.OwnerID == playerID {
			out = append(out, t)
		}
	}
	return out
}

func (g *GameState) tileCountByPlayer() map[string]int {
	out := make(map[string]int, len(g.Players))
	if g == nil || g.Board == nil {
		return out
	}
	for _, t := range g.Board.Tiles {
		if t.OwnerID == "" {
			continue
		}
		out[t.OwnerID]++
	}
	return out
}

func (g *GameState) diplomacyForTile(q, r int) *DiplomacyOffer {
	for i := range g.PendingDiplomacy {
		if g.PendingDiplomacy[i].Q == q && g.PendingDiplomacy[i].R == r {
			return &g.PendingDiplomacy[i]
		}
	}
	return nil
}

func (g *GameState) removeDiplomacy(q, r int) {
	for i := range g.PendingDiplomacy {
		if g.PendingDiplomacy[i].Q == q && g.PendingDiplomacy[i].R == r {
			g.PendingDiplomacy = append(g.PendingDiplomacy[:i], g.PendingDiplomacy[i+1:]...)
			return
		}
	}
}

func emptyResourceBank() ResourceBank {
	return ResourceBank{
		ResourceCredits: 0,
		ResourceSteel:   0,
		ResourceFuel:    0,
	}
}

func (b ResourceBank) canAfford(cost ResourceBank) bool {
	for k, v := range cost {
		if b[k] < v {
			return false
		}
	}
	return true
}

func (b ResourceBank) deduct(cost ResourceBank) {
	for k, v := range cost {
		b[k] -= v
	}
}

func (b ResourceBank) add(other ResourceBank) {
	for k, v := range other {
		b[k] += v
	}
}

// newPlayerState returns a PlayerState with non-nil empty Resources so JSON
// marshalling and downstream resource math always have a valid map. Mirrors
// the prior fix for the same JSON-null hazard on player slice fields.
func newPlayerState(id, name, color string) *PlayerState {
	return &PlayerState{
		ID:        id,
		Name:      name,
		Color:     color,
		Resources: emptyResourceBank(),
	}
}
