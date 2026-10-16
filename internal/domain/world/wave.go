package world

import "Sqwave/internal/domain/geometry"

type WaveKind int

const (
	WaveKindCommon WaveKind = iota
	WaveKindTimer
	WaveKindDefend
)

// Ограничения на поддержку.
const (
	MaxSupportsPerGroup = 2
	MaxSupportsPerWave  = 3
)

type WaveComposition struct {
	Infantry      int
	Shooter       int
	Scout         int
	Medic         int
	Rammer        int
	Sniper        int
	Rocketeer     int
	Buffer        int
	Fleer         int
	Engineer      int
	Flamer        int
	Protector     int
	Tank          int
	MachineGunner int
	Summoner      int
	Harpooner     int
}

func (c WaveComposition) Total() int {
	return c.Infantry + c.Shooter + c.Scout + c.Medic +
		c.Rammer + c.Sniper + c.Rocketeer + c.Buffer +
		c.Fleer + c.Engineer + c.Flamer + c.Protector +
		c.Tank + c.MachineGunner + c.Summoner + c.Harpooner
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

	WaveLevel       int
	GroupsTotal     int
	GroupsRemaining int

	BaseWaveSize int
	WavesLimit   int

	Kind WaveKind

	// Лимиты поддержки.
	WaveSupports  int // сколько поддержки уже заспавнено в этой волне
	GroupSupports int // сколько поддержки уже заспавнено в текущей группе
	// Лимит снайперов за волну.
	WaveSnipers    int
	MaxWaveSnipers int

	PauseTimer int
	Active     bool

	// Timer-волна.
	TimerTotal         int
	TimerRemaining     int
	TimerGroupCooldown int // тиков до следующей группы

	// Defend-волна.
	DefendSpawnCount  int  // сколько спавнов в этой волне
	DefendSpawnsAlive int  // сколько ещё живо
	DefendOutcome     int  // 0=в процессе, 1=победа, -1=поражение
	DefendPenalty     bool // штрафная группа уже спавнена
}

func (ws *WaveState) IsVictory() bool {
	return ws.WavesLimit > 0 && ws.Number >= ws.WavesLimit && !ws.Active
}

type intRng interface{ Intn(int) int }

// IsSupportType — является ли враг поддержкой (лечит/баффает/защищает).
func IsSupportType(t EnemyType) bool {
	switch t {
	case EnemyMedic, EnemyBuffer, EnemyProtector:
		return true
	}
	return false
}

// RandomWaveLevel — случайный уровень волны (1..5) в зависимости
// от номера волны. Ранние волны — в основном I–II, поздние — IV–V.
// Уже на первой волне есть шанс увидеть V (~1%).
func RandomWaveLevel(waveNumber int, rng intRng) int {
	if waveNumber < 1 {
		waveNumber = 1
	}

	// Базовые веса для волны 1.
	base := [5]int{100, 60, 30, 10, 2}
	// Рост весов за каждую волну.
	growth := [5]int{-15, -5, 5, 15, 20}

	var w [5]int
	total := 0
	for i := 0; i < 5; i++ {
		v := base[i] + (waveNumber-1)*growth[i]
		if v < 1 {
			v = 1
		}
		w[i] = v
		total += v
	}

	roll := rng.Intn(total)
	cum := 0
	for i := 0; i < 5; i++ {
		cum += w[i]
		if roll < cum {
			return i + 1
		}
	}
	return 1
}

// WaveSizeFor — количество врагов в волне по уровню.
func WaveSizeFor(level, base int, rng intRng) int {
	if base <= 0 {
		base = 100
	}
	offset := 0
	spread := 20
	switch level {
	case 1:
		offset = 0
	case 2:
		offset = 20
	case 3:
		offset = 40
	case 4:
		offset = 60
		spread = 40
	case 5:
		offset = -50
		spread = 50
	}
	min := base + offset
	if min < 20 {
		min = 20
	}
	return min + rng.Intn(spread+1)
}

// GroupCountFor — количество групп в волне по уровню.
func GroupCountFor(level int, rng intRng) int {
	switch level {
	case 1, 2:
		return 8 + rng.Intn(3) // 8..10
	case 3, 4:
		return 5 + rng.Intn(4) // 5..8
	case 5:
		return 1
	}
	return 8
}

