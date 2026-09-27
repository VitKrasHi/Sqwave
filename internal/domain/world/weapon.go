package world

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponSMG
	WeaponShotgun
)

type WeaponStats struct {
	Name         string
	FireCooldown int
	Range        float64
	Spread       float64
	Pellets      int
	VisualLife   int
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
}

func (w WeaponType) Stats() WeaponStats {
	return weapons[w]
}
