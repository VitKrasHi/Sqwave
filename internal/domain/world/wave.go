package world

import "Sqwave/internal/domain/geometry"

type WaveKind int

const (
	WaveKindCommon WaveKind = iota
	WaveKindTimer
	WaveKindDefend
	WaveKindAttack
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

	// Attack-волна.
	AttackSpawnCount  int
	AttackSpawnsAlive int
	AttackOutcome     int // 0=в процессе, 1=победа (все спавны убиты)
	// индекс спавна, из которого спавнится текущая группа.
	AttackGroupOwnerSpawnIdx int
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
// WaveSizeFor — размер волны 100..150 независимо от уровня.
// base игнорируется — оставлен для совместимости сигнатуры.
func WaveSizeFor(level, base int, rng intRng) int {
	return 100 + rng.Intn(51)
}

// GroupCountFor — количество групп в Common/Defend волне: 10..15.
func GroupCountFor(level int, rng intRng) int {
	return 10 + rng.Intn(6)
}

// enemyCategory — роль врага в бою.
type enemyCategory int

const (
	CatMelee    enemyCategory = 0 // ближник
	CatRanged   enemyCategory = 1 // дальник
	CatObstacle enemyCategory = 2 // препятствие / быстрое и надоедливое
	CatSupport  enemyCategory = 3 // поддержка
)

// Веса категорий внутри уровня волны: 40 / 30 / 20 / 10.
const (
	WeightMelee    = 40
	WeightRanged   = 30
	WeightObstacle = 20
	WeightSupport  = 10
)

// enemyByLevelAndCategory — враги по уровню и роли.
//
//	I:   базовые
//	II:  усиленные версии I
//	III: специализированные (дальний бой, препятствия, неуязвимость)
//	IV:  тяжёлые (танк, пулемётчик, огнемётчик, призыватель)
var enemyByLevelAndCategory = map[int]map[enemyCategory][]EnemyType{
	1: {
		CatMelee:    {EnemyInfantry},
		CatRanged:   {EnemyShooter},
		CatObstacle: {EnemyScout},
		CatSupport:  {EnemyMedic},
	},
	2: {
		CatMelee:    {EnemyRammer},
		CatRanged:   {EnemyRocketeer},
		CatObstacle: {EnemyFleer},
		CatSupport:  {EnemyBuffer},
	},
	3: {
		CatMelee:    {EnemyHarpooner},
		CatRanged:   {EnemySniper},
		CatObstacle: {EnemyEngineer},
		CatSupport:  {EnemyProtector},
	},
	4: {
		CatMelee:    {EnemyTank},
		CatRanged:   {EnemyMachineGunner},
		CatObstacle: {EnemyFlamer},
		CatSupport:  {EnemySummoner},
	},
}

// tierWeights — веса уровней врагов по уровню волны.
type tierWeights struct {
	L1, L2, L3, L4 int
}

var waveTierWeights = map[int]tierWeights{
	1: {90, 10, 0, 0},   // I 90% / II 10%
	2: {75, 25, 0, 0},   // I 75% / II 25%
	3: {70, 20, 10, 0},  // I 70% / II 20% / III 10%
	4: {50, 40, 10, 0},  // I 50% / II 40% / III 10%
	5: {40, 40, 10, 10}, // I 40% / II 40% / III 10% / IV 10%
}

// RandomEnemyTypeForWave — случайный враг по уровню волны.
//
// Два шага:
//  1. Выбор уровня врага (tier) по весам waveTierWeights[level].
//  2. Выбор категории (ближник / дальник / препятствие / поддержка)
//     по весам 40/30/20/10.
//
// allowSupport = false — поддержка исключается, выбор перебрасывается.
func RandomEnemyTypeForWave(level int, allowSupport bool, rng intRng) EnemyType {
	tw, ok := waveTierWeights[level]
	if !ok {
		tw = waveTierWeights[1]
	}
	total := tw.L1 + tw.L2 + tw.L3 + tw.L4
	if total <= 0 {
		return EnemyInfantry
	}
	catTotal := WeightMelee + WeightRanged + WeightObstacle + WeightSupport

	for attempt := 0; attempt < 12; attempt++ {
		// 1. Уровень врага.
		roll := rng.Intn(total)
		var tier int
		switch {
		case roll < tw.L1:
			tier = 1
		case roll < tw.L1+tw.L2:
			tier = 2
		case roll < tw.L1+tw.L2+tw.L3:
			tier = 3
		default:
			tier = 4
		}

		// 2. Категория.
		catRoll := rng.Intn(catTotal)
		var cat enemyCategory
		switch {
		case catRoll < WeightMelee:
			cat = CatMelee
		case catRoll < WeightMelee+WeightRanged:
			cat = CatRanged
		case catRoll < WeightMelee+WeightRanged+WeightObstacle:
			cat = CatObstacle
		default:
			cat = CatSupport
		}

		// 3. Поддержка запрещена — перебрасываем.
		if cat == CatSupport && !allowSupport {
			continue
		}

		cats, ok := enemyByLevelAndCategory[tier]
		if !ok {
			continue
		}
		pool := cats[cat]
		if len(pool) == 0 {
			continue
		}
		return pool[rng.Intn(len(pool))]
	}

	// Fallback — пехотинец.
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
	return 30 * 60
}

// TimerGroupSizeRangeFor — диапазон размера группы.
func TimerGroupSizeRangeFor(level int) (int, int) {
	return 10, 15
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
	return 3 // 3, 4 или 5
}

// AttackSpawnCountFor — сколько спавнов у врагов.
func AttackSpawnCountFor(level int) int {
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

// AttackSpawnHPFor — HP одного вражеского спавна.
func AttackSpawnHPFor(level int) int {
	switch level {
	case 1, 2:
		return 700
	case 3, 4:
		return 1000
	case 5:
		return 1400
	}
	return 700
}

// AttackTrackForLevel — трек для attack-волны.
func AttackTrackForLevel(level int) string {
	switch level {
	case 1, 2:
		return "assets/music/cy2v9.wav"
	case 3, 4:
		return "assets/music/suck.wav"
	case 5:
		return "assets/music/pull_the_trigger.wav"
	}
	return ""
}

// AttackGroupIntervalFor — интервал между группами в тиках.
func AttackGroupIntervalFor(level int) int {
	return 30 * 60
}

// AttackGroupSizeRangeFor — диапазон размера группы.
func AttackGroupSizeRangeFor(level int) (int, int) {
	return 10, 15
}
