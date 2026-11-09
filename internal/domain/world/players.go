package world

import "math"

// LocalPlayer — игрок локального пользователя. В кооперативе у каждого
// процесса свой Local, поэтому HUD и камера смотрят каждый на своего.
func (w *World) LocalPlayer() *Player {
	if len(w.Players) == 0 {
		return nil
	}
	if w.Local < 0 || w.Local >= len(w.Players) {
		w.Local = 0
	}
	return &w.Players[w.Local]
}

// NearestPlayer — ближайший живой игрок к точке. Если живых нет,
// возвращает ближайшего вообще (чтобы враги не паниковали).
func (w *World) NearestPlayer(x, y float64) *Player {
	var alive, any *Player
	aliveD, anyD := 0.0, 0.0
	for i := range w.Players {
		p := &w.Players[i]
		cx, cy := p.Center()
		d := (cx-x)*(cx-x) + (cy-y)*(cy-y)
		if any == nil || d < anyD {
			any, anyD = p, d
		}
		if p.IsDead() {
			continue
		}
		if alive == nil || d < aliveD {
			alive, aliveD = p, d
		}
	}
	if alive != nil {
		return alive
	}
	return any
}

// AliveCount — сколько игроков ещё живы.
func (w *World) AliveCount() int {
	n := 0
	for i := range w.Players {
		if !w.Players[i].IsDead() {
			n++
		}
	}
	return n
}

// DeadCount — сколько игроков мертвы.
func (w *World) DeadCount() int {
	n := 0
	for i := range w.Players {
		if w.Players[i].IsDead() {
			n++
		}
	}
	return n
}

// ReviveDead поднимает всех павших у зоны спавна и возвращает их число.
func (w *World) ReviveDead() int {
	if w.DeadCount() == 0 {
		return 0
	}
	spots := w.SpawnPointsForPlayers(len(w.Players))
	n := 0
	for i := range w.Players {
		p := &w.Players[i]
		if !p.IsDead() {
			continue
		}
		p.Revive()
		if i < len(spots) {
			p.X, p.Y = spots[i][0], spots[i][1]
			p.PrevX, p.PrevY = p.X, p.Y
			p.VelX, p.VelY = 0, 0
		}
		n++
	}
	if n > 0 {
		w.Stats.Revives += n
	}
	return n
}

// AllPlayersArmed — все игроки выбрали снаряжение.
// Кооп-волна не начинается, пока кто-то ещё в меню спавна.
func (w *World) AllPlayersArmed() bool {
	for i := range w.Players {
		if !w.Players[i].HasWeapon() {
			return false
		}
	}
	return true
}

// AllPlayersDead — все мертвы (кооп: конец игры).
func (w *World) AllPlayersDead() bool {
	return w.AliveCount() == 0
}

// PlayerIndex — индекс игрока в срезе, -1 если не найден.
func (w *World) PlayerIndex(p *Player) int {
	for i := range w.Players {
		if &w.Players[i] == p {
			return i
		}
	}
	return -1
}

// SpawnPointsForPlayers — позиции для n игроков вокруг центра зоны спавна.
func (w *World) SpawnPointsForPlayers(n int) [][2]float64 {
	if n <= 0 {
		n = 1
	}
	cx := w.SpawnZone.X + w.SpawnZone.W/2
	cy := w.SpawnZone.Y + w.SpawnZone.H/2
	if n == 1 {
		return [][2]float64{{cx - PlayerSize/2, cy - PlayerSize/2}}
	}
	out := make([][2]float64, n)
	radius := 64.0
	for i := 0; i < n; i++ {
		ang := 2 * math.Pi * float64(i) / float64(n)
		out[i] = [2]float64{
			cx + math.Cos(ang)*radius - PlayerSize/2,
			cy + math.Sin(ang)*radius - PlayerSize/2,
		}
	}
	return out
}

// InitPlayers — расставляет n игроков вокруг зоны спавна.
// Используется при старте кооперативной игры; одиночная игра живёт с одним.
func (w *World) InitPlayers(n int) {
	if n <= 0 {
		n = 1
	}
	spots := w.SpawnPointsForPlayers(n)
	players := make([]Player, n)
	for i := 0; i < n; i++ {
		x, y := spots[i][0], spots[i][1]
		players[i] = Player{
			X:         x,
			Y:         y,
			PrevX:     x,
			PrevY:     y,
			Weapon:    WeaponNone,
			Secondary: WeaponNone,
			HeldBy:    -1,
		}
	}
	w.Players = players
	w.Local = 0
}

// PlayersCenter — центр масс живых игроков. Если живых нет — центр зоны спавна.
func (w *World) PlayersCenter() (float64, float64) {
	var sx, sy float64
	n := 0
	for i := range w.Players {
		p := &w.Players[i]
		if p.IsDead() {
			continue
		}
		cx, cy := p.Center()
		sx += cx
		sy += cy
		n++
	}
	if n == 0 {
		return w.SpawnZone.X + w.SpawnZone.W/2, w.SpawnZone.Y + w.SpawnZone.H/2
	}
	return sx / float64(n), sy / float64(n)
}
