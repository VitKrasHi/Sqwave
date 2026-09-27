package world

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponSMG
	WeaponShotgun
	WeaponSniper
	WeaponRocket
)

type WeaponStats struct {
	Name         string
	FireCooldown int
	Range        float64
	Spread       float64
	Pellets      int
	VisualLife   int
	ChargeTime   int

	// Projectile — если > 0, оружие стреляет движущимся снарядом,
	// а не hitscan-лучом.
	ProjectileSpeed float64
	ProjectileSize  float64

	// Explosion — если > 0, снаряд при столкновении создаёт зону.
	ExplosionRadius float64
	ExplosionLife   int
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
}

func (w WeaponType) Stats() WeaponStats {
	return weapons[w]
}
