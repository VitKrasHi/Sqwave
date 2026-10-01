package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

const (
	separationRadiusMul = 1.35
	separationStrength  = 1.0
)

// applySeparation расталкивает врагов, которые слишком близко.
// Обнуление «назад от игрока» применяется только к тем, кто
// в фазе преследования. Стрелок на орбите и разведчик в уклонении
// получают полное расталкивание — иначе они не разводятся.
func applySeparation(w *world.World) {
	if w.SpatialHash == nil {
		return
	}

	pcx, pcy := w.Player.Center()

	type push struct {
		idx    int
		dx, dy float64
	}
	pushes := make([]push, 0, len(w.Enemies))

	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		size := e.Type.Stats().Size
		radius := size * separationRadiusMul
		ecx, ecy := e.Center()

		var sumX, sumY float64

		w.SpatialHash.ForEachNear(ecx, ecy, func(j int) {
			if j == i {
				return
			}
			other := &w.Enemies[j]
			if other.IsDead() {
				return
			}
			ocx, ocy := other.Center()
			dx := ecx - ocx
			dy := ecy - ocy
			d := math.Hypot(dx, dy)
			if d < 0.01 || d > radius {
				return
			}
			strength := (radius - d) / radius
			sumX += dx / d * strength
			sumY += dy / d * strength
		})

		if math.Abs(sumX) < 0.001 && math.Abs(sumY) < 0.001 {
			continue
		}
		l := math.Hypot(sumX, sumY)
		if l < 0.01 {
			continue
		}
		px := sumX / l
		py := sumY / l

		// Обнуляем «назад от игрока» только для тех, кто
		// сейчас преследует — пехотинец всегда, стрелок
		// только вне комфортной зоны.
		suppressBackward := false
		switch e.Type {
		case world.EnemyInfantry:
			suppressBackward = true
		case world.EnemyShooter:
			dist := math.Hypot(pcx-ecx, pcy-ecy)
			suppressBackward = dist > e.Type.Stats().PreferredMax
		case world.EnemyScout:
			suppressBackward = e.RecentlyHitTimer == 0
		}

		if suppressBackward {
			fx := e.FacingX
			fy := e.FacingY
			dot := px*fx + py*fy
			if dot < 0 {
				px -= dot * fx
				py -= dot * fy
				l2 := math.Hypot(px, py)
				if l2 < 0.01 {
					continue
				}
				px /= l2
				py /= l2
			}
		}

		pushes = append(pushes, push{i, px, py})
	}

	for _, p := range pushes {
		e := &w.Enemies[p.idx]
		speed := e.Speed() * separationStrength * 0.4
		applySeparationMove(w, e, p.dx*speed, p.dy*speed)
	}
}

func applySeparationMove(w *world.World, e *world.Enemy, dx, dy float64) {
	if dx == 0 && dy == 0 {
		return
	}
	size := e.Type.Stats().Size
	newX := e.X + dx
	newY := e.Y + dy

	if newX < 0 {
		newX = 0
	}
	if newY < 0 {
		newY = 0
	}
	if newX+size > world.WorldWidth {
		newX = world.WorldWidth - size
	}
	if newY+size > world.WorldHeight {
		newY = world.WorldHeight - size
	}

	rectX := geometry.Rect{X: newX, Y: e.Y, W: size, H: size}
	if !rectHitsWall(w, rectX) {
		e.X = newX
	}
	rectY := geometry.Rect{X: e.X, Y: newY, W: size, H: size}
	if !rectHitsWall(w, rectY) {
		e.Y = newY
	}
}

func rectHitsWall(w *world.World, r geometry.Rect) bool {
	for _, wall := range w.Walls {
		if r.Intersects(wall) {
			return true
		}
	}
	return false
}
