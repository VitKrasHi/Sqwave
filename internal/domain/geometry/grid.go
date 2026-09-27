package geometry

import (
	"container/heap"
	"math"
)

type Point struct {
	X, Y float64
}

type Grid struct {
	CellSize float64
	Cols     int
	Rows     int
	Blocked  []bool
}

func NewGrid(worldW, worldH, cellSize float64) *Grid {
	cols := int(math.Ceil(worldW / cellSize))
	rows := int(math.Ceil(worldH / cellSize))
	return &Grid{
		CellSize: cellSize,
		Cols:     cols,
		Rows:     rows,
		Blocked:  make([]bool, cols*rows),
	}
}

func (g *Grid) InBounds(cx, cy int) bool {
	return cx >= 0 && cx < g.Cols && cy >= 0 && cy < g.Rows
}

// IsBlocked возвращает true и для клеток за границей мира,
// чтобы A* не пытался выйти наружу.
func (g *Grid) IsBlocked(cx, cy int) bool {
	if !g.InBounds(cx, cy) {
		return true
	}
	return g.Blocked[cy*g.Cols+cx]
}

func (g *Grid) SetBlocked(cx, cy int, blocked bool) {
	if !g.InBounds(cx, cy) {
		return
	}
	g.Blocked[cy*g.Cols+cx] = blocked
}

func (g *Grid) WorldToCell(x, y float64) (int, int) {
	return int(x / g.CellSize), int(y / g.CellSize)
}

func (g *Grid) CellCenter(cx, cy int) (float64, float64) {
	return (float64(cx) + 0.5) * g.CellSize, (float64(cy) + 0.5) * g.CellSize
}

// MarkRect помечает все клетки, пересекающиеся с прямоугольником.
// Консервативно: даже частичное пересечение блокирует всю клетку,
// чтобы враг не пытался «протиснуться» сквозь угол.
// MarkRectForAgent помечает клетку как blocked, если агент заданного
// размера, поставленный центром в центр этой клетки, пересекается
// с прямоугольником r. Это точнее MarkRect: агент может идти вдоль
// стены, не считая прилегающие к стене клетки «фантомно» занятыми.
func (g *Grid) MarkRectForAgent(r Rect, agentSize float64) {
	half := agentSize / 2

	// Диапазон клеток, чей центр может попасть в зону досягаемости.
	x0 := int((r.X - half) / g.CellSize)
	y0 := int((r.Y - half) / g.CellSize)
	x1 := int((r.X + r.W + half) / g.CellSize)
	y1 := int((r.Y + r.H + half) / g.CellSize)

	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			if !g.InBounds(cx, cy) {
				continue
			}
			centerX, centerY := g.CellCenter(cx, cy)
			agent := Rect{
				X: centerX - half,
				Y: centerY - half,
				W: agentSize,
				H: agentSize,
			}
			if agent.Intersects(r) {
				g.SetBlocked(cx, cy, true)
			}
		}
	}
}

// NearestFree ищет ближайшую свободную клетку в радиусе maxRadius.
// Нужно, когда цель или старт оказались в «заблокированной» клетке —
// например, игрок вплотную прижался к стене.
func (g *Grid) NearestFree(cx, cy, maxRadius int) (int, int, bool) {
	if !g.IsBlocked(cx, cy) {
		return cx, cy, true
	}
	for r := 1; r <= maxRadius; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if absInt(dx) != r && absInt(dy) != r {
					continue
				}
				nx, ny := cx+dx, cy+dy
				if !g.IsBlocked(nx, ny) {
					return nx, ny, true
				}
			}
		}
	}
	return 0, 0, false
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ---------- A* ----------

var astarDirs = [8][2]int{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
}

const (
	astarStraightCost = 10
	astarDiagonalCost = 14
)

type astarNode struct {
	x, y   int
	g      int
	f      int
	parent *astarNode
}

type astarHeap []*astarNode

func (h astarHeap) Len() int           { return len(h) }
func (h astarHeap) Less(i, j int) bool { return h[i].f < h[j].f }
func (h astarHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *astarHeap) Push(x any)        { *h = append(*h, x.(*astarNode)) }
func (h *astarHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// heuristic — octile distance, согласованная с весами шагов.
func heuristic(ax, ay, bx, by int) int {
	dx := absInt(ax - bx)
	dy := absInt(ay - by)
	if dx < dy {
		dx, dy = dy, dx
	}
	return astarDiagonalCost*dy + astarStraightCost*(dx-dy)
}

// FindPath ищет кратчайший путь между клетками. Возвращает срез клеток
// от стартовой (не включена) до целевой (включена), либо nil.
func (g *Grid) FindPath(sx, sy, tx, ty int) [][2]int {
	if g.IsBlocked(sx, sy) || g.IsBlocked(tx, ty) {
		return nil
	}
	if sx == tx && sy == ty {
		return nil
	}

	start := &astarNode{x: sx, y: sy, f: heuristic(sx, sy, tx, ty)}
	open := &astarHeap{start}
	heap.Init(open)

	best := map[[2]int]*astarNode{{sx, sy}: start}

	for open.Len() > 0 {
		cur := heap.Pop(open).(*astarNode)
		key := [2]int{cur.x, cur.y}

		// Устаревшая запись: этот узел уже был улучшен.
		if best[key] != cur {
			continue
		}

		if cur.x == tx && cur.y == ty {
			return reconstructPath(cur)
		}

		for _, d := range astarDirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if g.IsBlocked(nx, ny) {
				continue
			}
			// Запрет «срезать угол» между двумя блокированными
			// клетками по диагонали: враг физически не пройдёт.
			if d[0] != 0 && d[1] != 0 {
				if g.IsBlocked(cur.x+d[0], cur.y) || g.IsBlocked(cur.x, cur.y+d[1]) {
					continue
				}
			}

			step := astarStraightCost
			if d[0] != 0 && d[1] != 0 {
				step = astarDiagonalCost
			}
			ng := cur.g + step

			nkey := [2]int{nx, ny}
			if prev, ok := best[nkey]; ok && prev.g <= ng {
				continue
			}
			n := &astarNode{
				x: nx, y: ny,
				g:      ng,
				f:      ng + heuristic(nx, ny, tx, ty),
				parent: cur,
			}
			best[nkey] = n
			heap.Push(open, n)
		}
	}
	return nil
}

func reconstructPath(n *astarNode) [][2]int {
	var rev [][2]int
	for n != nil {
		rev = append(rev, [2]int{n.x, n.y})
		n = n.parent
	}
	// Убираем стартовую клетку, оставляем всё остальное в порядке от старта.
	out := make([][2]int, 0, len(rev)-1)
	for i := len(rev) - 2; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out
}
