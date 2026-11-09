package game

import (
	netpkg "Sqwave/internal/net"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
)

// netMode — режим игры по сети.
type netMode uint8

const (
	netOff netMode = iota
	netHost
	netGuest
)

// snapshotEvery — как часто хост рассылает снапшот (в тиках).
// 1 = 60 Гц: клиент рисует без интерполяции, поэтому частота = частота кадров.
const snapshotEvery = 1

// EnableHost переводит игру в режим хоста: мир симулируется здесь,
// клиенты присылают только ввод.
func (g *Game) EnableHost(h *netpkg.Host, players []netpkg.LobbyPlayer) {
	g.netMode = netHost
	g.netHost = h
	g.remoteInputs = make(map[uint8]netpkg.WireInput, len(players))
	g.netIDs = g.netIDs[:0]
	for _, p := range players {
		g.netIDs = append(g.netIDs, p.ID)
	}
	g.world.Local = 0
}

// EnableGuest переводит игру в режим клиента: симуляции нет, только ввод
// и снапшоты от хоста.
func (g *Game) EnableGuest(c *netpkg.Client, start netpkg.StartInfo) {
	g.netMode = netGuest
	g.netClient = c
	g.netIDs = g.netIDs[:0]
	g.world.Local = 0
	for i, p := range start.Players {
		g.netIDs = append(g.netIDs, p.ID)
		if p.ID == c.LocalID() {
			g.world.Local = i
		}
	}
}

// idxForNet — индекс игрока в мире по сетевому ID.
func (g *Game) idxForNet(id uint8) int {
	for i, nid := range g.netIDs {
		if nid == id {
			return i
		}
	}
	return -1
}

// pollHostEvents — события лобби в матче: ушедший игрок считается убитым.
func (g *Game) pollHostEvents() {
	if g.netHost == nil {
		return
	}
	for _, ev := range g.netHost.Poll() {
		if ev.Kind != netpkg.EvLobby {
			continue
		}
		present := make(map[uint8]bool, len(ev.Players))
		for _, p := range ev.Players {
			present[p.ID] = true
		}
		for i, nid := range g.netIDs {
			if i == 0 || present[nid] || i >= len(g.world.Players) {
				continue
			}
			p := &g.world.Players[i]
			if !p.IsDead() {
				p.Damage = p.MaxHP()
			}
		}
	}
}

// stepRemotePlayers прогоняет ввод клиентов по их игрокам.
//
// Ввод от клиента — это состояние (флаги удержания + прицел), которое
// запоминается между кадрами, иначе игрок дёргался бы на пропусках пакетов.
// Разовые флаги (выстрел на отпускании, свап, E) срабатывают ровно один раз.
func (g *Game) stepRemotePlayers() {
	if g.netHost == nil {
		return
	}
	w := g.world

	for _, pi := range g.netHost.Inputs() {
		prev, seen := g.remoteInputs[pi.PlayerID]
		g.remoteInputs[pi.PlayerID] = mergeWire(prev, pi.Input, seen)
	}

	for netID, win := range g.remoteInputs {
		idx := g.idxForNet(netID)
		if idx <= 0 || idx >= len(w.Players) {
			continue
		}
		p := &w.Players[idx]
		if p.IsDead() {
			continue
		}
		in := fromWire(win)
		systems.StepPlayer(w, idx, in)
		systems.StepModules(w, idx)
		systems.StepWeaponSelection(w, idx, in)
		systems.StepShooting(w, idx, in)

		// Разовые флаги уже отработали — гасим.
		win.FireReleased = false
		win.SwapWeapon = false
		win.Interact = false
		win.MenuConfirm = false
		win.MenuUp = false
		win.MenuDown = false
		win.MenuLeft = false
		win.MenuRight = false
		g.remoteInputs[netID] = win
	}
}

// mergeWire сводит новое состояние ввода с предыдущим: удержания и прицел
// берутся из нового, разовые нажатия накапливаются.
func mergeWire(prev, next netpkg.WireInput, seen bool) netpkg.WireInput {
	if !seen {
		return next
	}
	next.FireReleased = next.FireReleased || prev.FireReleased
	next.SwapWeapon = next.SwapWeapon || prev.SwapWeapon
	next.Interact = next.Interact || prev.Interact
	next.MenuConfirm = next.MenuConfirm || prev.MenuConfirm
	next.MenuUp = next.MenuUp || prev.MenuUp
	next.MenuDown = next.MenuDown || prev.MenuDown
	next.MenuLeft = next.MenuLeft || prev.MenuLeft
	next.MenuRight = next.MenuRight || prev.MenuRight
	return next
}

