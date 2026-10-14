package game

import (
	"fmt"
	"image/color"

	"Sqwave/internal/domain/world"
)

func drawWavePanel(r SceneRenderer, w *world.World) {
	ws := &w.Wave

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
	if comp.Rocketeer > 0 {
		lines = append(lines, line{"Rocketeer", comp.Rocketeer, ColorRocketeer})
	}
	if comp.Buffer > 0 {
		lines = append(lines, line{"Buffer", comp.Buffer, ColorBuffer})
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

	headerLines := 1
	if ws.Number > 0 {
		headerLines++
	}
	progressLines := 3
	if ws.GroupRemaining > 0 || ws.Kind == world.WaveKindTimer {
		progressLines++
	}
	panelH := headerH +
		float64(headerLines-1)*rowH +
		float64(len(lines))*rowH +
		float64(progressLines)*slotH +
		footerH

	r.DrawScreenRect(panelX-2, panelY-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(panelX, panelY, panelW, panelH, ColorMenuPanel)

	title := "Wave -"
	if ws.Number > 0 {
		if ws.WavesLimit > 0 {
			title = fmt.Sprintf("Wave %d / %d", ws.Number, ws.WavesLimit)
		} else {
			title = fmt.Sprintf("Wave %d / infinity", ws.Number)
		}
	}
	r.DrawScreenText(title, panelX+12, panelY+topPad, ColorMenuSelected)

	y := panelY + headerH

	// Строка типа волны / таймера.
	if ws.Number > 0 {
		if ws.Kind == world.WaveKindTimer {
			sec := ws.TimerRemaining / 60
			min := sec / 60
			sec = sec % 60
			text := fmt.Sprintf("TIMER %02d:%02d", min, sec)
			r.DrawScreenText(text, panelX+12, y, ColorMenuSelected)
		} else {
			r.DrawScreenText(fmt.Sprintf("Level %d / %d", ws.WaveLevel, ws.MaxWaveLevel),
				panelX+12, y, ColorMenuText)
		}
		y += rowH
	}

	if !ws.Active && ws.Number == 0 {
		r.DrawScreenText("Starting...", panelX+12, y, ColorMenuText)
		return
	}

	if ws.IsVictory() {
		r.DrawScreenText("VICTORY!", panelX+12, y, ColorMenuSelected)
		y += rowH
	}

	if !ws.Active && ws.PauseTimer > 0 && !ws.IsVictory() {
		sec := ws.PauseTimer / 60
		r.DrawScreenText(fmt.Sprintf("Next wave in %ds", sec), panelX+12, y, ColorMenuText)
		y += rowH
	}

	for _, l := range lines {
		label := fmt.Sprintf("%-9s x%d", l.label, l.count)
		r.DrawScreenText(label, panelX+12, y, l.c)
		y += rowH
	}

	y += 4

	// Прогресс.
	if ws.Kind == world.WaveKindTimer {
		progress := 0.0
		if ws.TimerTotal > 0 {
			progress = 1 - float64(ws.TimerRemaining)/float64(ws.TimerTotal)
		}
		r.DrawScreenText(fmt.Sprintf("Difficulty: %d%%", int(progress*100)),
			panelX+12, y, ColorMenuText)
		y += slotH
	} else {
		r.DrawScreenText(fmt.Sprintf("Spawned: %d / %d", ws.SpawnedCount, ws.WaveSize),
			panelX+12, y, ColorMenuText)
		y += slotH
	}

	r.DrawScreenText(fmt.Sprintf("Killed: %d", ws.KilledCount),
		panelX+12, y, ColorMenuText)
	y += slotH

	alive := 0
	for i := range w.Enemies {
		if w.Enemies[i].GroupID > 0 {
			alive++
		}
	}
	r.DrawScreenText(fmt.Sprintf("Alive: %d", alive),
		panelX+12, y, ColorMenuText)
	y += slotH

	if ws.GroupRemaining > 0 {
		r.DrawScreenText(fmt.Sprintf("Group: %d left", ws.GroupRemaining),
			panelX+12, y, ColorMenuDim)
	} else if ws.Kind == world.WaveKindCommon && ws.GroupsRemaining > 0 {
		r.DrawScreenText(fmt.Sprintf("Groups left: %d", ws.GroupsRemaining),
			panelX+12, y, ColorMenuDim)
	}
}
