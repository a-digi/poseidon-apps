package repko

import (
	"errors"
	"log"
	"sort"
)

var (
	ErrNoGame = errors.New("game has not started")

	ErrNotYourTurn      = errors.New("not your turn")
	ErrInvalidPhase     = errors.New("invalid phase for this action")
	ErrInvalidTile      = errors.New("invalid tile")
	ErrInvalidUnit      = errors.New("invalid unit")
	ErrTileNotOwned     = errors.New("tile is not yours")
	ErrTileNotNeutral   = errors.New("tile is not neutral")
	ErrTileNotEnemy     = errors.New("tile is not held by an enemy")
	ErrOutOfRange            = errors.New("destination out of range")
	ErrInsufficientResources = errors.New("insufficient resources")
	ErrMaxLevel         = errors.New("unit already at max level")
	ErrCivAlreadyTaken  = errors.New("civilization already taken")
	ErrCivAlreadyPicked = errors.New("you already picked a civilization")
	ErrUnknownCivilization = errors.New("unknown civilization")
	ErrAttackLimitReached  = errors.New("attack limit reached this turn")
	ErrMoveLimitReached    = errors.New("move limit reached this turn")
	ErrDiplomacyAlreadyPending = errors.New("diplomacy offer already pending on this tile")
	ErrNoDiplomacyOffer        = errors.New("no diplomacy offer on this tile")
	ErrNotYourDiplomacyOffer   = errors.New("not your diplomacy offer")
	ErrEmptyUnitSelection      = errors.New("no units selected")
	ErrAttackFromEmptyTile     = errors.New("attacking from empty tile")
)

const (
	actionKeyMove   = "move"
	actionKeyAttack = "attack"
)

type ActionPickCivilization struct{ CivilizationID string }
type ActionPickStartingTile struct{ Q, R int }
type ActionRecruit struct {
	Q, R  int
	Unit  UnitType
	Count int
}
type ActionUpgrade struct {
	Q, R       int
	StackIndex int
}
type StackPick struct {
	StackIndex int `json:"stackIndex"`
	Count      int `json:"count"`
}
type ActionMove struct {
	FromQ, FromR int
	ToQ, ToR     int
	Units        []StackPick
}
type ActionAttack struct {
	FromQ, FromR int
	ToQ, ToR     int
	Units        []StackPick
}
type ActionBuyTile struct{ Q, R int }
type ActionUpgradeTile struct{ Q, R int }
type ActionOfferDiplomacy struct{ Q, R int }
type ActionAcceptDiplomacy struct{ Q, R int }
type ActionDeclineDiplomacy struct{ Q, R int }
type ActionCancelDiplomacy struct{ Q, R int }
type ActionEndTurn struct{}

func ValidateAndApply(state *GameState, playerID string, action any) error {
	if state == nil {
		return ErrNoGame
	}
	if state.Phase == PhaseGameOver {
		return ErrInvalidPhase
	}

	switch a := action.(type) {
	case ActionPickCivilization:
		if err := validateAndApplyPickCivilization(state, playerID, a); err != nil {
			return err
		}
	case ActionAcceptDiplomacy:
		if err := validateAndApplyAcceptDiplomacy(state, playerID, a); err != nil {
			return err
		}
	case ActionDeclineDiplomacy:
		if err := validateAndApplyDeclineDiplomacy(state, playerID, a); err != nil {
			return err
		}
	case ActionCancelDiplomacy:
		if err := validateAndApplyCancelDiplomacy(state, playerID, a); err != nil {
			return err
		}
	default:
		if state.Current == nil || state.Current.PlayerID != playerID {
			return ErrNotYourTurn
		}
		switch a := action.(type) {
		case ActionPickStartingTile:
			if err := validateAndApplyPickStartingTile(state, playerID, a); err != nil {
				return err
			}
		case ActionRecruit:
			if err := validateAndApplyRecruit(state, playerID, a); err != nil {
				return err
			}
		case ActionUpgrade:
			if err := validateAndApplyUpgrade(state, playerID, a); err != nil {
				return err
			}
		case ActionMove:
			if err := validateAndApplyMove(state, playerID, a); err != nil {
				return err
			}
		case ActionAttack:
			if err := validateAndApplyAttack(state, playerID, a); err != nil {
				return err
			}
		case ActionBuyTile:
			if err := validateAndApplyBuyTile(state, playerID, a); err != nil {
				return err
			}
		case ActionUpgradeTile:
			if err := validateAndApplyUpgradeTile(state, playerID, a); err != nil {
				return err
			}
		case ActionOfferDiplomacy:
			if err := validateAndApplyOfferDiplomacy(state, playerID, a); err != nil {
				return err
			}
		case ActionEndTurn:
			if err := validateAndApplyEndTurn(state, playerID); err != nil {
				return err
			}
		default:
			return ErrInvalidPhase
		}
	}

	recomputeCounters(state)
	return nil
}

