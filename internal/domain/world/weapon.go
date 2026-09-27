package world

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponSMG
	WeaponShotgun
	WeaponSniper
	WeaponRocket
	WeaponSword
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

	// Melee > 0 — оружие ближнего боя. Вместо выстрела запускает
	// проворот клинка в секторе MeleeArc градусов на дистанцию MeleeRange.
	MeleeRange    float64
	MeleeArc      float64 // полный угол проворота в градусах
	SwingDuration int     // длительность проворота в тиках
	BladeWidth    float64 // толщина клинка для визуала
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
}

func (w WeaponType) Stats() WeaponStats {
	return weapons[w]
}

// IsMelee — вспомогательный предикат, чтобы не проверять MeleeRange > 0 в каждом месте.
func (w WeaponType) IsMelee() bool {
	return w.Stats().MeleeRange > 0
}
