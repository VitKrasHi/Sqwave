package world

type ModuleType int

const (
	ModuleNone ModuleType = iota
	ModuleBloodExchange
	ModuleVulnerability
	ModuleRateOfFire
)

type ModuleStats struct {
	Name        string
	Description string
	Price       int
}

var modules = map[ModuleType]ModuleStats{
	ModuleBloodExchange: {
		Name:        "Blood Exchange",
		Description: "Damage = Heal (1/4). No base healing.",
		Price:       0,
	},
	ModuleVulnerability: {
		Name:        "Vulnerability",
		Description: "First hit: mark (5s). Second hit: x4 damage.",
		Price:       0,
	},
	ModuleRateOfFire: {
		Name:        "Rate of Fire",
		Description: "Hit: +10% speed. Miss: reset. Base -50%.",
		Price:       0,
	},
}

func (m ModuleType) Stats() ModuleStats {
	if s, ok := modules[m]; ok {
		return s
	}
	return ModuleStats{Name: "Unknown"}
}

func (m ModuleType) IsNone() bool {
	return m == ModuleNone
}
