package repko

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand/v2"
)

type Board struct {
	Tiles []*Tile `json:"tiles"`

	hexSet      map[Hex]struct{}
	tileByCoord map[Hex]*Tile
	neighbors   map[Hex][]Hex
}

const (
	neutralBaseSize       = 5
	neutralYieldMultiplier = 8
)

func hexCoords() []Hex {
	rows := []struct {
		r    int
		qMin int
		qMax int
	}{
		{-3, 0, 2},
		{-2, -1, 2},
		{-1, -2, 2},
		{0, -3, 2},
		{1, -3, 1},
		{2, -3, 0},
		{3, -3, -1},
	}
	out := make([]Hex, 0, 30)
	for _, row := range rows {
		for q := row.qMin; q <= row.qMax; q++ {
			out = append(out, Hex{Q: q, R: row.r})
		}
	}
	return out
}

func shuffleProductions(rng *rand.Rand) []ResourceType {
	out := make([]ResourceType, 0, 30)
	add := func(rt ResourceType, n int) {
		for i := 0; i < n; i++ {
			out = append(out, rt)
		}
	}
	add(ResourceCredits, 8)
	add(ResourceSteel, 8)
	add(ResourceFuel, 8)
	add(ResourceNone, 6)
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func newRNG() *rand.Rand {
	var seed [32]byte
	if _, err := cryptorand.Read(seed[:]); err != nil {
		panic(err)
	}
	s1 := binary.LittleEndian.Uint64(seed[0:8])
	s2 := binary.LittleEndian.Uint64(seed[8:16])
	return rand.New(rand.NewPCG(s1, s2))
}

var goldNames = []string{
	"Sunspire", "Goldcrest", "Mintvale", "Lustrian Coast", "Glittermark",
	"Sovereign Heights", "Coinhold", "Aureate Run", "Marigold Bend", "Auric Fields",
	"Glintwater", "Sunken Treasury", "Crownwood", "Solspire", "Brilliance Reach",
}

var ironNames = []string{
	"Ironpeak", "Forgevale", "Steelhold", "Bleak Crags", "Anvilrest",
	"Greymarch", "Slagmoor", "Ferrum Ridge", "Hammerfall", "Coldforge",
	"Rustwatch", "Edgehold", "Ore Pass", "Mountains of Mir", "Iron Hollow",
}

var foodNames = []string{
	"Greenmeadow", "Wheatfields", "Orchard Glen", "Plentyvale", "Harvest Run",
	"Verdantmoor", "Sweetwater Bend", "Berryside", "Fallow Hill", "Threshing Plain",
	"Honeydown", "Larksong Vale", "Mossbottom", "Reapfield", "Granary Reach",
}

var noneNames = []string{
	"Driedrock", "Lonely Mesa", "Wasteland", "Bonefield", "Last Stop",
	"Wraithwaste", "Hollow March", "Sunbleach", "Quiet Reach", "Stillrock",
	"Forgotten Bluff", "Crowsong", "Greywillow", "Salt Flats", "Empty Quarter",
}

func NewBoard(rng *rand.Rand) *Board {
	coords := hexCoords()
	productions := shuffleProductions(rng)

	gold := append([]string(nil), goldNames...)
	iron := append([]string(nil), ironNames...)
	food := append([]string(nil), foodNames...)
	none := append([]string(nil), noneNames...)
	rng.Shuffle(len(gold), func(i, j int) { gold[i], gold[j] = gold[j], gold[i] })
	rng.Shuffle(len(iron), func(i, j int) { iron[i], iron[j] = iron[j], iron[i] })
	rng.Shuffle(len(food), func(i, j int) { food[i], food[j] = food[j], food[i] })
	rng.Shuffle(len(none), func(i, j int) { none[i], none[j] = none[j], none[i] })

	var goldI, ironI, foodI, noneI int
	pickName := func(p ResourceType) string {
		switch p {
		case ResourceCredits:
			if goldI < len(gold) {
				n := gold[goldI]
				goldI++
				return n
			}
		case ResourceSteel:
			if ironI < len(iron) {
				n := iron[ironI]
				ironI++
				return n
			}
		case ResourceFuel:
			if foodI < len(food) {
				n := food[foodI]
				foodI++
				return n
			}
		case ResourceNone:
			if noneI < len(none) {
				n := none[noneI]
				noneI++
				return n
			}
		}
		return "Wildlands"
	}

	tiles := make([]*Tile, 0, len(coords))
	for i, h := range coords {
		prod := productions[i]
		yields := map[ResourceType]int{}
		if prod != ResourceNone {
			yields[prod] = 2 + rng.IntN(2)
		}
		for _, t := range []ResourceType{ResourceCredits, ResourceSteel, ResourceFuel} {
			if _, exists := yields[t]; exists {
				continue
			}
			yields[t] = rng.IntN(2)
		}
		tiles = append(tiles, &Tile{
			Q:          h.Q,
			R:          h.R,
			Production: prod,
			Yields:     yields,
			OwnerID:    "",
			Name:       pickName(prod),
			Garrison:   make([]GarrisonStack, 0),
		})
	}

	for _, tile := range tiles {
		total := 0
		for _, y := range tile.Yields {
			total += y
		}
		if total == 0 {
			continue
		}
		size := neutralBaseSize + total*neutralYieldMultiplier
		tile.Garrison = []GarrisonStack{{Type: UnitRiflemen, Level: 1, Count: size}}
	}

	hexSet := make(map[Hex]struct{}, len(coords))
	for _, h := range coords {
		hexSet[h] = struct{}{}
	}

	tileByCoord := make(map[Hex]*Tile, len(tiles))
	for _, t := range tiles {
		tileByCoord[Hex{Q: t.Q, R: t.R}] = t
	}

	neighbors := make(map[Hex][]Hex, len(coords))
	for _, h := range coords {
		ns := hexNeighbors(h)
		realNs := make([]Hex, 0, 6)
		for _, n := range ns {
			if _, ok := hexSet[n]; ok {
				realNs = append(realNs, n)
			}
		}
		neighbors[h] = realNs
	}

	return &Board{
		Tiles:       tiles,
		hexSet:      hexSet,
		tileByCoord: tileByCoord,
		neighbors:   neighbors,
	}
}

func (b *Board) tile(h Hex) *Tile {
	return b.tileByCoord[h]
}

func (b *Board) hasTile(h Hex) bool {
	_, ok := b.hexSet[h]
	return ok
}

func (b *Board) neighborsOf(h Hex) []Hex {
	return b.neighbors[h]
}
