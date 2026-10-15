package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepSummoner(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	// Первый тик — присваиваем ID.
	if e.SummonerID == 0 {
		e.SummonerID = w.AllocSummonerID()
		// Ставим кулдаун чуть меньше первой волны, чтобы он сразу начал.
		if e.SummonCooldown == 0 {
			e.SummonCooldown = 60 // 1 сек до первого призыва
		}
	}

	if e.SummonCooldown > 0 {
		e.SummonCooldown--
	}

	// --- Призыв ---
	if e.SummonCooldown == 0 {
		count := stats.SummonCountMin
		if stats.SummonCountMax > stats.SummonCountMin {
			count += w.Rng.Intn(stats.SummonCountMax - stats.SummonCountMin + 1)
		}
		spawnGhosts(w, e, count)
		e.SummonCooldown = stats.SummonCooldown
	}

	// --- Движение ---
	// Идёт к игроку (как обычный враг).
	if hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
		return
	}
	moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
}

func spawnGhosts(w *world.World, summoner *world.Enemy, count int) {
	stats := summoner.Type.Stats()
	scx, scy := summoner.Center()

	for i := 0; i < count; i++ {
		// Ищем свободную точку вокруг призывателя.
		x, y := scx, scy
		for attempt := 0; attempt < 8; attempt++ {
			angle := w.Rng.Float64() * math.Pi * 2
			r := stats.SummonRange * (0.5 + w.Rng.Float64()*0.5)
			tx := scx + math.Cos(angle)*r
			ty := scy + math.Sin(angle)*r
			// Ghost проходит сквозь стены, поэтому точки в стене допустимы.
			if tx < 40 || tx > world.WorldWidth-40 ||
				ty < 40 || ty > world.WorldHeight-40 {
				continue
			}
			x, y = tx, ty
			break
		}

		g := world.NewEnemy(world.EnemyGhost, x, y)
		g.OwnerSummonerID = summoner.SummonerID
		w.Enemies = append(w.Enemies, g)
	}
}
