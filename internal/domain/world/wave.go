package world

import "Sqwave/internal/domain/geometry"

type WaveComposition struct {
	Infantry int
	Shooter  int
	Scout    int
	Medic    int
	Rammer   int
	Sniper   int
}

func (c WaveComposition) Total() int {
	return c.Infantry + c.Shooter + c.Scout + c.Medic + c.Rammer + c.Sniper
}

type WaveState struct {
	Number      int
	Composition WaveComposition

	PoolInfantry int
	PoolShooter  int
	PoolScout    int
	PoolMedic    int
	PoolRammer   int
	PoolSniper   int

	GroupID         int
	GroupRemaining  int
	GroupSpawnTimer int
	GroupOriginX    float64
	GroupOriginY    float64

	SpawnedCount int
	KilledCount  int

	PauseTimer int
	Active     bool

	WaveSize int
}

func (ws *WaveState) TotalPoolRemaining() int {
	return ws.PoolInfantry + ws.PoolShooter + ws.PoolScout +
		ws.PoolMedic + ws.PoolRammer + ws.PoolSniper
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

func RandomEnemyType(rng intRng) EnemyType {
	switch rng.Intn(6) {
	case 0:
		return EnemyInfantry
	case 1:
		return EnemyShooter
	case 2:
		return EnemyScout
	case 3:
		return EnemyMedic
	case 4:
		return EnemyRammer
	default:
		return EnemySniper
	}
}

// WaveCompositionFor — состав волны по номеру.
func WaveCompositionFor(n, total int, rng intRng) WaveComposition {
	infW := 30 + rng.Intn(41)

	shW := 0
	if rng.Intn(100) < 70 {
		shW = 10 + rng.Intn(21)
	}

	scW := 0
	if rng.Intn(100) < 50 {
		scW = 5 + rng.Intn(16)
	}

	mdW := 0
	if rng.Intn(100) < 30 {
		mdW = 2 + rng.Intn(9)
	}

	rmW := 0
	if rng.Intn(100) < 25 {
		rmW = 3 + rng.Intn(8)
	}

	snW := 0
	if rng.Intn(100) < 20 {
		snW = 2 + rng.Intn(5)
	}

	sum := infW + shW + scW + mdW + rmW + snW
	if sum == 0 {
		return WaveComposition{Infantry: total}
	}

	inf := total * infW / sum
	sh := total * shW / sum
	sc := total * scW / sum
	md := total * mdW / sum
	rm := total - inf - sh - sc - md
	sn := total - inf - sh - sc - md - rm
	if sn < 0 {
		sn = 0
	}

	diff := total - (inf + sh + sc + md + rm + sn)
	inf += diff
	if inf < 0 {
		inf = 0
	}

	return WaveComposition{
		Infantry: inf,
		Shooter:  sh,
		Scout:    sc,
		Medic:    md,
		Rammer:   rm,
		Sniper:   sn,
	}
}

// SpawnPoints — базовые точки для выбора «центра» группы.
// Используются в systems.pickGroupOrigin: сначала берётся одна
// из этих точек, потом добавляется случайный разброс ±80.
// SpawnPoints — базовые точки для выбора центра группы.
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