// applyRemoteCommands — команды клиентов (смена снаряжения, покупка HP).
func (g *Game) applyRemoteCommands() {
	if g.netHost == nil {
		return
	}
	w := g.world
	for _, cmd := range g.netHost.Commands() {
		idx := g.idxForNet(cmd.PlayerID)
		if idx <= 0 || idx >= len(w.Players) {
			continue
		}
		p := &w.Players[idx]
		if p.IsDead() || !w.InSpawnZone(p) {
			continue
		}
		switch cmd.Kind {
		case netpkg.CmdSetLoadout:
			primary := world.WeaponType(cmd.Primary)
			secondary := world.WeaponType(cmd.Secondary)
			if primary != world.WeaponNone && primary != secondary {
				p.EquipPrimary(primary)
			}
			if secondary != world.WeaponNone && secondary != primary {
				p.EquipSecondary(secondary)
			}
		case netpkg.CmdBuyHP:
			if p.ActiveModule == world.ModuleBloodExchange {
				continue
			}
			if p.Coins < world.FullHPPrice || p.Damage == 0 {
				continue
			}
			p.Coins -= world.FullHPPrice
			p.Damage = 0
		case netpkg.CmdRevive:
			dead := w.DeadCount()
			if dead == 0 {
				continue
			}
			cost := world.RevivePricePerPlayer * dead
			if p.Coins < cost {
				continue
			}
			p.Coins -= cost
			w.ReviveDead()
		}
	}
}

// broadcastSnapshot отдаёт клиентам состояние мира.
func (g *Game) broadcastSnapshot() {
	if g.netHost == nil {
		return
	}
	if g.ticks%snapshotEvery != 0 {
		return
	}
	snap := g.world.Capture(uint32(g.ticks))
	data, err := world.EncodeSnapshot(&snap)
	if err != nil {
		return
	}
	g.netHost.BroadcastSnapshot(uint32(g.ticks), data)
}

// pollGuestEvents применяет снапшоты и служебные события от хоста.
func (g *Game) pollGuestEvents() {
	if g.netClient == nil {
		return
	}
	// За кадр применяем только последний снапшот: промежуточные не нужны,
	// хост и так шлёт «последний победил».
	var lastSnap *netpkg.SnapshotMsg
	events := g.netClient.Poll()
	for i := range events {
		if events[i].Kind == netpkg.EvSnapshot && events[i].Snapshot != nil {
			lastSnap = events[i].Snapshot
		}
	}
	if lastSnap != nil {
		// Пока меню снаряжения открыто, не даём снапшоту откатить выбор:
		// команда ещё в пути, а без этого снаряжение мигало бы и «не выбиралось».
		var keepWeapon, keepSecondary world.WeaponType
		var keepCoins int
		keepLocal := g.menu.Open
		if p := g.world.LocalPlayer(); p != nil && keepLocal {
			keepWeapon, keepSecondary, keepCoins = p.Weapon, p.Secondary, p.Coins
		}

		if snap, err := world.DecodeSnapshot(lastSnap.Data); err == nil {
			g.world.Apply(snap)
			g.lastSnapshotTick = lastSnap.Tick
			if keepLocal {
				if p := g.world.LocalPlayer(); p != nil {
					p.Weapon, p.Secondary, p.Coins = keepWeapon, keepSecondary, keepCoins
				}
			}
		}
	}
	for _, ev := range events {
		switch ev.Kind {
		case netpkg.EvGameOver:
			if ev.Over == nil {
				continue
			}
			if ev.Over.Victory {
				if !g.victoryFired && g.onVictory != nil {
					g.victoryFired = true
					g.onVictory()
				}
			} else {
				if !g.deathFired && g.onPlayerDied != nil {
					g.deathFired = true
					g.onPlayerDied()
				}
			}
		case netpkg.EvDisconnected:
			g.connectionLost = true
			if g.onDisconnected != nil {
				g.onDisconnected()
			}
		}
	}
}

// sendGuestInput отправляет ввод локального игрока хосту.
func (g *Game) sendGuestInput(in input.PlayerInput) {
	if g.netClient == nil {
		return
	}
	g.netClient.SendInput(toWire(in))
}

// cameraPlayer — за кем следит камера: локальный игрок, а если он мёртв —
// первый живой напарник.
func (g *Game) cameraPlayer() *world.Player {
	w := g.world
	if p := w.LocalPlayer(); p != nil && !p.IsDead() {
		return p
	}
	for i := range w.Players {
		if !w.Players[i].IsDead() {
			return &w.Players[i]
		}
	}
	return w.LocalPlayer()
}

// toWire — доменный ввод в сетевой.
func toWire(in input.PlayerInput) netpkg.WireInput {
	return netpkg.WireInput{
		Up: in.Up, Down: in.Down, Left: in.Left, Right: in.Right,
		Dash: in.Dash, Fire: in.Fire, FireReleased: in.FireReleased,
		SwapWeapon: in.SwapWeapon, Interact: in.Interact,
		MenuUp: in.MenuUp, MenuDown: in.MenuDown,
		MenuLeft: in.MenuLeft, MenuRight: in.MenuRight,
		MenuConfirm: in.MenuConfirm, DashChargeHeld: in.DashChargeHeld,
		AimX: in.AimX, AimY: in.AimY,
	}
}

