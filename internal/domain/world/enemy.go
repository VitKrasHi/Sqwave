package world

import "Sqwave/internal/domain/geometry"

type EnemyType int

const (
	EnemyInfantry EnemyType = iota
	EnemyShooter
	EnemyScout
	EnemyMedic
	EnemyRammer
	EnemySniper
	EnemyRocketeer
	EnemyBuffer
	EnemyFleer
	EnemyEngineer
	EnemyFlamer
	EnemyProtector
	EnemyTank
	EnemyMachineGunner
	EnemySummoner
	EnemyGhost
	EnemyHarpooner
)

type EnemyStats struct {
	Name   string
	MaxHP  int
	SP     int
	Damage int
	Size   float64

	MeleeRange        float64
	AttackCooldownMin int
	AttackCooldownMax int
	SwingDuration     int
	MeleeTriggerRange float64

	PreferredMin     float64
	PreferredMax     float64
	ProjectileSpeed  float64
	ProjectileSize   float64
	ProjectileLife   int
	ProjectileCount  int     // пеллет за выстрел (дробовик разведчика = 8)
	ProjectileSpread float64 // полный угол разброса в градусах

	// Таран: длинный рывок с ускорением.
	DashSpeedMul    float64
	DashDuration    int
	DashCooldownMin int
	DashCooldownMax int
	DashDamage      int

	SniperAimDuration int
	SniperIdleMin     int
	SniperIdleMax     int
	SniperDamageMin   int
	SniperDamageMax   int

	// Ракетчик.
	RocketSpeed           float64
	RocketHomingRate      float64
	RocketLife            int
	RocketExplosionDamage int
	RocketExplosionRadius float64
	RocketExplosionLife   int
	RocketFireRange       float64
	RocketCooldown        int
	RocketSize            float64

	// Буфер.
	BufferRange        float64
	BufferMaxTargets   int
	BufferDamageMul    float64
	BufferSpeedMul     float64
	BufferDuration     int
	BufferRefreshTicks int

	// Беглец.
	FleerTurnRate  float64 // радиан за тик, медленный поворот
	FleerConeWidth float64 // полный угол конуса в градусах
	FleerGrabRange float64 // радиус захвата
	FleerLatchDmg  int     // урон в тик при захвате

	// Инженер.
	TurretCooldownMin int
	TurretCooldownMax int
	TurretMaxActive   int
	TurretLife        int
	TurretHP          int
	TurretSize        float64
	TurretDetectRange float64
	TurretProjSpeed   float64
	TurretProjSize    float64
	TurretProjLife    int
	TurretProjDamage  int
	TurretBurstCount  int
	TurretCooldown    int // кулдаун между очередями
	TurretShotGap     int // тиков между выстрелами в очереди
	TurretSetupTime   int // сколько тиков ставится турель (600 = 10 сек)
	SetupMoveTime     int // сколько тиков даётся на подход к точке (180 = 3 сек)
	FleeDuration      int // сколько тиков убегать после ранения

	// Огнемётчик.
	FlameDamage      int     // разовый урон от частицы
	FlameBurnDamage  int     // урон догорания за тик (накапливается)
	FlameBurnTicks   int     // длительность догорания в тиках
	FlameBurnRate    int     // раз в сколько тиков наносится 1 урона догорания
	WeaponTurnRate   float64 // рад/тик
	WeaponArcDegrees float64 // допустимое отклонение от цели для стрельбы

	// Протектор.
	ProtectRange      float64
	ProtectMaxTargets int
	ProtectDuration   int

	// Танк.
	ShieldArcDegrees   float64
	HammerWindup       int
	HammerRange        float64
	HammerTriggerRange float64
	HammerDamage       int
	HammerCooldown     int
	SupportCoverRange  float64

	// Пулемётчик.
	MGMaxDamage   int     // урон в упор
	MGMinDamage   int     // урон на максимальной дистанции
	MGBurstCount  int     // выстрелов в очереди
	MGBurstGap    int     // тиков между выстрелами в очереди
	MGCooldown    int     // тиков между очередями
	MGPredictLead float64 // 0..1, доля опережения по скорости игрока
	MGSpreadDeg   float64 // разброс в градусах на выстрел

	// Призыватель.
	SummonCooldown int
	SummonCountMin int
	SummonCountMax int
	SummonRange    float64

	// Призрак.
	GhostContactDamage int
	GhostContactRange  float64
	GhostContactCool   int

	// Гарпунист.
	HarpoonAimTime    int
	HarpoonSpeed      float64
	HarpoonSize       float64
	HarpoonLife       int
	HarpoonHP         int
	HarpoonPullSpeed  float64
	HarpoonCooldown   int
	HarpoonAttachDist float64
}

