package world

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponSMG
)

type WeaponStats struct {
	Name           string
	FireCooldown   int
	BulletSpeed    float64
	BulletSize     float64
	BulletLifetime int
	TrailLen       float64
}

var weapons = map[WeaponType]WeaponStats{
	WeaponPistol: {
		Name:           "Pistol",
		FireCooldown:   18, // ~0.3 сек
		BulletSpeed:    14,
		BulletSize:     5,
		BulletLifetime: 90,
		TrailLen:       28,
	},
	WeaponSMG: {
		Name:           "SMG",
		FireCooldown:   4, // ~0.07 сек — почти непрерывный поток
		BulletSpeed:    10,
		BulletSize:     4,
		BulletLifetime: 60,
		TrailLen:       16,
	},
}

func (w WeaponType) Stats() WeaponStats {
	return weapons[w]
}
