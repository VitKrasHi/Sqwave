package world

// Snapshot — сетевое состояние мира: всё, что клиенту нужно для отрисовки
// и HUD. Статика (стены, nav-сетка, RNG) не входит — она строится локально
// по тем же seed/карте.
//
// Порядок Players совпадает с сетевыми ID игроков (0 — хост).
type Snapshot struct {
	Tick uint32

	Players          []Player
	Enemies          []Enemy
	Bullets          []Bullet
	Rockets          []Rocket
	Explosions       []Explosion
	EnemyProjectiles []EnemyProjectile
	EnemyRockets     []EnemyRocket
	Turrets          []Turret
	SpawnPoints      []SpawnPoint
	Coins            []Coin

	Wave  WaveState
	Stats Stats

	PendingTrack string
	CurrentTrack string
	// StartWaveRequested нужен клиенту, чтобы показать подсказку «R — волна».
	StartWaveRequested bool
}

// Capture снимает копию изменяемого состояния мира.
func (w *World) Capture(tick uint32) Snapshot {
	return Snapshot{
		Tick:               tick,
		Players:            append([]Player(nil), w.Players...),
		Enemies:            append([]Enemy(nil), w.Enemies...),
		Bullets:            append([]Bullet(nil), w.Bullets...),
		Rockets:            append([]Rocket(nil), w.Rockets...),
		Explosions:         append([]Explosion(nil), w.Explosions...),
		EnemyProjectiles:   append([]EnemyProjectile(nil), w.EnemyProjectiles...),
		EnemyRockets:       append([]EnemyRocket(nil), w.EnemyRockets...),
		Turrets:            append([]Turret(nil), w.Turrets...),
		SpawnPoints:        append([]SpawnPoint(nil), w.SpawnPoints...),
		Coins:              append([]Coin(nil), w.Coins...),
		Wave:               w.Wave,
		Stats:              w.Stats,
		PendingTrack:       w.PendingTrack,
		CurrentTrack:       w.CurrentTrack,
		StartWaveRequested: w.StartWaveRequested,
	}
}

// Apply переносит снапшот в мир. Срезы переиспользуются, чтобы не
// мусорить аллокациями каждый тик.
func (w *World) Apply(s Snapshot) {
	w.Players = append(w.Players[:0], s.Players...)
	w.Enemies = append(w.Enemies[:0], s.Enemies...)
	w.Bullets = append(w.Bullets[:0], s.Bullets...)
	w.Rockets = append(w.Rockets[:0], s.Rockets...)
	w.Explosions = append(w.Explosions[:0], s.Explosions...)
	w.EnemyProjectiles = append(w.EnemyProjectiles[:0], s.EnemyProjectiles...)
	w.EnemyRockets = append(w.EnemyRockets[:0], s.EnemyRockets...)
	w.Turrets = append(w.Turrets[:0], s.Turrets...)
	w.SpawnPoints = append(w.SpawnPoints[:0], s.SpawnPoints...)
	w.Coins = append(w.Coins[:0], s.Coins...)
	w.Wave = s.Wave
	w.Stats = s.Stats
	w.PendingTrack = s.PendingTrack
	w.CurrentTrack = s.CurrentTrack
	w.StartWaveRequested = s.StartWaveRequested
}
