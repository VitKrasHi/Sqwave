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
	IsThrust      bool

	// HPBonus — сколько HP добавляет это оружие в общий пул игрока.
	HPBonus    int
	SpeedBonus int
}

var weapons = map[WeaponType]WeaponStats{
	WeaponPistol: {
		Name: "Pistol", FireCooldown: 18, Range: 700,
		Spread: 2, Pellets: 1, VisualLife: 6,
		HPBonus: 40, SpeedBonus: 50,
	},
	WeaponSMG: {
		Name: "SMG", FireCooldown: 4, Range: 550,
		Spread: 6, Pellets: 1, VisualLife: 4,
		HPBonus: 60, SpeedBonus: 30,
	},
	WeaponShotgun: {
		Name: "Shotgun", FireCooldown: 45, Range: 260,
		Spread: 25, Pellets: 8, VisualLife: 5,
		HPBonus: 80, SpeedBonus: 20,
	},
	WeaponSniper: {
		Name: "Sniper", FireCooldown: 60, Range: 3000,
		Spread: 0, Pellets: 1, VisualLife: 12, ChargeTime: 45,
		HPBonus: 20, SpeedBonus: 40,
	},
	WeaponRocket: {
		Name:         "Rocket",
		FireCooldown: 60, Range: 900, Spread: 1, Pellets: 1,
		ProjectileSpeed: 12, ProjectileSize: 8,
		ExplosionRadius: 130, ExplosionLife: 14,
		HPBonus: 120, SpeedBonus: 20,
	},
	WeaponSword: {
		Name: "Sword", FireCooldown: 24,
		MeleeRange: 82, MeleeArc: 150, SwingDuration: 12, BladeWidth: 5,
		HPBonus: 100, SpeedBonus: 50,
	},
	WeaponShield: {
		Name: "Shield", FireCooldown: 30,
		MeleeRange: 58, MeleeArc: 0, SwingDuration: 10, BladeWidth: 18,
		IsShield: true, IsThrust: true,
		HPBonus: 50, SpeedBonus: 50,
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
