package game

import (
	"fmt"
	"image/color"

	"Sqwave/internal/domain/world"
)

// drawWavePanel — панель справа сверху: волна, уровень, состав,
// прогресс, статус игрока.
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
	if comp.Fleer > 0 {
		lines = append(lines, line{"Fleer", comp.Fleer, ColorFleer})
	}
	if comp.Engineer > 0 {
		lines = append(lines, line{"Engineer", comp.Engineer, ColorEngineer})
	}
	if comp.Flamer > 0 {
		lines = append(lines, line{"Flamer", comp.Flamer, ColorFlamer})
	}
	if comp.Protector > 0 {
		lines = append(lines, line{"Protector", comp.Protector, ColorProtector})
	}
	if comp.Tank > 0 {
		lines = append(lines, line{"Tank", comp.Tank, ColorTank})
	}
	if comp.MachineGunner > 0 {
		lines = append(lines, line{"MGunner", comp.MachineGunner, ColorMachineGunner})
	}
	if comp.Summoner > 0 {
		lines = append(lines, line{"Summoner", comp.Summoner, ColorSummoner})
	}
	if comp.Harpooner > 0 {
		lines = append(lines, line{"Harpooner", comp.Harpooner, ColorHarpooner})
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

	// Шапка: заголовок + level + (timer).
	headerLines := 1
	if ws.Number > 0 {
		headerLines++
	}
	if ws.Active && ws.Kind == world.WaveKindTimer {
		headerLines++
	}

	// Прогресс: Spawned/Killed/Alive + (Spawns для defend) +
	// (группа или таймер следующей группы).
	progressLines := 3
	if ws.Kind == world.WaveKindDefend {
		progressLines++
	}
	if ws.GroupRemaining > 0 {
		progressLines++
	} else if ws.Kind == world.WaveKindCommon && ws.GroupsRemaining > 0 {
		progressLines++
	} else if ws.Kind == world.WaveKindTimer && ws.TimerGroupCooldown > 0 {
		progressLines++
	}

	panelH := headerH +
		float64(headerLines-1)*rowH +
		float64(len(lines))*rowH +
		float64(progressLines)*slotH +
		footerH

	r.DrawScreenRect(panelX-2, panelY-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(panelX, panelY, panelW, panelH, ColorMenuPanel)

	// Заголовок.
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

	// Строка уровня + типа волны.
	if ws.Number > 0 {
		roman := []string{"", "I", "II", "III", "IV", "V"}
		lvl := ws.WaveLevel
		if lvl < 1 {
			lvl = 1
		}
		if lvl > 5 {
			lvl = 5
		}

		kindStr := "COMMON"
		kindCol := ColorMenuText
		switch ws.Kind {
		case world.WaveKindTimer:
			kindStr = "TIMER"
			kindCol = ColorMenuSelected
		case world.WaveKindDefend:
			kindStr = "DEFEND"
			kindCol = color.NRGBA{R: 100, G: 220, B: 140, A: 255}
		}

		r.DrawScreenText(fmt.Sprintf("%s  Level %s", kindStr, roman[lvl]),
			panelX+12, y, kindCol)
		y += rowH
	}

	// Таймер обратного отсчёта для timer-волны.
	if ws.Active && ws.Kind == world.WaveKindTimer {
		sec := ws.TimerRemaining / 60
		min := sec / 60
		sec = sec % 60
		timerStr := fmt.Sprintf("TIMER %02d:%02d", min, sec)
		r.DrawScreenText(timerStr, panelX+12, y, ColorMenuSelected)
		y += rowH
	}

	if !ws.Active && ws.Number == 0 {
		r.DrawScreenText("Starting...", panelX+12, y, ColorMenuText)
		return
	}

	// Победа.
	if ws.IsVictory() {
		r.DrawScreenText("VICTORY!", panelX+12, y, ColorMenuSelected)
		y += rowH
	}

	// Пауза между волнами.
	if !ws.Active && ws.PauseTimer > 0 && !ws.IsVictory() {
		sec := ws.PauseTimer / 60

		// Спец-сообщение для defend-волны.
		switch ws.DefendOutcome {
		case 1:
			r.DrawScreenText("DEFEND SUCCESS", panelX+12, y, ColorMenuSelected)
			y += rowH
		case -1:
			r.DrawScreenText("SPAWN LOST", panelX+12, y, ColorHPLow)
			y += rowH
		default:
			r.DrawScreenText(fmt.Sprintf("Next wave in %ds", sec),
				panelX+12, y, ColorMenuText)
			y += rowH
		}
	}

	// Состав волны.
	for _, l := range lines {
		label := fmt.Sprintf("%-9s x%d", l.label, l.count)
		r.DrawScreenText(label, panelX+12, y, l.c)
		y += rowH
	}

	y += 4

	// Прогресс.
	if ws.Kind == world.WaveKindTimer {
		r.DrawScreenText(fmt.Sprintf("Spawned: %d", ws.SpawnedCount),
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

	// Для Defend — сколько спавнов живо.
	if ws.Kind == world.WaveKindDefend {
		c := ColorMenuSelected
		if ws.DefendSpawnsAlive <= 1 {
			c = ColorHPLow
		}
		r.DrawScreenText(fmt.Sprintf("Spawns: %d / %d",
			ws.DefendSpawnsAlive, ws.DefendSpawnCount),
			panelX+12, y, c)
		y += slotH
	}

	// Статус текущей группы.
	if ws.GroupRemaining > 0 {
		r.DrawScreenText(fmt.Sprintf("Group: %d left", ws.GroupRemaining),
			panelX+12, y, ColorMenuDim)
	} else if ws.Kind == world.WaveKindCommon && ws.GroupsRemaining > 0 {
		r.DrawScreenText(fmt.Sprintf("Groups left: %d", ws.GroupsRemaining),
			panelX+12, y, ColorMenuDim)
	} else if ws.Kind == world.WaveKindTimer && ws.TimerGroupCooldown > 0 {
		sec := ws.TimerGroupCooldown / 60
		r.DrawScreenText(fmt.Sprintf("Next group: %ds", sec),
			panelX+12, y, ColorMenuDim)
	}

	// Индикатор установки турели инженером.
	engineerSettingUp := false
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.Type == world.EnemyEngineer && e.TurretSetupTimer > 0 {
			engineerSettingUp = true
			break
		}
	}
	if engineerSettingUp {
		r.DrawScreenTextScaled("ENGINEER SETTING UP TURRET",
			8, float64(world.ScreenHeight)-150, 1.6,
			color.NRGBA{R: 255, G: 100, B: 100, A: 255})
	}
}

// drawHPBar — полоса здоровья игрока внизу слева.
func drawHPBar(r SceneRenderer, hp, maxHP int) {
	const (
		barW = 220.0
		barH = 18.0
	)
	x := 8.0
	y := float64(world.ScreenHeight) - barH - 12

	r.DrawScreenRect(x-2, y-2, barW+4, barH+4, ColorHPBack)

	ratio := 0.0
	if maxHP > 0 {
		ratio = float64(hp) / float64(maxHP)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	c := ColorHPFull
	switch {
	case ratio < 0.3:
		c = ColorHPLow
	case ratio < 0.6:
		c = ColorHPMid
	}

	if ratio > 0 {
		r.DrawScreenRect(x, y, barW*ratio, barH, c)
	}
	r.DrawScreenText(fmt.Sprintf("HP %d / %d", hp, maxHP), x+6, y+6, ColorMenuText)
}
