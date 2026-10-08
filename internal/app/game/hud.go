package game

import (
	"fmt"
	"image/color"

	"Sqwave/internal/domain/world"
)

func drawWavePanel(r SceneRenderer, w *world.World) {
	ws := &w.Wave

	// Считаем строки состава — только те, что > 0.
	comp := ws.Composition
	type line struct {
		label string
		count int
		c     color.NRGBA
	}
	var lines []line
	if comp.Infantry > 0 {
		lines = append(lines, line{"Infantry", comp.Infantry, ColorEnemy})
	}
	if comp.Shooter > 0 {
		lines = append(lines, line{"Shooter", comp.Shooter, ColorShooter})
	}
	if comp.Scout > 0 {
		lines = append(lines, line{"Scout", comp.Scout, ColorScout})
	}
	if comp.Medic > 0 {
		lines = append(lines, line{"Medic", comp.Medic, ColorMedic})
	}
	if comp.Rammer > 0 {
		lines = append(lines, line{"Rammer", comp.Rammer, ColorRammer})
	}
	if comp.Sniper > 0 {
		lines = append(lines, line{"Sniper", comp.Sniper, ColorSniper})
	}

	const (
		panelW  = 260.0
		headerH = 26.0
		rowH    = 20.0
		slotH   = 20.0
		footerH = 60.0
		topPad  = 6.0
	)

	panelX := float64(world.ScreenWidth) - panelW - 12
	panelY := 12.0

	// Высота панели = заголовок + строки состава + разделитель + прогресс + отступ.
	progressLines := 3 // Spawned, Killed, Alive
	if ws.GroupRemaining > 0 {
		progressLines++
	}
	panelH := headerH +
		float64(len(lines))*rowH +
		float64(progressLines)*slotH +
		footerH

	r.DrawScreenRect(panelX-2, panelY-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(panelX, panelY, panelW, panelH, ColorMenuPanel)

	// Заголовок.
	title := "Wave —"
	if ws.Number > 0 {
		title = fmt.Sprintf("Wave %d", ws.Number)
	}
	r.DrawScreenText(title, panelX+12, panelY+topPad, ColorMenuSelected)

	y := panelY + headerH

	// Состояние до первой волны.
	if !ws.Active && ws.Number == 0 {
		r.DrawScreenText("Starting...", panelX+12, y, ColorMenuText)
		return
	}

	// Пауза между волнами.
	if !ws.Active && ws.PauseTimer > 0 {
		sec := ws.PauseTimer / 60
		r.DrawScreenText(fmt.Sprintf("Next wave in %ds", sec), panelX+12, y, ColorMenuText)
		return
	}

	// Состав волны.
	for _, l := range lines {
		label := fmt.Sprintf("%-9s x%d", l.label, l.count)
		r.DrawScreenText(label, panelX+12, y, l.c)
		y += rowH
	}

	// Разделитель.
	y += 4

	// Прогресс.
	r.DrawScreenText(fmt.Sprintf("Spawned: %d / %d", ws.SpawnedCount, comp.Total()),
		panelX+12, y, ColorMenuText)
	y += slotH
	r.DrawScreenText(fmt.Sprintf("Killed: %d", ws.KilledCount),
		panelX+12, y, ColorMenuText)
	y += slotH

	// Alive — считаем врагов с GroupID > 0 (все живущие в этой волне).
	alive := 0
	for i := range w.Enemies {
		if w.Enemies[i].GroupID > 0 {
			alive++
		}
	}
	r.DrawScreenText(fmt.Sprintf("Alive: %d", alive),
		panelX+12, y, ColorMenuText)
	y += slotH

	// Статус группы.
	if ws.GroupRemaining > 0 {
		r.DrawScreenText(fmt.Sprintf("Group: %d left", ws.GroupRemaining),
			panelX+12, y, ColorMenuDim)
	}
}

// groupAliveProxy — то же, что systems.groupAlive, но не импортирует systems.
func groupAliveProxy(w *world.World, gid int) bool {
	if gid == 0 {
		return false
	}
	for i := range w.Enemies {
		if w.Enemies[i].GroupID == gid {
			return true
		}
	}
	return false
}
