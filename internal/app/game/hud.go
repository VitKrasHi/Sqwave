package game

import (
	"fmt"

	"Sqwave/internal/domain/world"
)

func drawWavePanel(r SceneRenderer, w *world.World) {
	ws := &w.Wave

	panelW := 240.0
	panelX := float64(world.ScreenWidth) - panelW - 12
	panelY := 12.0

	// Считаем высоту: заголовок + строки состава + прогресс.
	lines := 0
	if ws.Composition.Infantry > 0 {
		lines++
	}
	if ws.Composition.Shooter > 0 {
		lines++
	}
	if ws.Composition.Scout > 0 {
		lines++
	}

	headerH := 26.0
	rowH := 20.0
	footerH := 40.0
	panelH := headerH + float64(lines)*rowH + footerH

	r.DrawScreenRect(panelX-2, panelY-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(panelX, panelY, panelW, panelH, ColorMenuPanel)

	// Заголовок.
	title := "Wave —"
	if ws.Number > 0 {
		title = fmt.Sprintf("Wave %d", ws.Number)
	}
	r.DrawScreenText(title, panelX+12, panelY+6, ColorMenuSelected)

	// Состав.
	y := panelY + headerH
	if ws.Composition.Infantry > 0 {
		line := fmt.Sprintf("Infantry  x%d", ws.Composition.Infantry)
		r.DrawScreenText(line, panelX+12, y, ColorEnemy)
		y += rowH
	}
	if ws.Composition.Shooter > 0 {
		line := fmt.Sprintf("Shooter   x%d", ws.Composition.Shooter)
		r.DrawScreenText(line, panelX+12, y, ColorShooter)
		y += rowH
	}
	if ws.Composition.Scout > 0 {
		line := fmt.Sprintf("Scout     x%d", ws.Composition.Scout)
		r.DrawScreenText(line, panelX+12, y, ColorScout)
		y += rowH
	}

	// Прогресс.
	footerY := y + 4

	if ws.PauseTimer > 0 && !ws.Active {
		sec := ws.PauseTimer / 60
		r.DrawScreenText(fmt.Sprintf("Next wave in %ds", sec), panelX+12, footerY, ColorMenuText)
		return
	}

	if !ws.Active && ws.Number == 0 {
		r.DrawScreenText("Starting...", panelX+12, footerY, ColorMenuText)
		return
	}

	// Спавн прогресс.
	total := ws.Composition.Total()
	spawned := ws.SpawnedCount()
	alive := len(w.Enemies)
	killed := spawned - alive

	r.DrawScreenText(fmt.Sprintf("Spawned: %d/%d", spawned, total), panelX+12, footerY, ColorMenuText)
	r.DrawScreenText(fmt.Sprintf("Killed: %d  Alive: %d", killed, alive),
		panelX+12, footerY+16, ColorMenuText)
}