func validateAndApplyPickCivilization(state *GameState, playerID string, a ActionPickCivilization) error {
	if state.Phase != PhaseCivPick {
		return ErrInvalidPhase
	}
	if !civilizationExists(state, a.CivilizationID) {
		return ErrUnknownCivilization
	}
	if _, taken := state.PickedCivs[a.CivilizationID]; taken {
		return ErrCivAlreadyTaken
	}
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	if player.CivilizationID != "" {
		return ErrCivAlreadyPicked
	}
	player.CivilizationID = a.CivilizationID
	state.PickedCivs[a.CivilizationID] = playerID
	advanceCivPick(state)
	return nil
}

func validateAndApplyPickStartingTile(state *GameState, playerID string, a ActionPickStartingTile) error {
	if state.Phase != PhaseTilePick {
		return ErrInvalidPhase
	}
	tile := state.tile(a.Q, a.R)
	if tile == nil {
		return ErrInvalidTile
	}
	if tile.OwnerID != "" {
		return ErrTileNotNeutral
	}
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	tile.OwnerID = playerID
	tile.FoundedBy = playerID
	civ := civilizationByID(state, player.CivilizationID)
	if civ != nil {
		tile.Name = civ.Name
	}
	tile.Yields = map[ResourceType]int{
		ResourceCredits: 5,
		ResourceSteel:   10,
		ResourceFuel:    10,
	}
	tile.Garrison = buildStartingGarrison(state, player.CivilizationID)
	player.Resources.add(ResourceBank{
		ResourceCredits: 30,
		ResourceSteel:   5,
		ResourceFuel:    60,
	})
	advanceTilePick(state)
	return nil
}

func buildStartingGarrison(state *GameState, civID string) []GarrisonStack {
	var loadout map[UnitType]int
	for _, c := range state.Civilizations {
		if c.ID == civID {
			loadout = c.StartingLoadout
			break
		}
	}
	order := []UnitType{UnitInfantry, UnitArmor, UnitJet}
	garrison := make([]GarrisonStack, 0, len(order))
	for _, ut := range order {
		count := loadout[ut]
		if count <= 0 {
			continue
		}
		garrison = append(garrison, GarrisonStack{Type: ut, Level: 1, Count: count})
	}
	return garrison
}

func validateAndApplyRecruit(state *GameState, playerID string, a ActionRecruit) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	tile := state.tile(a.Q, a.R)
	if tile == nil {
		return ErrInvalidTile
	}
	if tile.OwnerID != playerID {
		return ErrTileNotOwned
	}
	if !isValidUnitType(a.Unit) {
		return ErrInvalidUnit
	}
	if a.Count < 1 {
		return ErrInvalidUnit
	}
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	per := unitCost(a.Unit)
	total := scaleCost(per, a.Count)
	if !player.Resources.canAfford(total) {
		return ErrInsufficientResources
	}
	player.Resources.deduct(total)
	addToStack(&tile.Garrison, a.Unit, 1, a.Count)
	advancePlayingTurn(state)
	return nil
}

func addToStack(garrison *[]GarrisonStack, ut UnitType, level UnitLevel, count int) {
	for i := range *garrison {
		s := &(*garrison)[i]
		if s.Type == ut && s.Level == level {
			s.Count += count
			return
		}
	}
	*garrison = append(*garrison, GarrisonStack{Type: ut, Level: level, Count: count})
}

