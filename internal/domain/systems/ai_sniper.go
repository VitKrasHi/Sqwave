package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

// sniperSettleTicks — сколько последних тиков прицел замирает
// перед выстрелом. Игрок за это время должен уйти с линии.
const sniperSettleTicks = 15

func stepSniper(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	if e.SniperShotLife > 0 {
		e.SniperShotLife--
	}

	if dist < stats.PreferredMin {
		if e.SniperState == world.SniperStateAiming {
			e.SniperState = world.SniperStateIdle
			e.SniperTimer = stats.SniperIdleMin
		}
		runSniperAway(w, e, ecx, ecy, pcx, pcy)
		return
	}

	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)

	switch e.SniperState {
	case world.SniperStateIdle:
		e.SniperTimer--
		if e.SniperTimer <= 0 && los {
			e.SniperState = world.SniperStateAiming
			e.SniperTimer = stats.SniperAimDuration
			e.SniperAimX = ecx
			e.SniperAimY = ecy
			return
		}
		if !los {
			moveEnemyToPoint(w, e, ecx, ecy, pcx, pcy)
			return
		}
		if dist < stats.PreferredMin*2 {
			runSniperAway(w, e, ecx, ecy, pcx, pcy)
			return
		}

	case world.SniperStateAiming:
		e.SniperTimer--

		// Первые (duration - settle) тиков — прицел активно ведётся
		// за игроком с опережением. Последние settle тиков — замер,
		// точка фиксируется, у игрока окно на уклонение.
		if e.SniperTimer > sniperSettleTicks {
			updateSniperAim(w, e, pcx, pcy)
		}
		// иначе — не трогаем e.SniperAimX/Y, они замёрзли.

		if e.SniperTimer <= 0 {
			fireSniperShot(w, e)
			e.SniperState = world.SniperStateIdle
			span := stats.SniperIdleMax - stats.SniperIdleMin
			e.SniperTimer = stats.SniperIdleMin + w.Rng.Intn(span+1)
		}
	}
}

// updateSniperAim — точка прицела стремится к предсказанной позиции
// игрока. Предсказание: pcx + vel * remaining * leadFactor.
//
// Луч не «залипает» за игроком, а живёт чуть впереди — там, где он
// будет через remaining тиков. Если игрок бежит по прямой — попадание.
// Если меняет направление — снайпер промахивается.
func updateSniperAim(w *world.World, e *world.Enemy, pcx, pcy float64) {
	remaining := float64(e.SniperTimer)
	if remaining < 0 {
		remaining = 0
	}

	if e.SniperLeadFactor == 0 {
		e.SniperLeadFactor = 0.7 + w.Rng.Float64()*0.4 // 0.7..1.1
	}

	// Сколько времени игрок продолжит идти в текущем направлении —
	// предсказываем на remaining тиков (не на весь aim duration,
	// потому что прицел всё равно «подтягивается» каждый тик).
	targetX := pcx + w.Player.VelX*remaining*e.SniperLeadFactor
	targetY := pcy + w.Player.VelY*remaining*e.SniperLeadFactor

	// Lerp. 0.18 — луч быстро наводится за 10–15 тиков.
	const lerp = 0.18
	e.SniperAimX += (targetX - e.SniperAimX) * lerp
	e.SniperAimY += (targetY - e.SniperAimY) * lerp
}

func runSniperAway(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) {
	awayX := ecx - pcx
	awayY := ecy - pcy
	l := math.Hypot(awayX, awayY)
	if l < 0.01 {
		return
	}
	awayX /= l
	awayY /= l
	speed := e.Speed() * 1.2
	moveEnemyX(w, e, awayX*speed)
	moveEnemyY(w, e, awayY*speed)
}

// fireSniperShot — выстрел в ЗАФИКСИРОВАННУЮ точку прицела,
// а не в текущую позицию игрока. Это даёт игроку реальный шанс
// увернуться: если он сместился за последние settle-тиков — промах.
func fireSniperShot(w *world.World, e *world.Enemy) {
	stats := e.Type.Stats()
	ecx, ecy := e.Center()

	ex := e.SniperAimX
	ey := e.SniperAimY

	// Обрезаем по первой стене.
	t := 1.0
	for _, wall := range w.Walls {
		if tt, ok := geometry.RaySegmentIntersectsRect(ecx, ecy, ex, ey, wall); ok {
			if tt < t {
				t = tt
			}
		}
	}
	ex = ecx + (ex-ecx)*t
	ey = ecy + (ey-ecy)*t

	e.SniperShotX = ex
	e.SniperShotY = ey
	e.SniperShotLife = 12

	e.SniperLeadFactor = 0

	// Попадание в игрока — только если луч реально пересекает его хитбокс.
	p := &w.Player
	pRect := p.Rect()
	if _, ok := geometry.RaySegmentIntersectsRect(ecx, ecy, ex, ey, pRect); ok {
		dmg := stats.SniperDamageMin +
			w.Rng.Intn(stats.SniperDamageMax-stats.SniperDamageMin+1)
		p.TakeDamage(dmg)
	}
}
