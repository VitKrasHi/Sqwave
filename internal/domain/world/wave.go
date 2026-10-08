package world

import "Sqwave/internal/domain/geometry"

type WaveComposition struct {
	Infantry int
	Shooter  int
	Scout    int
	Medic    int
}

func (c WaveComposition) Total() int {
	return c.Infantry + c.Shooter + c.Scout + c.Medic
}

type WaveState struct {
	Number      int
	Composition WaveComposition

	PoolInfantry int
	PoolShooter  int
	PoolScout    int
	PoolMedic    int

	GroupID         int
	GroupRemaining  int
	GroupSpawnTimer int
	GroupOriginX    float64
	GroupOriginY    float64

	SpawnedCount int
	KilledCount  int

	PauseTimer int
	Active     bool
}

func (ws *WaveState) TotalPoolRemaining() int {
	return ws.PoolInfantry + ws.PoolShooter + ws.PoolScout + ws.PoolMedic
}

// intRng — минимальный интерфейс для генератора случайных чисел.
// *rand.Rand удовлетворяет ему.
type intRng interface{ Intn(int) int }

// WaveSizeFor — размер волны N (100..200).
func WaveSizeFor(n int, rng intRng) int {
	base := 100
	if n > 1 {
		base += (n - 1) * 15
	}
	if base > 200 {
		base = 200
	}
	spread := base / 6
	return base - spread + rng.Intn(spread*2+1)
}

// WaveCompositionFor — состав волны по номеру.
func WaveCompositionFor(n, total int, rng intRng) WaveComposition {
	// Вес каждого типа — случайный в своём диапазоне.
	// Диапазон подобран так, чтобы:
	//   - Пехотинец всегда есть, но его доля 30..70%.
	//   - Стрелок появляется с шансом ~70%.
	//   - Разведчик появляется с шансом ~50%.
	//   - Медик появляется с шансом ~30%.
	infW := 30 + rng.Intn(41) // 30..70

	shW := 0
	if rng.Intn(100) < 70 {
		shW = 10 + rng.Intn(21) // 10..30
	}

	scW := 0
	if rng.Intn(100) < 50 {
		scW = 5 + rng.Intn(16) // 5..20
	}

	mdW := 0
	if rng.Intn(100) < 30 {
		mdW = 2 + rng.Intn(9) // 2..10
	}

	sum := infW + shW + scW + mdW
	if sum == 0 {
		return WaveComposition{Infantry: total}
	}

	inf := total * infW / sum
	sh := total * shW / sum
	sc := total * scW / sum
	md := total - inf - sh - sc
	if md < 0 {
		md = 0
	}

	// Компенсация округлений.
	diff := total - (inf + sh + sc + md)
	inf += diff
	if inf < 0 {
		inf = 0
	}

	return WaveComposition{
		Infantry: inf,
		Shooter:  sh,
		Scout:    sc,
		Medic:    md,
	}
}

// SpawnPoints — базовые точки для выбора «центра» группы.
// Используются в systems.pickGroupOrigin: сначала берётся одна
// из этих точек, потом добавляется случайный разброс ±80.
var SpawnPoints = []geometry.Point{
	{X: 1050, Y: 700},
	{X: 2150, Y: 700},
	{X: 1050, Y: 1700},
	{X: 2150, Y: 1700},
	{X: 1650, Y: 500},
	{X: 1650, Y: 1900},
	{X: 800, Y: 1200},
	{X: 2400, Y: 1200},
}