func validateAndApplyUpgrade(state *GameState, playerID string, a ActionUpgrade) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	tile := state.tile(a.Q, a.R)
	if tile == nil {
		return ErrInvalidTile
	}
	if tile.OwnerID != playerID {
		return ErrTileNotOwned
	}
	if a.StackIndex < 0 || a.StackIndex >= len(tile.Garrison) {
		return ErrInvalidUnit
	}
	s := &tile.Garrison[a.StackIndex]
	if s.Count == 0 {
		return ErrInvalidUnit
	}
	if s.Level >= 3 {
		return ErrMaxLevel
	}
	cost := upgradeCost(GarrisonStack{Type: s.Type, Level: s.Level, Count: 1})
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	if !player.Resources.canAfford(cost) {
		return ErrInsufficientResources
	}
	player.Resources.deduct(cost)
	srcType := s.Type
	srcLevel := s.Level
	s.Count--
	if s.Count == 0 {
		tile.Garrison = append(tile.Garrison[:a.StackIndex], tile.Garrison[a.StackIndex+1:]...)
	}
	addToStack(&tile.Garrison, srcType, srcLevel+1, 1)
	advancePlayingTurn(state)
	return nil
}

func validateAndApplyMove(state *GameState, playerID string, a ActionMove) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	if state.UsedActionsThisTurn[actionKeyMove] >= 1 {
		return ErrMoveLimitReached
	}
	from := Hex{Q: a.FromQ, R: a.FromR}
	to := Hex{Q: a.ToQ, R: a.ToR}
	if from == to {
		return ErrInvalidTile
	}
	fromTile := state.tile(from.Q, from.R)
	toTile := state.tile(to.Q, to.R)
	if fromTile == nil || toTile == nil {
		return ErrInvalidTile
	}
	if fromTile.OwnerID != playerID {
		return ErrTileNotOwned
	}
	if toTile.OwnerID != playerID {
		return ErrTileNotOwned
	}
	if cubeDistance(from, to) != 1 {
		return ErrOutOfRange
	}
	picks, err := validateStackPicks(a.Units, fromTile.Garrison)
	if err != nil {
		return err
	}
	for _, p := range picks {
		src := &fromTile.Garrison[p.StackIndex]
		addToStack(&toTile.Garrison, src.Type, src.Level, p.Count)
		src.Count -= p.Count
	}
	fromTile.Garrison = filterNonEmpty(fromTile.Garrison)
	state.UsedActionsThisTurn[actionKeyMove]++
	advancePlayingTurn(state)
	return nil
}

func validateAndApplyAttack(state *GameState, playerID string, a ActionAttack) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	if state.UsedActionsThisTurn[actionKeyAttack] >= 1 {
		return ErrAttackLimitReached
	}
	from := Hex{Q: a.FromQ, R: a.FromR}
	to := Hex{Q: a.ToQ, R: a.ToR}
	if from == to {
		return ErrInvalidTile
	}
	fromTile := state.tile(from.Q, from.R)
	toTile := state.tile(to.Q, to.R)
	if fromTile == nil || toTile == nil {
		return ErrInvalidTile
	}
	if fromTile.OwnerID != playerID {
		return ErrTileNotOwned
	}
	if toTile.OwnerID == playerID {
		return ErrTileNotEnemy
	}
	if cubeDistance(from, to) != 1 {
		return ErrOutOfRange
	}
	if len(fromTile.Garrison) == 0 {
		return ErrAttackFromEmptyTile
	}
	picks, err := validateStackPicks(a.Units, fromTile.Garrison)
	if err != nil {
		return err
	}
	defenderID := toTile.OwnerID
	attackerForces := make([]GarrisonStack, 0, len(picks))
	for _, p := range picks {
		src := fromTile.Garrison[p.StackIndex]
		attackerForces = append(attackerForces, GarrisonStack{Type: src.Type, Level: src.Level, Count: p.Count})
	}
	defendingUnits := append([]GarrisonStack{}, toTile.Garrison...)

	attackerWins, neutralResult, survivors := resolveCombat(attackerForces, defendingUnits)

	for _, p := range picks {
		fromTile.Garrison[p.StackIndex].Count -= p.Count
	}
	fromTile.Garrison = filterNonEmpty(fromTile.Garrison)

	var outcome string
	switch {
	case neutralResult:
		toTile.OwnerID = ""
		toTile.Garrison = []GarrisonStack{}
		outcome = "tie_neutral"
	case attackerWins:
		toTile.OwnerID = playerID
		toTile.Garrison = append([]GarrisonStack{}, survivors...)
		outcome = "attacker_won"
	default:
		toTile.Garrison = append([]GarrisonStack{}, survivors...)
		outcome = "defender_won"
	}
	state.UsedActionsThisTurn[actionKeyAttack]++

	log.Printf("game: repko combat (attacker=%s defender=%s from=(%d,%d) to=(%d,%d) result=%s survivors_power=%d)",
		playerID, defenderID, from.Q, from.R, to.Q, to.R, outcome, totalPower(survivors))

	checkAndApplyWin(state)
	advancePlayingTurn(state)
	return nil
}

