package world

import "Sqwave/internal/domain/geometry"

type WaveComposition struct {
	Infantry int
	Shooter  int
	Scout    int
}

func (c WaveComposition) Total() int {
	return c.Infantry + c.Shooter + c.Scout
}

type WaveState struct {
	Number      int
	Composition WaveComposition
	Queue       []EnemyType

	SpawnTimer    int // тик внутри группы
	SpawnInterval int // интервал между врагами в группе

	GroupSize       int // размер группы
	GroupRemaining  int // сколько ещё в текущей группе
	GroupPauseTimer int // пауза между группами
	GroupPauseTicks int // длительность паузы

	PauseTimer int
	Active     bool
	AllSpawned bool
}

func (ws *WaveState) SpawnedCount() int {
	return ws.Composition.Total() - len(ws.Queue)
}

func WaveCompositionFor(n int) WaveComposition {
	switch n {
	case 1:
		return WaveComposition{Infantry: 40}
	case 2:
		return WaveComposition{Infantry: 40, Shooter: 10}
	case 3:
		return WaveComposition{Infantry: 45, Shooter: 15, Scout: 5}
	case 4:
		return WaveComposition{Infantry: 50, Shooter: 20, Scout: 10}
	case 5:
		return WaveComposition{Infantry: 55, Shooter: 25, Scout: 15}
	default:
		extra := n - 5
		return WaveComposition{
			Infantry: 55 + extra*5,
			Shooter:  25 + extra*4,
			Scout:    15 + extra*3,
		}
	}
}

type intRng interface{ Intn(int) int }

func QueueFor(c WaveComposition, rng intRng) []EnemyType {
	type bucket struct {
		t     EnemyType
		count int
	}
	buckets := []bucket{
		{EnemyInfantry, c.Infantry},
		{EnemyShooter, c.Shooter},
		{EnemyScout, c.Scout},
	}
	total := c.Total()
	queue := make([]EnemyType, 0, total)
	for i := 0; i < total; i++ {
		remaining := 0
		for _, b := range buckets {
			if b.count > 0 {
				remaining++
			}
		}
		if remaining == 0 {
			break
		}
		pick := rng.Intn(remaining)
		for j := range buckets {
			if buckets[j].count == 0 {
				continue
			}
			if pick == 0 {
				queue = append(queue, buckets[j].t)
				buckets[j].count--
				break
			}
			pick--
		}
	}
	return queue
}

// SpawnPoints — заранее выбранные точки вне арены и вне стен.
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
