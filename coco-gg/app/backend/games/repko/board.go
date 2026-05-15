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
	add(ResourceGold, 8)
	add(ResourceIron, 8)
	add(ResourceFood, 8)
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

func NewBoard(rng *rand.Rand) *Board {
	coords := hexCoords()
	productions := shuffleProductions(rng)

	tiles := make([]*Tile, 0, len(coords))
	for i, h := range coords {
		prod := productions[i]
		yield := 0
		if prod != ResourceNone {
			yield = rng.IntN(3) + 1
		}
		tiles = append(tiles, &Tile{
			Q:          h.Q,
			R:          h.R,
			Production: prod,
			Yield:      yield,
			OwnerID:    "",
			Garrison:   make([]Unit, 0),
		})
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