func validateAndApplyBuyTile(state *GameState, playerID string, a ActionBuyTile) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	tile := state.tile(a.Q, a.R)
	if tile == nil {
		return ErrInvalidTile
	}
	if tile.OwnerID != "" {
		return ErrTileNotNeutral
	}
	dest := Hex{Q: a.Q, R: a.R}
	adjacent := false
	for _, src := range state.ownedTiles(playerID) {
		if cubeDistance(Hex{Q: src.Q, R: src.R}, dest) == 1 {
			adjacent = true
			break
		}
	}
	if !adjacent {
		return ErrOutOfRange
	}
	cost := ResourceBank{ResourceCredits: buyTileCost(tile)}
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	if !player.Resources.canAfford(cost) {
		return ErrInsufficientResources
	}
	player.Resources.deduct(cost)
	tile.OwnerID = playerID
	tile.Garrison = []GarrisonStack{}
	log.Printf("game: repko buy_tile (player=%s tile=(%d,%d) cost=%dc)",
		playerID, a.Q, a.R, cost[ResourceCredits])
	advancePlayingTurn(state)
	return nil
}

func validateAndApplyUpgradeTile(state *GameState, playerID string, a ActionUpgradeTile) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	tile := state.tile(a.Q, a.R)
	if tile == nil {
		return ErrInvalidTile
	}
	if tile.OwnerID != playerID {
		return ErrTileNotOwned
	}
	if tile.Production == ResourceNone {
		return ErrInvalidTile
	}
	cost := ResourceBank{ResourceCredits: 10}
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	if !player.Resources.canAfford(cost) {
		return ErrInsufficientResources
	}
	player.Resources.deduct(cost)
	tile.Yields[tile.Production]++
	log.Printf("game: repko upgrade_tile (player=%s tile=(%d,%d) production=%s yield=%d)",
		playerID, a.Q, a.R, tile.Production, tile.Yields[tile.Production])
	advancePlayingTurn(state)
	return nil
}

func validateAndApplyOfferDiplomacy(state *GameState, playerID string, a ActionOfferDiplomacy) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	tile := state.tile(a.Q, a.R)
	if tile == nil {
		return ErrInvalidTile
	}
	if tile.OwnerID == "" || tile.OwnerID == playerID {
		return ErrTileNotEnemy
	}
	if state.diplomacyForTile(a.Q, a.R) != nil {
		return ErrDiplomacyAlreadyPending
	}
	dest := Hex{Q: a.Q, R: a.R}
	adjacent := false
	for _, src := range state.ownedTiles(playerID) {
		if cubeDistance(Hex{Q: src.Q, R: src.R}, dest) == 1 {
			adjacent = true
			break
		}
	}
	if !adjacent {
		return ErrOutOfRange
	}
	player := state.playerByID(playerID)
	if player == nil {
		return ErrNotYourTurn
	}
	cost := diplomacyCost(tile.Garrison)
	if !player.Resources.canAfford(cost) {
		return ErrInsufficientResources
	}
	player.Resources.deduct(cost)
	state.PendingDiplomacy = append(state.PendingDiplomacy, DiplomacyOffer{
		Q:          a.Q,
		R:          a.R,
		AttackerID: playerID,
		DefenderID: tile.OwnerID,
	})
	log.Printf("game: repko diplomacy offer cost (player=%s credits=%d)", playerID, cost[ResourceCredits])
	advancePlayingTurn(state)
	return nil
}

