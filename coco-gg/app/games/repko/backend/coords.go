package repko

import "math"

type Hex struct{ Q, R int }

func hexNeighbors(h Hex) [6]Hex {
	return [6]Hex{
		{Q: h.Q + 1, R: h.R},
		{Q: h.Q - 1, R: h.R},
		{Q: h.Q, R: h.R + 1},
		{Q: h.Q, R: h.R - 1},
		{Q: h.Q + 1, R: h.R - 1},
		{Q: h.Q - 1, R: h.R + 1},
	}
}

func cubeDistance(a, b Hex) int {
	dq := a.Q - b.Q
	dr := a.R - b.R
	return (abs(dq) + abs(dr) + abs(dq+dr)) / 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func hexToPixel(q, r int, size float64) (x, y float64) {
	x = size * (math.Sqrt(3)*float64(q) + math.Sqrt(3)/2*float64(r))
	y = size * (3.0 / 2 * float64(r))
	return
}

func hexCorner(cx, cy, size float64, i int) (x, y float64) {
	angle := math.Pi / 180 * (60*float64(i) - 30)
	return cx + size*math.Cos(angle), cy + size*math.Sin(angle)
}
