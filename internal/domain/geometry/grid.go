package geometry

import (
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

	// Переиспользуемые буферы A*. Инициализируются лениво.
	astarG      []int32
	astarParent []int32
	astarClosed []bool
	astarHeap   []int32 // индексы клеток, min-heap по f
	astarF      []int32
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

// FindPath — итеративный A* без аллокаций.
// Внимание: не потокобезопасен — буферы разделяются между вызовами.
func (g *Grid) FindPath(sx, sy, tx, ty int) [][2]int {
	return g.FindPathAvoid(sx, sy, tx, ty, nil)
}

// FindPathAvoid — то же, что FindPath, но с учётом карты занятости
// avoid (размер Cols*Rows). Клетки, помеченные в avoid, считаются
// заблокированными. Стартовая и целевая клетки всегда разрешены —
// иначе агент, стоящий в занятой клетке, не смог бы стартовать.
func (g *Grid) FindPathAvoid(sx, sy, tx, ty int, avoid []bool) [][2]int {
	if g.IsBlocked(sx, sy) || g.IsBlocked(tx, ty) {
		return nil
	}
	if sx == tx && sy == ty {
		return nil
	}

	total := g.Cols * g.Rows
	if avoid != nil && len(avoid) != total {
		avoid = nil
	}
	g.ensureAstarBuffers(total)

	for i := range g.astarG {
		g.astarG[i] = -1
		g.astarClosed[i] = false
		g.astarParent[i] = -1
		g.astarF[i] = 0
	}
	g.astarHeap = g.astarHeap[:0]

	startIdx := sy*g.Cols + sx
	targetIdx := ty*g.Cols + tx

	blocked := func(x, y int) bool {
		if g.IsBlocked(x, y) {
			return true
		}
		idx := y*g.Cols + x
		if idx == startIdx || idx == targetIdx {
			return false
		}
		return avoid != nil && avoid[idx]
	}

	g.astarG[startIdx] = 0
	g.astarF[startIdx] = int32(heuristic(sx, sy, tx, ty))
	g.astarHeap = append(g.astarHeap, int32(startIdx))

	for len(g.astarHeap) > 0 {
		best := 0
		for i := 1; i < len(g.astarHeap); i++ {
			if g.astarF[g.astarHeap[i]] < g.astarF[g.astarHeap[best]] {
				best = i
			}
		}
		curIdx := int(g.astarHeap[best])
		g.astarHeap[best] = g.astarHeap[len(g.astarHeap)-1]
		g.astarHeap = g.astarHeap[:len(g.astarHeap)-1]

		if curIdx == targetIdx {
			return g.reconstructAstar(startIdx, targetIdx)
		}
		if g.astarClosed[curIdx] {
			continue
		}
		g.astarClosed[curIdx] = true

		cx := curIdx % g.Cols
		cy := curIdx / g.Cols

		for _, d := range astarDirs {
			nx, ny := cx+d[0], cy+d[1]
			if blocked(nx, ny) {
				continue
			}
			if d[0] != 0 && d[1] != 0 {
				if blocked(cx+d[0], cy) || blocked(cx, cy+d[1]) {
					continue
				}
			}

			nIdx := ny*g.Cols + nx
			if g.astarClosed[nIdx] {
				continue
			}

			step := astarStraightCost
			if d[0] != 0 && d[1] != 0 {
				step = astarDiagonalCost
			}
			ng := g.astarG[curIdx] + int32(step)

			if g.astarG[nIdx] != -1 && g.astarG[nIdx] <= ng {
				continue
			}
			g.astarG[nIdx] = ng
			g.astarParent[nIdx] = int32(curIdx)
			g.astarF[nIdx] = ng + int32(heuristic(nx, ny, tx, ty))
			g.astarHeap = append(g.astarHeap, int32(nIdx))
		}
	}
	return nil
}

func (g *Grid) ensureAstarBuffers(total int) {
	if cap(g.astarG) < total {
		g.astarG = make([]int32, total)
		g.astarParent = make([]int32, total)
		g.astarClosed = make([]bool, total)
		g.astarF = make([]int32, total)
	} else {
		g.astarG = g.astarG[:total]
		g.astarParent = g.astarParent[:total]
		g.astarClosed = g.astarClosed[:total]
		g.astarF = g.astarF[:total]
	}
}

func (g *Grid) reconstructAstar(startIdx, targetIdx int) [][2]int {
	var rev []int
	cur := targetIdx
	for cur != startIdx && cur != -1 {
		rev = append(rev, cur)
		cur = int(g.astarParent[cur])
	}
	out := make([][2]int, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		idx := rev[i]
		out = append(out, [2]int{idx % g.Cols, idx / g.Cols})
	}
	return out
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