var enemyStats = map[EnemyType]EnemyStats{
	EnemyInfantry: {
		Name:              "Infantry",
		MaxHP:             120,
		SP:                150,
		Damage:            15,
		Size:              24,
		MeleeRange:        60,
		MeleeTriggerRange: 140,
		AttackCooldownMin: 24, // 400 мс
		AttackCooldownMax: 60, // 1 сек
		SwingDuration:     7,
	},
	EnemyShooter: {
		Name:              "Shooter",
		MaxHP:             100,
		SP:                80,
		Damage:            10,
		Size:              22,
		MeleeRange:        0,
		AttackCooldownMin: 90,
		AttackCooldownMax: 150,
		PreferredMin:      180,
		PreferredMax:      340,
		ProjectileSpeed:   12, // было 4 — почти вдвое быстрее
		ProjectileSize:    8,
		ProjectileLife:    150, // было 150 — короче, чтобы не летели через пол-карты
	},
	EnemyScout: {
		Name:              "Scout",
		MaxHP:             70,
		SP:                200,
		Damage:            5, // на пеллету, 8 пеллет — до 80 в упор
		Size:              20,
		MeleeRange:        60, // дистанция выстрела дробью
		AttackCooldownMin: 50,
		AttackCooldownMax: 90,
		ProjectileSpeed:   9,
		ProjectileSize:    4,
		ProjectileLife:    20, // короткая жизнь — дробь летит недалеко
		ProjectileCount:   8,
		ProjectileSpread:  30,
	},
	EnemyMedic: {
		Name:              "Medic",
		MaxHP:             100,
		SP:                120,
		Damage:            0,
		Size:              22,
		PreferredMin:      80,  // держится не ближе этого от игрока
		PreferredMax:      200, // и не дальше этого от союзников
		AttackCooldownMin: 0,
		AttackCooldownMax: 0,
	},
	EnemyRammer: {
		Name:              "Rammer",
		MaxHP:             200,
		SP:                100,
		Damage:            25, // урон щитом в ближнем бою
		Size:              28,
		MeleeRange:        70,
		AttackCooldownMin: 40,
		AttackCooldownMax: 70,
		SwingDuration:     10,
		DashSpeedMul:      4.0,
		DashDuration:      50,
		DashCooldownMin:   60,
		DashCooldownMax:   140,
		DashDamage:        50, // ×2 от базового 50
	},
	EnemySniper: {
		Name:              "Sniper",
		MaxHP:             100,
		SP:                120,
		Size:              20,
		PreferredMin:      350, // радиус паники
		SniperAimDuration: 45,
		SniperIdleMin:     60,
		SniperIdleMax:     300,
		SniperDamageMin:   50,
		SniperDamageMax:   50,
	},
	EnemyRocketeer: {
		Name:                  "Rocketeer",
		MaxHP:                 300,
		SP:                    70,
		Size:                  26,
		RocketSpeed:           5.5,
		RocketHomingRate:      0.04,
		RocketLife:            180, // меньше — ракета взрывается быстрее
		RocketExplosionDamage: 50,
		RocketExplosionRadius: 120,
		RocketExplosionLife:   14,
		RocketFireRange:       420,
		RocketCooldown:        150,
		RocketSize:            10,
	},
	EnemyBuffer: {
		Name:               "Buffer",
		MaxHP:              120,
		SP:                 100,
		Size:               22,
		PreferredMin:       80,
		PreferredMax:       200,
		BufferRange:        260,
		BufferMaxTargets:   5,
		BufferDamageMul:    1.4,
		BufferSpeedMul:     1.3,
		BufferDuration:     30,
		BufferRefreshTicks: 6,
	},
	EnemyFleer: {
		Name:           "Fleer",
		MaxHP:          200,
		SP:             200,
		Size:           22,
		MeleeRange:     55,    // дистанция захвата
		FleerTurnRate:  0.035, // ~2° за тик, полный оборот ~3 сек
		FleerConeWidth: 40,    // узкий конус (полный угол)
		FleerGrabRange: 55,
		FleerLatchDmg:  0, // HP за тик при захвате
	},
	EnemyEngineer: {
		Name:              "Engineer",
		MaxHP:             150,
		SP:                120,
		Size:              22,
		Damage:            25,
		PreferredMin:      150,
		PreferredMax:      260,
		AttackCooldownMin: 90,
		AttackCooldownMax: 140,
		ProjectileSpeed:   8,
		ProjectileSize:    5,
		ProjectileLife:    40,
		ProjectileCount:   4,
		ProjectileSpread:  28,

		TurretSetupTime: 600, // 10 сек
		SetupMoveTime:   180, // 3 сек на подход
		FleeDuration:    120, // 2 сек убегания

		TurretMaxActive:   6,
		TurretLife:        1200,
		TurretHP:          60,
		TurretSize:        20,
		TurretDetectRange: 550,
		TurretProjSpeed:   15,
		TurretProjSize:    5,
		TurretProjLife:    60,
		TurretProjDamage:  15,
		TurretBurstCount:  3,
		TurretCooldown:    120,
		TurretShotGap:     10,
	},
	EnemyFlamer: {
		Name:              "Flamer",
		MaxHP:             300,
		SP:                100,
		Size:              24,
		PreferredMin:      60,
		PreferredMax:      220,
		AttackCooldownMin: 3,
		AttackCooldownMax: 6,
		ProjectileSpeed:   7,
		ProjectileSize:    10,
		ProjectileLife:    35,
		ProjectileCount:   3,
		ProjectileSpread:  45,

		FlameDamage:      5,
		FlameBurnDamage:  1,
		FlameBurnTicks:   600,   // 10 сек
		FlameBurnRate:    60,    // 1 урон раз в секунду
		WeaponTurnRate:   0.035, // ~2°/тик, полный оборот ~3 сек
		WeaponArcDegrees: 20,    // стреляет, если оружие в 25° от игрока
	},
	EnemyProtector: {
		Name:         "Protector",
		MaxHP:        150,
		SP:           150,
		Size:         22,
		PreferredMin: 100,
		PreferredMax: 240,

		ProtectRange:      300,
		ProtectMaxTargets: 3,
		ProtectDuration:   20, // 0.33 сек без обновления — купол гаснет
	},
	EnemyTank: {
		Name:         "Tank",
		MaxHP:        500,
		SP:           50,
		Size:         38,
		PreferredMin: 80,
		PreferredMax: 200,

		WeaponTurnRate: 0.018, // ~1°/тик, полный оборот ~5.8 сек

		ShieldArcDegrees:   140,
		HammerWindup:       45,  // 0.75 сек замах
		HammerRange:        130, // радиус AoE удара
		HammerTriggerRange: 100, // триггер начала замаха
		HammerDamage:       150,
		HammerCooldown:     90, // 1.5 сек между ударами
		SupportCoverRange:  400,
	},
	EnemyMachineGunner: {
		Name:         "MachineGunner",
		MaxHP:        500,
		SP:           50,
		Size:         28,
		PreferredMin: 150, // ближе — в упор, максимальный урон
		PreferredMax: 450, // дальше — минимальный урон

		WeaponTurnRate:   0.05,
		WeaponArcDegrees: 20,

		ProjectileSpeed: 14,
		ProjectileSize:  5,
		ProjectileLife:  45,
		ProjectileCount: 1,

		MGMaxDamage:   10, // в упор
		MGMinDamage:   1,  // на пределе
		MGBurstCount:  6,  // 6 выстрелов в очереди
		MGBurstGap:    3,  // очень быстро
		MGCooldown:    90, // 1.5 сек между очередями
		MGPredictLead: 0.75,
		MGSpreadDeg:   4,

		SupportCoverRange: 400,
	},
	EnemySummoner: {
		Name:         "Summoner",
		MaxHP:        400,
		SP:           70,
		Size:         26,
		PreferredMin: 80,
		PreferredMax: 400,

		SummonCooldown: 1200, // 5 сек между призывами
		SummonCountMin: 3,
		SummonCountMax: 5,
		SummonRange:    60, // радиус, в котором появляются призраки
	},
	EnemyGhost: {
		Name:  "Ghost",
		MaxHP: 50,
		SP:    100,
		Size:  20,

		GhostContactDamage: 10,
		GhostContactRange:  40,
		GhostContactCool:   45, // 0.75 сек между ударами
	},
	EnemyHarpooner: {
		Name:         "Harpooner",
		MaxHP:        200,
		SP:           150,
		Size:         24,
		PreferredMin: 350,
		PreferredMax: 700,

		WeaponTurnRate:   0.04,
		WeaponArcDegrees: 15,

		HarpoonAimTime:    90, // 1.5 сек выцеливает
		HarpoonSpeed:      15, // быстро летит
		HarpoonSize:       12,
		HarpoonLife:       240,
		HarpoonHP:         30,  // чтобы отстрелить
		HarpoonPullSpeed:  1.2, // насколько быстро тянет игрока
		HarpoonCooldown:   240, // 4 сек после промаха/разрыва
		HarpoonAttachDist: 22,  // на каком расстоянии от игрока висит гарпун
	},
}

