package world

import "Sqwave/internal/domain/geometry"

type WaveComposition struct {
	Infantry  int
	Shooter   int
	Scout     int
	Medic     int
	Rammer    int
	Sniper    int
	Rocketeer int
	Buffer    int
	Fleer     int
}

func (c WaveComposition) Total() int {
	return c.Infantry + c.Shooter + c.Scout + c.Medic +
		c.Rammer + c.Sniper + c.Rocketeer + c.Buffer + c.Fleer
}

type WaveKind int

const (
	WaveKindCommon WaveKind = iota
	WaveKindTimer
)

type WaveState struct {
	Number int

	Composition WaveComposition

	GroupID         int
	GroupRemaining  int
	GroupSpawnTimer int
	GroupOriginX    float64
	GroupOriginY    float64

	SpawnedCount int
	KilledCount  int

	WaveSize int

	// Common-волна.
	MaxWaveLevel    int
	WaveLevel       int
	GroupsTotal     int
	GroupsRemaining int

	// Настройки.
	BaseWaveSize int
	WavesLimit   int

	// Тип волны.
	Kind WaveKind

	// Timer-волна.
	TimerTotal     int // всего тиков
	TimerRemaining int // осталось тиков

	PauseTimer int
	Active     bool
}

func (ws *WaveState) IsVictory() bool {
	return ws.WavesLimit > 0 && ws.Number >= ws.WavesLimit && !ws.Active
}

// intRng — минимальный интерфейс для генератора случайных чисел.
type intRng interface{ Intn(int) int }

// WaveSizeFor — размер волны N.
func WaveSizeFor(n, base int, rng intRng) int {
	if base <= 0 {
		base = 100
	}
	size := base
	if n > 1 {
		size += (n - 1) * 15
	}
	maxSize := base * 2
	if size > maxSize {
		size = maxSize
	}
	spread := size / 6
	if spread < 1 {
		spread = 1
	}
	return size - spread + rng.Intn(spread*2+1)
}

// RandomEnemyType — равновероятный выбор типа врага.
func RandomEnemyType(rng intRng) EnemyType {
	switch rng.Intn(9) {
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
	case 5:
		return EnemySniper
	case 6:
		return EnemyRocketeer
	case 7:
		return EnemyBuffer
	default:
		return EnemyFleer
	}
}

// RandomEnemyTypeTimer — выбор типа врага для timer-волны.
// Чем больше progress (0..1), тем шире пул типов: от Infantry/Shooter
// до всех семи.
func RandomEnemyTypeTimer(progress float64, rng intRng) EnemyType {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	pool := []EnemyType{EnemyInfantry, EnemyShooter}
	if progress > 1 {
		pool = append(pool, EnemyScout)
	}
	if progress > 1 {
		pool = append(pool, EnemyMedic)
	}
	if progress > 1 {
		pool = append(pool, EnemyRammer)
	}
	if progress > 1 {
		pool = append(pool, EnemySniper)
	}
	if progress > 1 {
		pool = append(pool, EnemyRocketeer)
	}
	if progress > 1 {
		pool = append(pool, EnemyBuffer)
	}
	if progress > 1 {
		pool = append(pool, EnemyFleer)
	}

	return pool[rng.Intn(len(pool))]
}

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