func validateAndApplyAcceptDiplomacy(state *GameState, playerID string, a ActionAcceptDiplomacy) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	offer := state.diplomacyForTile(a.Q, a.R)
	if offer == nil {
		return ErrNoDiplomacyOffer
	}
	if offer.DefenderID != playerID {
		return ErrNotYourDiplomacyOffer
	}
	fromTile := state.tile(a.Q, a.R)
	if fromTile == nil {
		state.removeDiplomacy(a.Q, a.R)
		return ErrInvalidTile
	}
	if fromTile.OwnerID != playerID {
		state.removeDiplomacy(a.Q, a.R)
		return ErrNoDiplomacyOffer
	}
	attackerID := offer.AttackerID
	from := Hex{Q: a.Q, R: a.R}
	dest, ok := chooseRetreatDestination(state, playerID, from)
	if ok {
		toTile := state.tile(dest.Q, dest.R)
		for _, s := range fromTile.Garrison {
			if s.Count <= 0 {
				continue
			}
			addToStack(&toTile.Garrison, s.Type, s.Level, s.Count)
		}
	}
	fromTile.Garrison = []GarrisonStack{}
	fromTile.OwnerID = attackerID
	state.removeDiplomacy(a.Q, a.R)

	checkAndApplyWin(state)
	return nil
}

func validateAndApplyDeclineDiplomacy(state *GameState, playerID string, a ActionDeclineDiplomacy) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	offer := state.diplomacyForTile(a.Q, a.R)
	if offer == nil {
		return ErrNoDiplomacyOffer
	}
	if offer.DefenderID != playerID {
		return ErrNotYourDiplomacyOffer
	}
	state.removeDiplomacy(a.Q, a.R)
	return nil
}

func validateAndApplyCancelDiplomacy(state *GameState, playerID string, a ActionCancelDiplomacy) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	if state.Current == nil || state.Current.PlayerID != playerID {
		return ErrNotYourTurn
	}
	offer := state.diplomacyForTile(a.Q, a.R)
	if offer == nil {
		return ErrNoDiplomacyOffer
	}
	if offer.AttackerID != playerID {
		return ErrNotYourDiplomacyOffer
	}
	state.removeDiplomacy(a.Q, a.R)
	return nil
}

func validateAndApplyEndTurn(state *GameState, playerID string) error {
	if state.Phase != PhasePlaying {
		return ErrInvalidPhase
	}
	advancePlayingTurn(state)
	return nil
}

func diplomacyCost(defender []GarrisonStack) ResourceBank {
	return ResourceBank{ResourceCredits: totalPower(defender) * 2}
}

const (
	buyTileMin = 3
	buyTileMax = 15
)

func buyTileCost(tile *Tile) int {
	total := 0
	for _, y := range tile.Yields {
		total += y
	}
	cost := total * 3
	if cost < buyTileMin {
		cost = buyTileMin
	}
	if cost > buyTileMax {
		cost = buyTileMax
	}
	return cost
}

func unitCost(t UnitType) ResourceBank {
	switch t {
	case UnitInfantry:
		return ResourceBank{ResourceCredits: 5, ResourceSteel: 1, ResourceFuel: 0}
	case UnitArmor:
		return ResourceBank{ResourceCredits: 10, ResourceSteel: 2, ResourceFuel: 0}
	case UnitJet:
		return ResourceBank{ResourceCredits: 20, ResourceSteel: 3, ResourceFuel: 0}
	}
	return emptyResourceBank()
}

func upgradeCost(s GarrisonStack) ResourceBank {
	base := unitCost(s.Type)
	multiplier := 1
	if s.Level == 2 {
		multiplier = 2
	}
	return scaleCost(base, multiplier)
}

