package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

// findEnemyByUID — поиск врага по уникальному ID.
// Используется гарпунами, чтобы найти своего владельца
// даже после удаления соседних врагов из слайса.
func findEnemyByUID(w *world.World, uid int) *world.Enemy {
	if uid == 0 {
		return nil
	}
	for i := range w.Enemies {
		if w.Enemies[i].UniqueID == uid {
			return &w.Enemies[i]
		}
	}
	return nil
}

// StepHarpoons — обработка всех гарпунов.
// Летящие: движение, коллизии, попадание в игрока.
// Прицепленные: тянут игрока к владельцу.
func StepHarpoons(w *world.World) {
	p := &w.Player

	for i := range w.Harpoons {
		h := &w.Harpoons[i]

		// ВРЕМЕННАЯ ДИАГНОСТИКА
		if h.Life%20 == 0 {
			owner := findEnemyByUID(w, h.OwnerUID)
			ownerStr := "nil"
			if owner != nil {
				ocx, ocy := owner.Center()
				ownerStr = "(" + itoa(int(ocx)) + "," + itoa(int(ocy)) + ")"
			}
			println("harpoon", h.ID, "uid=", h.OwnerUID,
				"pos=(", int(h.X), int(h.Y), ")",
				"attached=", h.Attached,
				"owner=", ownerStr)
		}

		// Уже мёртв — пропускаем (удаление в конце).
		if h.IsDead() {
			if p.Harpooned && p.HarpoonID == h.ID {
				p.Harpooned = false
				p.HarpoonID = -1
			}
			continue
		}

		// Владелец мёртв или исчез — гарпун умирает.
		owner := findEnemyByUID(w, h.OwnerUID)
		if owner == nil || owner.IsDead() {
			h.HP = 0
			if p.Harpooned && p.HarpoonID == h.ID {
				p.Harpooned = false
				p.HarpoonID = -1
			}
			continue
		}

		// Уже зацеплен — тянем игрока и обновляем позицию.
		if h.Attached {
			attachHarpoonToPlayer(w, p, h, owner)
			continue
		}

		// --- Полёт ---
		oldX, oldY := h.X, h.Y
		h.X += h.VX
		h.Y += h.VY
		h.Life--

		if h.Life <= 0 {
			h.HP = 0
			continue
		}

		// Попадание в игрока — CCD по отрезку.
		if !p.IsDead() && !p.Harpooned {
			playerRect := p.Rect()
			if _, ok := geometry.RaySegmentIntersectsRect(oldX, oldY, h.X, h.Y, playerRect); ok {
				h.Attached = true
				p.Harpooned = true
				p.HarpoonID = h.ID

				// Сразу сдвигаем наконечник к игроку на attachDist.
				pcx, pcy := p.Center()
				dx := h.X - pcx
				dy := h.Y - pcy
				l := math.Hypot(dx, dy)
				attachDist := world.EnemyHarpooner.Stats().HarpoonAttachDist
				if l > attachDist {
					h.X = pcx + dx/l*attachDist
					h.Y = pcy + dy/l*attachDist
				}
				continue
			}
		}

		// Попадание в стену — CCD по отрезку.
		hitWall := false
		for _, wall := range w.Walls {
			if _, ok := geometry.RaySegmentIntersectsRect(oldX, oldY, h.X, h.Y, wall); ok {
				hitWall = true
				break
			}
		}
		if hitWall {
			h.HP = 0
			continue
		}

		// За границей мира.
		if h.X < 0 || h.X > world.WorldWidth ||
			h.Y < 0 || h.Y > world.WorldHeight {
			h.HP = 0
			continue
		}
	}

	// Удаление мёртвых гарпунов.
	alive := w.Harpoons[:0]
	for i := range w.Harpoons {
		if !w.Harpoons[i].IsDead() {
			alive = append(alive, w.Harpoons[i])
		}
	}
	w.Harpoons = alive
}

// attachHarpoonToPlayer — пока гарпун зацеплен, тянет игрока
// к владельцу и держит наконечник на заданной дистанции от игрока.
func attachHarpoonToPlayer(w *world.World, p *world.Player, h *world.Harpoon, owner *world.Enemy) {
	stats := world.EnemyHarpooner.Stats()

	pcx, pcy := p.Center()
	ocx, ocy := owner.Center()

	// Вектор от игрока к владельцу.
	dx := ocx - pcx
	dy := ocy - pcy
	l := math.Hypot(dx, dy)

	// Игрок почти дотянут — отпускаем.
	if l < 30 {
		p.Harpooned = false
		p.HarpoonID = -1
		h.HP = 0
		return
	}

	// Тянем игрока в сторону владельца.
	pull := stats.HarpoonPullSpeed
	p.X += dx / l * pull
	p.Y += dy / l * pull

	// Обновляем позицию наконечника: он висит между игроком
	// и владельцем, на расстоянии attachDist от центра игрока.
	attachDist := stats.HarpoonAttachDist
	if l < attachDist {
		// Если игрок уже вплотную к владельцу — наконечник
		// ровно посередине.
		h.X = pcx + dx*0.5
		h.Y = pcy + dy*0.5
	} else {
		h.X = pcx + dx/l*attachDist
		h.Y = pcy + dy/l*attachDist
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