const (
	SniperStateIdle = iota
	SniperStateAiming
)

func (t EnemyType) Stats() EnemyStats {
	if s, ok := enemyStats[t]; ok {
		return s
	}
	return EnemyStats{Name: "Unknown"}
}

type Enemy struct {
	Type EnemyType

	X, Y float64
	HP   int

	FacingX, FacingY float64

	AttackTimer int
	SwingActive bool
	SwingTimer  int

	// Оглушение (от тарана игрока).
	StunTimer     int
	LastDashHitID int  // DashInstanceID, от которого уже получил оглушение
	Unparryable   bool // удар нельзя парировать (только блок урона)

	KnockbackVX    float64
	KnockbackVY    float64
	KnockbackTimer int

	LastSwingHitID int

	// Путь — срез waypoint'ов (мировых координат) для обхода стен.
	// Пока игрок виден напрямую, Path = nil, враг идёт по прямой.
	Path         []geometry.Point
	PathIndex    int
	PathCooldown int
	PathGoalCell [2]int

	// Детекция застревания.
	LastX, LastY float64
	StuckTicks   int

	ForcePathTimer int // если > 0 — игнорировать direct и идти только по A*

	// PrevPlayerDist — расстояние до игрока на прошлом тике.
	// Нужно стрелку, чтобы понимать, с какой скоростью игрок приближается.
	PrevPlayerDist float64
	// Кеш найденной позиции для стрельбы из-за укрытия.
	ShootPosX, ShootPosY float64
	ShootPosTimer        int
	// Трекинг попаданий: если HP упало — недавно ранен.
	LastHP int

	RecentlyHitTimer int

	// Разведчик: таймер для паттернов уклонения.
	AITimer int

	// Множитель скорости (используется разведчиком для рывков).
	SpeedBurst int

	// Разведчик: серия рывков при резком/точном наведении.
	DodgeTimer    int
	DodgeDirX     float64
	DodgeDirY     float64
	DodgeCooldown int
	DodgeQueue    int // сколько ещё рывков в текущей серии
	DodgePause    int // пауза между рывками серии

	// Стрелок: движение по орбите вокруг игрока.
	OrbitDir         int // -1 или +1, направление обхода
	OrbitChangeTimer int // тиков до смены направления

	// Медик.
	HealTargetIdx int // индекс союзника, которого лечит (-1 если нет)
	HealTick      int // счётчик лечения для визуальных эффектов
	HealPulse     int // для пульсации луча

	// Таран.
	DashActive   bool
	DashTimer    int
	DashDirX     float64
	DashDirY     float64
	DashCooldown int
	DashHit      bool

	// Melee lunge (общий для пехоты и меча тарана).
	SwingTargetX float64 // позиция игрока в момент начала замаха
	SwingTargetY float64
	LungeActive  bool // выполняется рывок после замаха
	LungeTimer   int
	LungeDur     int
	LungeDirX    float64
	LungeDirY    float64
	LungeSpeed   float64
	LungeHit     bool // уже нанёс урон в этом рывке

	// Таран: замах перед рывком (замедление).
	DashWindup      bool
	DashWindupTimer int

	// Снайпер.
	SniperState      int
	SniperTimer      int
	SniperAimX       float64
	SniperAimY       float64
	SniperShotX      float64
	SniperShotY      float64
	SniperShotLife   int
	SniperLeadFactor float64

	// Ракетчик.
	RocketFireTimer int

	// Буфер (на самом буфере).
	BufferTargets []int // индексы целей, обновляется каждый тик

	// Баффы (на получателе).
	BuffDamageMul float64
	BuffSpeedMul  float64
	BuffTimer     int

	// Беглец.
	AimAngle float64 // текущий угол взгляда, медленно вращается
	Latched  bool    // сцепился с игроком

	// Инженер.
	TurretSetupTimer int     // оставшееся время установки (тикает только в точке)
	MoveTimer        int     // оставшееся время движения к точке установки
	FleeTimer        int     // отступление после ранения
	HasPlacedTurret  bool    // турель уже поставлена
	SetupX, SetupY   float64 // координаты точки установки

	// Множитель скорости (используется огнемётчиком по HP).
	SpeedScale float64
	// Огнемётчик: угол оружия.
	WeaponAngle float64
	WeaponInit  bool

	// Протектор — цели, которых защищает.
	ProtectTargets []int
	// Флаг неуязвимости (на получателе).
	Invulnerable      bool
	InvulnerableTimer int

	// Танк.
	HammerState    int // 0 idle, 1 windup
	HammerTimer    int
	HammerCooldown int

	// Пулемётчик.
	MGBurstLeft    int
	MGBurstTimer   int
	MGCooldownLeft int

	// Призыватель.
	SummonCooldown int
	SummonerID     int // уникальный ID этого призывателя

	// Призрак.
	OwnerSummonerID int // к какому призывателю принадлежит

	// Гарпунист.
	HarpoonAiming     bool
	HarpoonAimTimer   int
	HarpoonCooldown   int
	HarpoonAimTargetX float64
	HarpoonAimTargetY float64
	// Уникальный идентификатор врага (не меняется при удалении соседей).
	UniqueID int

	GroupID int
	// Множители сложности (для timer wave). 0 = 1.0 (база).
	HPMultiplier    float64
	SpeedMultiplier float64
	// Defend: индекс спавна-цели (-1 = цель игрок).
	TargetSpawnIdx int
	// Defend: изначальное намерение — атаковать спавн.
	OriginalTargetSpawn int // -1 = на игрока, >= 0 = на спавн
}