// fromWire — сетевой ввод в доменный.
func fromWire(in netpkg.WireInput) input.PlayerInput {
	return input.PlayerInput{
		Up: in.Up, Down: in.Down, Left: in.Left, Right: in.Right,
		Dash: in.Dash, Fire: in.Fire, FireReleased: in.FireReleased,
		SwapWeapon: in.SwapWeapon, Interact: in.Interact,
		MenuUp: in.MenuUp, MenuDown: in.MenuDown,
		MenuLeft: in.MenuLeft, MenuRight: in.MenuRight,
		MenuConfirm: in.MenuConfirm, DashChargeHeld: in.DashChargeHeld,
		AimX: in.AimX, AimY: in.AimY,
	}
}

// sendLoadout — клиент сообщает хосту новое снаряжение.
func (g *Game) sendLoadout() {
	if g.netClient == nil {
		return
	}
	p := g.world.LocalPlayer()
	if p == nil {
		return
	}
	g.netClient.SendCommand(netpkg.Command{
		Kind:      netpkg.CmdSetLoadout,
		Primary:   int32(p.Weapon),
		Secondary: int32(p.Secondary),
	})
}

// sendRevive — клиент просит хост воскресить павших напарников.
func (g *Game) sendRevive() {
	if g.netClient == nil {
		return
	}
	g.netClient.SendCommand(netpkg.Command{Kind: netpkg.CmdRevive})
}

// sendBuyHP — клиент просит хост купить лечение.
func (g *Game) sendBuyHP() {
	if g.netClient == nil {
		return
	}
	g.netClient.SendCommand(netpkg.Command{Kind: netpkg.CmdBuyHP})
}

// SetOnDisconnected — что делать, если связь с хостом пропала.
func (g *Game) SetOnDisconnected(f func()) {
	g.onDisconnected = f
}

// notifyGameOver сообщает клиентам результат матча.
func (g *Game) notifyGameOver(victory bool) {
	if g.netHost == nil {
		return
	}
	data, _ := world.EncodeSnapshot(func() *world.Snapshot {
		s := g.world.Capture(uint32(g.ticks))
		return &s
	}())
	g.netHost.SendGameOver(victory, data)
}

// guestFrame — кадр клиента: локальный UI, отправка ввода, применение
// снапшотов (уже сделано в Update) и камера.
func (g *Game) guestFrame(in input.PlayerInput) {
	p := g.world.LocalPlayer()
	if p == nil {
		return
	}

	// Меню снаряжения — локальный UI. Ввод в мир при открытом меню не идёт.
	if in.Interact && !g.menu.Open && !p.IsDead() && g.world.PlayerInSpawnZone() {
		g.menu.Open = true
		g.menu.Reset(p)
		return
	}
	if g.menu.Open && p.IsDead() {
		g.menu.Open = false
	}
	if g.menu.Open {
		beforeW, beforeS := p.Weapon, p.Secondary
		beforeCoins := p.Coins
		beforeDead := g.world.DeadCount()
		if handleMenuInput(g.world, &g.menu, in) {
			g.menu.Open = false
		}
		if p.Weapon != beforeW || p.Secondary != beforeS {
			g.sendLoadout()
		}
		if p.Coins != beforeCoins {
			if g.world.DeadCount() < beforeDead {
				g.sendRevive()
			} else {
				g.sendBuyHP()
			}
		}
		g.sendGuestInput(input.PlayerInput{})
		return
	}

	g.sendGuestInput(in)

	// Музыка: клиент повторяет трек хоста.
	if g.music != nil && g.world.CurrentTrack != "" && g.world.CurrentTrack != g.lastTrack {
		if err := g.music.PlayLoopForce(g.world.CurrentTrack); err == nil {
			g.lastTrack = g.world.CurrentTrack
		}
	}

	cp := g.cameraPlayer()
	if cp != nil {
		cx, cy := cp.Center()
		g.camera.Follow(cx, cy, in.AimX, in.AimY, p.AimChargeRatio())
	}

	// Клиент не симулирует смерть — её присылает хост событием GameOver,
	// но на всякий случай проверяем снапшот.
	if !g.deathFired && g.world.AllPlayersDead() {
		g.deathFired = true
		if g.onPlayerDied != nil {
			g.onPlayerDied()
		}
	}
	if !g.victoryFired && g.world.Wave.IsVictory() && !g.world.AllPlayersDead() {
		g.victoryFired = true
		if g.onVictory != nil {
			g.onVictory()
		}
	}
}
