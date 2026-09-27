package world

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponSMG
	WeaponShotgun
	WeaponSniper
	WeaponRocket
	WeaponSword
	WeaponShield
)

type WeaponStats struct {
	Name         string
	FireCooldown int
	Range        float64
	Spread       float64
	Pellets      int
	VisualLife   int
	ChargeTime   int

	ProjectileSpeed float64
	ProjectileSize  float64

	ExplosionRadius float64
	ExplosionLife   int

	MeleeRange    float64
	MeleeArc      float64
	SwingDuration int
	BladeWidth    float64
	IsShield      bool

	// IsThrust — melee-атака «толчок»: не крутит дугу,
	// а выдвигает оружие вперёд по направлению курсора и возвращает назад.
	IsThrust bool
}

var weapons = map[WeaponType]WeaponStats{
	WeaponPistol: {
		Name: "Pistol", FireCooldown: 18, Range: 700,
		Spread: 2, Pellets: 1, VisualLife: 6,
	},
	WeaponSMG: {
		Name: "SMG", FireCooldown: 4, Range: 550,
		Spread: 6, Pellets: 1, VisualLife: 4,
	},
	WeaponShotgun: {
		Name: "Shotgun", FireCooldown: 45, Range: 260,
		Spread: 25, Pellets: 8, VisualLife: 5,
	},
	WeaponSniper: {
		Name: "Sniper", FireCooldown: 60, Range: 3000,
		Spread: 0, Pellets: 1, VisualLife: 12, ChargeTime: 45,
	},
	WeaponRocket: {
		Name:            "Rocket",
		FireCooldown:    60,
		Range:           900,
		Spread:          1,
		Pellets:         1,
		ProjectileSpeed: 12,
		ProjectileSize:  8,
		ExplosionRadius: 130,
		ExplosionLife:   14,
	},
	WeaponSword: {
		Name:          "Sword",
		FireCooldown:  24,
		MeleeRange:    82,
		MeleeArc:      150,
		SwingDuration: 12,
		BladeWidth:    6,
	},
	WeaponShield: {
		Name:          "Shield",
		FireCooldown:  30,
		MeleeRange:    58, // докуда выдвигается щит
		MeleeArc:      0,  // толчок — дуги нет
		SwingDuration: 10, // чуть медленнее для ощутимости
		BladeWidth:    18, // сам щит — квадрат 18×18
		IsShield:      true,
		IsThrust:      true,
	},
}

const WeaponNone WeaponType = -1

func (w WeaponType) Stats() WeaponStats {
	if s, ok := weapons[w]; ok {
		return s
	}
	return WeaponStats{Name: "Unarmed"}
}

func (w WeaponType) IsMelee() bool {
	return w.Stats().MeleeRange > 0
}

func (w WeaponType) IsShield() bool {
	return w.Stats().IsShield
}

func (w WeaponType) IsNone() bool {
	return w == WeaponNone
}