func NewEnemy(t EnemyType, x, y float64) Enemy {
	hp := t.Stats().MaxHP
	return Enemy{
		Type:                t,
		X:                   x,
		Y:                   y,
		LastX:               x,
		LastY:               y,
		HP:                  hp,
		LastHP:              hp,
		HealTargetIdx:       -1,
		TargetSpawnIdx:      -1,
		OriginalTargetSpawn: -1,
	}
}

func (e *Enemy) Rect() geometry.Rect {
	s := e.Type.Stats().Size
	return geometry.Rect{X: e.X, Y: e.Y, W: s, H: s}
}

func (e *Enemy) Center() (float64, float64) {
	s := e.Type.Stats().Size
	return e.X + s/2, e.Y + s/2
}

func (e *Enemy) IsDead() bool {
	return e.HP <= 0
}

func (e *Enemy) TakeDamage(n int) {
	if n <= 0 {
		return
	}
	e.HP -= n
}

func (e *Enemy) Speed() float64 {
	base := EnemyBaseSpeed + float64(e.Type.Stats().SP)*EnemySpeedPerSP
	mul := e.BuffSpeedMul
	if mul <= 0 {
		mul = 1
	}
	scale := e.SpeedScale
	if scale <= 0 {
		scale = 1
	}
	return base * mul * scale
}