func scaleCost(cost ResourceBank, n int) ResourceBank {
	out := make(ResourceBank, len(cost))
	for k, v := range cost {
		out[k] = v * n
	}
	return out
}

func isValidUnitType(t UnitType) bool {
	return t == UnitInfantry || t == UnitArmor || t == UnitJet
}

func civilizationExists(state *GameState, civID string) bool {
	for _, c := range state.Civilizations {
		if c.ID == civID {
			return true
		}
	}
	return false
}

func civilizationByID(state *GameState, id string) *Civilization {
	for i := range state.Civilizations {
		if state.Civilizations[i].ID == id {
			return &state.Civilizations[i]
		}
	}
	return nil
}

func validateStackPicks(picks []StackPick, garrison []GarrisonStack) ([]StackPick, error) {
	if len(picks) == 0 {
		return nil, ErrEmptyUnitSelection
	}
	requested := make(map[int]int, len(picks))
	out := make([]StackPick, 0, len(picks))
	for _, p := range picks {
		if p.StackIndex < 0 || p.StackIndex >= len(garrison) {
			return nil, ErrInvalidUnit
		}
		if p.Count <= 0 {
			return nil, ErrInvalidUnit
		}
		requested[p.StackIndex] += p.Count
		if requested[p.StackIndex] > garrison[p.StackIndex].Count {
			return nil, ErrInvalidUnit
		}
		out = append(out, p)
	}
	return out, nil
}

func intermediatesAtDistance2(from, to Hex) []Hex {
	if cubeDistance(from, to) != 2 {
		return nil
	}
	fromNs := hexNeighbors(from)
	toNs := hexNeighbors(to)
	toSet := make(map[Hex]struct{}, len(toNs))
	for _, n := range toNs {
		toSet[n] = struct{}{}
	}
	out := make([]Hex, 0, 2)
	for _, n := range fromNs {
		if _, ok := toSet[n]; ok {
			out = append(out, n)
		}
	}
	return out
}

func hasRetreatPath(state *GameState, defenderID string, from, to Hex) bool {
	d := cubeDistance(from, to)
	if d == 1 {
		return true
	}
	if d != 2 {
		return false
	}
	for _, mid := range intermediatesAtDistance2(from, to) {
		tile := state.tile(mid.Q, mid.R)
		if tile == nil {
			continue
		}
		if tile.OwnerID == "" || tile.OwnerID == defenderID {
			return true
		}
	}
	return false
}

func chooseRetreatDestination(state *GameState, defenderID string, from Hex) (Hex, bool) {
	type candidate struct {
		hex      Hex
		garrison int
	}
	candidates := make([]candidate, 0)
	for _, t := range state.ownedTiles(defenderID) {
		h := Hex{Q: t.Q, R: t.R}
		if h == from {
			continue
		}
		if cubeDistance(h, from) > 2 {
			continue
		}
		if !hasRetreatPath(state, defenderID, from, h) {
			continue
		}
		soldiers := 0
		for _, s := range t.Garrison {
			soldiers += s.Count
		}
		candidates = append(candidates, candidate{hex: h, garrison: soldiers})
	}
	if len(candidates) == 0 {
		return Hex{}, false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].garrison != candidates[j].garrison {
			return candidates[i].garrison < candidates[j].garrison
		}
		if candidates[i].hex.Q != candidates[j].hex.Q {
			return candidates[i].hex.Q < candidates[j].hex.Q
		}
		return candidates[i].hex.R < candidates[j].hex.R
	})
	return candidates[0].hex, true
}

func advanceCivPick(state *GameState) {
	for _, p := range state.Players {
		if p.Disconnected || p.LeftGame {
			continue
		}
		if p.CivilizationID == "" {
			return
		}
	}
	state.Phase = PhaseTilePick
	for _, p := range state.Players {
		if p.Disconnected || p.LeftGame {
			continue
		}
		state.Current = &CurrentTurn{PlayerID: p.ID}
		log.Printf("game: repko phase advance (phase=tile_pick first=%s)", p.ID)
		return
	}
	state.Current = nil
}

