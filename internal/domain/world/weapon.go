package world

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponSMG
	WeaponShotgun
	WeaponSniper
)

type WeaponStats struct {
	Name         string
	FireCooldown int
	Range        float64
	Spread       float64
	Pellets      int
	VisualLife   int
	ChargeTime   int // для снайперки: тиков до полного заряда
}

var weapons = map[WeaponType]WeaponStats{
	WeaponPistol: {
		Name:         "Pistol",
		FireCooldown: 18,
		Range:        700,
		Spread:       2,
		Pellets:      1,
		VisualLife:   6,
	},
	WeaponSMG: {
		Name:         "SMG",
		FireCooldown: 4,
		Range:        550,
		Spread:       6,
		Pellets:      1,
		VisualLife:   4,
	},
	WeaponShotgun: {
		Name:         "Shotgun",
		FireCooldown: 45,
		Range:        260,
		Spread:       25,
		Pellets:      8,
		VisualLife:   5,
	},
	WeaponSniper: {
		Name:         "Sniper",
		FireCooldown: 60,
		Range:        3000,
		Spread:       0,
		Pellets:      1,
		VisualLife:   12,
		ChargeTime:   45, // 0.75 сек до полного
	},
}

func (w WeaponType) Stats() WeaponStats {
	return weapons[w]
}
