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
}

func (c WaveComposition) Total() int {
	return c.Infantry + c.Shooter + c.Scout + c.Medic +
		c.Rammer + c.Sniper + c.Rocketeer
}

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

	// Уровни волн.
	MaxWaveLevel    int // из настроек, 1..10
	WaveLevel       int // текущий уровень (1..MaxWaveLevel)
	GroupsTotal     int // сколько групп в текущей волне
	GroupsRemaining int // сколько групп ещё не запущено

	// Базовые параметры из настроек.
	BaseWaveSize int // базовое количество врагов за волну
	WavesLimit   int // 0 = бесконечные

	PauseTimer int
	Active     bool
}

// IsVictory — все волны пройдены в не-бесконечном режиме.
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
	switch rng.Intn(7) {
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
	default:
		return EnemyRocketeer
	}
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