func advanceTilePick(state *GameState) {
	tileCounts := state.tileCountByPlayer()
	for _, p := range state.Players {
		if p.Disconnected || p.LeftGame {
			continue
		}
		if tileCounts[p.ID] == 0 {
			state.Current = &CurrentTurn{PlayerID: p.ID}
			return
		}
	}
	state.Phase = PhasePlaying
	state.RoundNumber = 1
	state.PlayersActedThisRound = map[string]bool{}
	for _, p := range state.Players {
		if p.Disconnected || p.LeftGame {
			continue
		}
		state.Current = &CurrentTurn{PlayerID: p.ID}
		applyTurnIncome(state, p.ID)
		log.Printf("game: repko phase advance (phase=playing first=%s)", p.ID)
		return
	}
	state.Current = nil
}

func playerWithMostTiles(state *GameState) string {
	counts := state.tileCountByPlayer()
	var bestID string
	best := -1
	for _, p := range state.Players {
		c := counts[p.ID]
		if c > best || (c == best && (bestID == "" || p.ID < bestID)) {
			best = c
			bestID = p.ID
		}
	}
	return bestID
}

func advancePlayingTurn(state *GameState) {
	if state == nil || state.Phase == PhaseGameOver {
		return
	}
	cleanupStaleDiplomacy(state)
	state.UsedActionsThisTurn = map[string]int{}

	if state.Current != nil {
		if state.PlayersActedThisRound == nil {
			state.PlayersActedThisRound = map[string]bool{}
		}
		state.PlayersActedThisRound[state.Current.PlayerID] = true
	}

	allActed := true
	for _, p := range state.Players {
		if p.Disconnected || p.Eliminated || p.LeftGame {
			continue
		}
		if !state.PlayersActedThisRound[p.ID] {
			allActed = false
			break
		}
	}
	if allActed {
		state.RoundNumber++
		state.PlayersActedThisRound = map[string]bool{}
		if state.MaxRounds > 0 && state.RoundNumber > state.MaxRounds {
			state.Phase = PhaseGameOver
			state.WinnerID = playerWithMostTiles(state)
			state.Current = nil
			log.Printf("game: round limit reached (rounds=%d max=%d winner=%s)", state.RoundNumber-1, state.MaxRounds, state.WinnerID)
			return
		}
	}

	currentIdx := -1
	if state.Current != nil {
		currentIdx = state.playerIndex(state.Current.PlayerID)
	}
	n := len(state.Players)
	next := -1
	for i := 1; i <= n; i++ {
		candidate := (currentIdx + i) % n
		p := state.Players[candidate]
		if p.Eliminated || p.Disconnected || p.LeftGame {
			continue
		}
		next = candidate
		break
	}
	if next == -1 {
		state.Phase = PhaseGameOver
		state.Current = nil
		state.WinnerID = ""
		log.Printf("game: repko phase advance (phase=game_over reason=no_players)")
		return
	}
	state.Current = &CurrentTurn{PlayerID: state.Players[next].ID}
	applyTurnIncome(state, state.Players[next].ID)
	checkAndApplyWin(state)
}

func cleanupStaleDiplomacy(state *GameState) {
	if len(state.PendingDiplomacy) == 0 {
		return
	}
	kept := make([]DiplomacyOffer, 0, len(state.PendingDiplomacy))
	for _, offer := range state.PendingDiplomacy {
		attacker := state.playerByID(offer.AttackerID)
		defender := state.playerByID(offer.DefenderID)
		if attacker == nil || attacker.Disconnected || attacker.Eliminated || attacker.LeftGame {
			continue
		}
		if defender == nil || defender.Disconnected || defender.Eliminated || defender.LeftGame {
			continue
		}
		tile := state.tile(offer.Q, offer.R)
		if tile == nil || tile.OwnerID != offer.DefenderID {
			continue
		}
		kept = append(kept, offer)
	}
	state.PendingDiplomacy = kept
}

