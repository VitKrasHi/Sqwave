package game

import (
	"fmt"
	"image/color"

	"Sqwave/internal/domain/world"
)

// DrawStatsOverlay рисует блок статистики по прохождению.
// startY — верхняя точка блока.
func DrawStatsOverlay(r SceneRenderer, stats *world.Stats, startY float64) {
	const (
		scale       = 1.2
		rowH        = 26.0
		colLabelW   = 240.0
		headerScale = 1.6
	)

	sw := float64(world.ScreenWidth)
	blockW := colLabelW + 100
	px := (sw - blockW) / 2

	header := "STATISTICS"
	hw := float64(len(header)) * 7.0 * headerScale
	r.DrawScreenTextScaled(header, (sw-hw)/2, startY, headerScale,
		color.NRGBA{R: 200, G: 220, B: 255, A: 255})

	type row struct{ label, value string }
	rows := []row{
		{"Damage Dealt", fmt.Sprintf("%d", stats.DamageDealt)},
		{"Enemies Killed", fmt.Sprintf("%d", stats.EnemiesKilled)},
		{"Waves Survived", fmt.Sprintf("%d", stats.WavesSurvived)},
		{"Coins Earned", fmt.Sprintf("%d", stats.CoinsEarned)},
		{"Playtime", formatPlaytime(stats.PlaytimeTicks)},
	}

	listY := startY + 40
	for i, rw := range rows {
		y := listY + float64(i)*rowH
		r.DrawScreenTextScaled(rw.label, px, y, scale,
			color.NRGBA{R: 180, G: 190, B: 210, A: 255})
		r.DrawScreenTextScaled(rw.value, px+colLabelW, y, scale,
			color.NRGBA{R: 240, G: 230, B: 180, A: 255})
	}
}

func formatPlaytime(ticks int) string {
	totalSec := ticks / 60
	min := totalSec / 60
	sec := totalSec % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}
