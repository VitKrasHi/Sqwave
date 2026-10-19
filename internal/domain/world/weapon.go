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

const WeaponNone WeaponType = -1

type DamageKind int

const (
	DamageFixed           DamageKind = iota // постоянный урон
	DamageDistanceFalloff                   // спад по дистанции
	DamageChargeScaled                      // спад по заряду
	DamageRocket                            // считается отдельно в StepRockets
	DamageNone                              // щит — без урона, только толчок
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

	HPBonus    int
	SpeedBonus int

	// Урон.
	DamageKind DamageKind

	Damage int // для DamageFixed

	DamageNear int     // для DamageDistanceFalloff — урон в упор
	DamageFar  int     // урон на FalloffAt
	FalloffAt  float64 // дистанция, на которой урон падает до DamageFar

	DamageMinCharge int // для DamageChargeScaled
	DamageMaxCharge int

	DirectHitDamage int // для DamageRocket — попадание
	ExplosionDamage int // взрыв по площади

	KnockbackForce float64 // для щита — сила отбрасывания
}

var weapons = map[WeaponType]WeaponStats{
	WeaponPistol: {
		Name: "Pistol", FireCooldown: 18, Range: 700,
		Spread: 2, Pellets: 1, VisualLife: 6,
		HPBonus: 40, SpeedBonus: 50,
		DamageKind: DamageFixed, Damage: 30,
	},
	WeaponSMG: {
		Name: "SMG", FireCooldown: 4, Range: 550,
		Spread: 6, Pellets: 1, VisualLife: 4,
		HPBonus: 60, SpeedBonus: 30,
		DamageKind: DamageDistanceFalloff,
		DamageNear: 20, DamageFar: 5, FalloffAt: 550,
	},
	WeaponShotgun: {
		Name: "Shotgun", FireCooldown: 45, Range: 260,
		Spread: 25, Pellets: 8, VisualLife: 5,
		HPBonus: 80, SpeedBonus: 20,
		DamageKind: DamageDistanceFalloff,
		DamageNear: 80, DamageFar: 5, FalloffAt: 260,
	},
	WeaponSniper: {
		Name: "Sniper", FireCooldown: 60, Range: 3000,
		Spread: 0, Pellets: 1, VisualLife: 12, ChargeTime: 45,
		HPBonus: 20, SpeedBonus: 40,
		DamageKind:      DamageChargeScaled,
		DamageMinCharge: 50, DamageMaxCharge: 200,
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
		HPBonus:         120, SpeedBonus: 20,
		DamageKind:      DamageRocket,
		DirectHitDamage: 120, ExplosionDamage: 40,
	},
	WeaponSword: {
		Name: "Sword", FireCooldown: 24,
		MeleeRange: 82, MeleeArc: 150, SwingDuration: 12, BladeWidth: 5,
		HPBonus: 100, SpeedBonus: 50,
		DamageKind: DamageFixed, Damage: 50,
	},
	WeaponShield: {
		Name: "Shield", FireCooldown: 30,
		MeleeRange: 58, MeleeArc: 0, SwingDuration: 10, BladeWidth: 18,
		IsShield: true, IsThrust: true,
		HPBonus: 50, SpeedBonus: 50,
		DamageKind:     DamageNone,
		KnockbackForce: 15,
	},
}

func (w WeaponType) Stats() WeaponStats {
	if s, ok := weapons[w]; ok {
		return s
	}
	return WeaponStats{Name: "Unarmed"}
}

func (w WeaponType) IsMelee() bool  { return w.Stats().MeleeRange > 0 }
func (w WeaponType) IsShield() bool { return w.Stats().IsShield }
func (w WeaponType) IsNone() bool   { return w == WeaponNone }

// DamageAt возвращает урон при попадании на дистанции distance
// с учётом chargeRatio (для снайперки — 0..1).
// DamageAt возвращает урон при попадании на дистанции distance
// с учётом chargeRatio (для снайперки — 0..1).
func (w WeaponType) DamageAt(distance, chargeRatio float64) int {
	s := w.Stats()
	switch s.DamageKind {
	case DamageFixed:
		return s.Damage
	case DamageDistanceFalloff:
		if s.FalloffAt <= 0 {
			return s.DamageNear
		}
		t := distance / s.FalloffAt
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		return s.DamageNear + int(float64(s.DamageFar-s.DamageNear)*t)
	case DamageChargeScaled:
		if chargeRatio < 0 {
			chargeRatio = 0
		}
		if chargeRatio > 1 {
			chargeRatio = 1
		}
		return s.DamageMinCharge + int(float64(s.DamageMaxCharge-s.DamageMinCharge)*chargeRatio)
	}
	return 0
}
