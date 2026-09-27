package world

import "Sqwave/internal/domain/geometry"

type SwingState struct {
	Active     bool
	Timer      int
	Duration   int
	StartAngle float64 // радианы, начало проворота
	ArcRadians float64 // полный угол проворота
	Range      float64
	Weapon     WeaponType
}

type Player struct {
	X, Y              float64
	DashTimer         int
	DashCooldownTimer int
	Weapon            WeaponType // активный слот
	Secondary         WeaponType // второй слот
	FireCooldownTimer int
	AimCharge         int
	Swing             SwingState
	HP                int
}

func (p *Player) Rect() geometry.Rect {
	return geometry.Rect{X: p.X, Y: p.Y, W: PlayerSize, H: PlayerSize}
}

func (p *Player) Center() (float64, float64) {
	return p.X + PlayerSize/2, p.Y + PlayerSize/2
}

func (p *Player) Dashing() bool {
	return p.DashTimer > 0
}

// IsAiming — снайперка заряжается прямо сейчас.
func (p *Player) IsAiming() bool {
	return p.Weapon == WeaponSniper && p.AimCharge > 0
}

// AimChargeRatio — 0 при начале прицеливания, 1 при полном заряде.
func (p *Player) AimChargeRatio() float64 {
	if !p.IsAiming() {
		return 0
	}
	ct := p.Weapon.Stats().ChargeTime
	if ct <= 0 {
		return 0
	}
	r := float64(p.AimCharge) / float64(ct)
	if r > 1 {
		return 1
	}
	return r
}

// CurrentSwingAngle — угол клинка в текущий момент проворота.
// Используется и логикой (для будущего хит-скана), и рендером.
func (p *Player) CurrentSwingAngle() float64 {
	if !p.Swing.Active || p.Swing.Duration <= 0 {
		return 0
	}
	progress := float64(p.Swing.Timer) / float64(p.Swing.Duration)
	return p.Swing.StartAngle + p.Swing.ArcRadians*progress
}

// SwingProgress — 0 в начале проворота, 1 в конце.
func (p *Player) SwingProgress() float64 {
	if !p.Swing.Active || p.Swing.Duration <= 0 {
		return 0
	}
	return float64(p.Swing.Timer) / float64(p.Swing.Duration)
}

func (p *Player) TakeDamage(n int) {
	if n <= 0 {
		return
	}
	p.HP -= n
	if p.HP < 0 {
		p.HP = 0
	}
}

func (p *Player) Heal(n int) {
	if n <= 0 {
		return
	}
	p.HP += n
	if p.HP > PlayerMaxHP {
		p.HP = PlayerMaxHP
	}
}

func (p *Player) IsDead() bool {
	return p.HP <= 0
}

func (p *Player) HPRatio() float64 {
	if PlayerMaxHP <= 0 {
		return 0
	}
	r := float64(p.HP) / float64(PlayerMaxHP)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

func (p *Player) HasWeapon() bool {
	return p.Weapon != WeaponNone
}

// SwapWeapon меняет активный слот со второстепенным.
// Не делает ничего, если второй слот пуст.
func (p *Player) SwapWeapon() {
	if p.Secondary == WeaponNone {
		return
	}
	p.Weapon, p.Secondary = p.Secondary, p.Weapon
	p.resetWeaponState()
}

// EquipPrimary ставит оружие в активный слот и сбрасывает
// состояние, привязанное к предыдущему оружию.
func (p *Player) EquipPrimary(w WeaponType) {
	p.Weapon = w
	p.resetWeaponState()
	p.FireCooldownTimer = 0
}

// EquipSecondary кладёт оружие во второстепенный слот.
// Активный слот не трогает.
func (p *Player) EquipSecondary(w WeaponType) {
	p.Secondary = w
}

func (p *Player) resetWeaponState() {
	p.AimCharge = 0
	p.Swing.Active = false
	if !p.Weapon.IsShield() {
		p.DashTimer = 0
	}
}