func checkAndApplyWin(state *GameState) {
	tileCounts := state.tileCountByPlayer()
	totalTiles := 0
	if state.Board != nil {
		totalTiles = len(state.Board.Tiles)
	}

	for _, p := range state.Players {
		if tileCounts[p.ID] == 0 && !p.Eliminated {
			p.Eliminated = true
			log.Printf("game: repko player eliminated (player_id=%s)", p.ID)
		}
	}

	var alive []*PlayerState
	for _, p := range state.Players {
		if !p.Eliminated {
			alive = append(alive, p)
		}
	}

	var winner *PlayerState
	if len(alive) == 1 {
		winner = alive[0]
	} else {
		threshold := totalTiles / 2
		for _, p := range state.Players {
			if tileCounts[p.ID] > threshold {
				winner = p
				break
			}
		}
	}

	if winner != nil {
		state.Phase = PhaseGameOver
		state.WinnerID = winner.ID
		state.Current = nil
		log.Printf("game: repko phase advance (phase=game_over winner=%s)", winner.ID)
	}
}

func releasePlayerTiles(state *GameState, playerID string) {
	if state.Board == nil {
		return
	}
	for _, t := range state.Board.Tiles {
		if t.OwnerID == playerID {
			t.OwnerID = ""
			t.Garrison = []GarrisonStack{}
		}
	}
}

func endGameIfOnlyOneActive(state *GameState) {
	if state.Phase == PhaseGameOver {
		return
	}
	active := 0
	var lastActive string
	for _, p := range state.Players {
		if p.LeftGame || p.Eliminated {
			continue
		}
		active++
		lastActive = p.ID
	}
	if active <= 1 {
		state.Phase = PhaseGameOver
		if active == 1 {
			state.WinnerID = lastActive
		} else {
			state.WinnerID = ""
		}
		state.Current = nil
	}
}

func recomputeCounters(state *GameState) {
	if state == nil || state.Board == nil {
		return
	}
	tileCounts := make(map[string]int, len(state.Players))
	unitCounts := make(map[string]int, len(state.Players))
	for _, t := range state.Board.Tiles {
		if t.OwnerID == "" {
			continue
		}
		tileCounts[t.OwnerID]++
		for _, s := range t.Garrison {
			unitCounts[t.OwnerID] += s.Count
		}
	}
	for _, p := range state.Players {
		p.TileCount = tileCounts[p.ID]
		p.UnitCount = unitCounts[p.ID]
	}
}

func applyTurnIncome(state *GameState, playerID string) {
	player := state.playerByID(playerID)
	if player == nil {
		return
	}
	owned := state.ownedTiles(playerID)
	income := computeIncome(owned)
	player.Resources.add(income)

	upkeep := sumUnits(owned)
	player.Resources[ResourceFuel] -= upkeep

	for player.Resources[ResourceFuel] < 0 {
		if !destroyLowestPowerUnit(owned) {
			break
		}
		player.Resources[ResourceFuel]++
	}
	for _, t := range owned {
		t.Garrison = filterNonEmpty(t.Garrison)
	}
	if player.Resources[ResourceFuel] < 0 {
		player.Resources[ResourceFuel] = 0
	}
}

func destroyLowestPowerUnit(tiles []*Tile) bool {
	type candidate struct {
		tileIdx  int
		stackIdx int
		power    int
		q, r     int
	}
	var best *candidate
	for ti, t := range tiles {
		for si, s := range t.Garrison {
			if s.Count <= 0 {
				continue
			}
			p := stackPower(s)
			c := candidate{tileIdx: ti, stackIdx: si, power: p, q: t.Q, r: t.R}
			if best == nil {
				cc := c
				best = &cc
				continue
			}
			if c.power < best.power ||
				(c.power == best.power && c.q < best.q) ||
				(c.power == best.power && c.q == best.q && c.r < best.r) ||
				(c.power == best.power && c.q == best.q && c.r == best.r && c.stackIdx < best.stackIdx) {
				cc := c
				best = &cc
			}
		}
	}
	if best == nil {
		return false
	}
	tile := tiles[best.tileIdx]
	tile.Garrison[best.stackIdx].Count--
	return true
}