// enemyTiers — враги по уровням (1..4).
var enemyTiers = map[int][]EnemyType{
	1: {EnemyInfantry, EnemyShooter, EnemyScout, EnemyMedic},
	2: {EnemyRammer, EnemySniper, EnemyRocketeer, EnemyBuffer},
	3: {EnemyFleer, EnemyEngineer, EnemyFlamer, EnemyProtector},
	4: {EnemyTank, EnemyMachineGunner, EnemySummoner, EnemyHarpooner},
}

type tierWeights struct {
	L1, L2, L3, L4 int
}

// waveTierWeights — веса уровней врагов для уровня волны.
var waveTierWeights = map[int]tierWeights{
	1: {85, 15, 0, 0},
	2: {40, 45, 15, 0},
	3: {5, 45, 45, 5},
	4: {30, 30, 30, 10},
	5: {5, 10, 45, 40},
}

// RandomEnemyTypeForWave — случайный враг по уровню волны.
// allowSupport = false — поддержка исключается (перебрасывается).
func RandomEnemyTypeForWave(level int, allowSupport bool, rng intRng) EnemyType {
	w, ok := waveTierWeights[level]
	if !ok {
		w = waveTierWeights[1]
	}
	total := w.L1 + w.L2 + w.L3 + w.L4
	if total <= 0 {
		return EnemyInfantry
	}

	for attempt := 0; attempt < 8; attempt++ {
		roll := rng.Intn(total)
		var tier int
		switch {
		case roll < w.L1:
			tier = 1
		case roll < w.L1+w.L2:
			tier = 2
		case roll < w.L1+w.L2+w.L3:
			tier = 3
		default:
			tier = 4
		}
		pool := enemyTiers[tier]
		t := pool[rng.Intn(len(pool))]
		if allowSupport || !IsSupportType(t) {
			return t
		}
	}
	// Не смогли избежать поддержки — пехотинец.
	return EnemyInfantry
}

// TrackForLevel — путь к треку для уровня волны.
func TrackForLevel(level int) string {
	switch level {
	case 1, 2:
		return "assets/music/all_systems_go.wav"
	case 3, 4:
		return "assets/music/aesthetic_failure.wav"
	case 5:
		return "assets/music/elfire.wav"
	}
	return ""
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

// TimerTrackForLevel — путь к треку для timer-волны по уровню.
func TimerTrackForLevel(level int) string {
	switch level {
	case 1, 2:
		return "assets/music/experiment_7.wav"
	case 3, 4:
		return "assets/music/bounce_automation.wav"
	case 5:
		return "assets/music/cultivating_madness.wav"
	}
	return ""
}

// TimerGroupIntervalFor — интервал между группами в тиках.
func TimerGroupIntervalFor(level int) int {
	switch level {
	case 1, 2:
		return 30 * 60 // 30 сек
	case 3, 4:
		return 25 * 60 // 25 сек
	case 5:
		return 20 * 60 // 20 сек
	}
	return 30 * 60
}

// TimerGroupSizeRangeFor — диапазон размера группы.
func TimerGroupSizeRangeFor(level int) (int, int) {
	switch level {
	case 1:
		return 20, 40
	case 2, 3:
		return 25, 50
	case 4, 5:
		return 30, 60
	}
	return 20, 40
}

// DefendSpawnCountFor — сколько спавнов защищать игроку.
func DefendSpawnCountFor(level int) int {
	switch level {
	case 1, 2:
		return 1
	case 3, 4:
		return 2
	case 5:
		return 3
	}
	return 1
}

// DefendSpawnHPFor — HP одного спавна.
func DefendSpawnHPFor(level int) int {
	switch level {
	case 1, 2:
		return 600
	case 3, 4:
		return 900
	case 5:
		return 1200
	}
	return 600
}

// DefendTrackForLevel — трек для defend-волны.
func DefendTrackForLevel(level int) string {
	switch level {
	case 1, 2:
		// Два трека на выбор — 50/50.
		return "assets/music/necro_syncrosis.wav" // выбор на месте
	case 3, 4:
		return "assets/music/twisted.wav"
	case 5:
		return "assets/music/dark_subconscious.wav"
	}
	return ""
}

// SniperLimitFor — сколько снайперов максимум за волну.
// Возвращает 3..5 случайно.
func SniperLimitFor(rng intRng) int {
	return 3 + rng.Intn(3) // 3, 4 или 5
}