// HPRatio — для отрисовки полоски здоровья над врагом.
func (e *Enemy) HPRatio() float64 {
	m := e.Type.Stats().MaxHP
	if m <= 0 {
		return 0
	}
	r := float64(e.HP) / float64(m)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// DamageMul — множитель урона, накладываемый баффом.
func (e *Enemy) DamageMul() float64 {
	if e.BuffDamageMul <= 0 {
		return 1
	}
	return e.BuffDamageMul
}

// ApplyBuff вешает/обновляет бафф. Больший множитель перекрывает
// меньший, длительность обновляется.
func (e *Enemy) ApplyBuff(dmgMul, spdMul float64, duration int) {
	if dmgMul > e.BuffDamageMul {
		e.BuffDamageMul = dmgMul
	}
	if spdMul > e.BuffSpeedMul {
		e.BuffSpeedMul = spdMul
	}
	if duration > e.BuffTimer {
		e.BuffTimer = duration
	}
}

// TickBuffs тикает таймеры баффов. Возвращает false, если враг
// живой и получает бафф; true если бафф истёк — но это неважно,
// важно что поля обнуляются.
func (e *Enemy) TickBuffs() {
	if e.BuffTimer > 0 {
		e.BuffTimer--
	}
	if e.BuffTimer == 0 {
		e.BuffDamageMul = 0
		e.BuffSpeedMul = 0
	}
}
