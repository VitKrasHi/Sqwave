package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 960
	screenHeight = 640

	playerSize   = 24.0
	playerSpeed  = 3.0
	dashSpeed    = 12.0
	dashDuration = 8  // тиков
	dashCooldown = 45 // тиков

	aimLineLength = 70.0
)

// Rect — простой AABB-прямоугольник. Используется и для стен, и для игрока.
type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Intersects(o Rect) bool {
	return r.X < o.X+o.W &&
		r.X+r.W > o.X &&
		r.Y < o.Y+o.H &&
		r.Y+r.H > o.Y
}

// Player — состояние игрока. В будущем это станет ECS-компонентами
// (Position, Velocity, Dash), но для прототипа хватит одной структуры.
type Player struct {
	X, Y              float64
	DashTimer         int
	DashCooldownTimer int
}

func (p *Player) Rect() Rect {
	return Rect{p.X, p.Y, playerSize, playerSize}
}

// Game — корневой объект, реализует ebiten.Game.
type Game struct {
	player     Player
	walls      []Rect
	aimX, aimY float64
}

func NewGame() *Game {
	return &Game{
		player: Player{
			X: screenWidth/2 - playerSize/2,
			Y: screenHeight/2 - playerSize/2,
		},
		walls: []Rect{
			{X: 80, Y: 80, W: 220, H: 30},
			{X: 400, Y: 160, W: 30, H: 320},
			{X: 600, Y: 80, W: 260, H: 30},
			{X: 120, Y: 380, W: 320, H: 30},
			{X: 680, Y: 420, W: 200, H: 30},
			{X: 80, Y: 480, W: 30, H: 120},
			{X: 850, Y: 200, W: 30, H: 200},
		},
	}
}

func (g *Game) Update() error {
	p := &g.player

	// --- Прицел ---
	mx, my := ebiten.CursorPosition()
	g.aimX, g.aimY = float64(mx), float64(my)

	// --- Таймеры рывка ---
	if p.DashCooldownTimer > 0 {
		p.DashCooldownTimer--
	}
	if p.DashTimer > 0 {
		p.DashTimer--
	}

	// --- Ввод направления ---
	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += 1
	}

	// Нормализация, чтобы по диагонали не бегать быстрее
	if len := math.Hypot(dx, dy); len > 0 {
		dx /= len
		dy /= len
	}

	// --- Старт рывка ---
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) &&
		p.DashCooldownTimer == 0 &&
		(dx != 0 || dy != 0) {
		p.DashTimer = dashDuration
		p.DashCooldownTimer = dashCooldown
	}

	speed := playerSpeed
	if p.DashTimer > 0 {
		speed = dashSpeed
	}

	// --- Движение с раздельным разрешением коллизий по осям ---
	// Это даёт «скольжение» вдоль стен, а не залипание.
	g.tryMove(dx*speed, 0)
	g.tryMove(0, dy*speed)

	return nil
}

// tryMove пробует сдвинуть игрока и откатывает сдвиг при столкновении.
// tryMove двигает игрока с раздельным разрешением коллизий по осям.
// При столкновении игрок «прижимается» к стене вплотную — без зазора.
func (g *Game) tryMove(dx, dy float64) {
	g.tryMoveX(dx)
	g.tryMoveY(dy)
}

func (g *Game) tryMoveX(dx float64) {
	if dx == 0 {
		return
	}
	p := &g.player
	targetX := p.X + dx

	for _, w := range g.walls {
		// Стена не пересекает игрока по Y — по X она не мешает.
		if p.Y >= w.Y+w.H || p.Y+playerSize <= w.Y {
			continue
		}
		if dx > 0 {
			// Стена должна быть строго справа от игрока.
			if w.X < p.X+playerSize {
				continue
			}
			// Не даём правому краю игрока зайти за левый край стены.
			if limit := w.X - playerSize; limit < targetX {
				targetX = limit
			}
		} else {
			// Стена должна быть строго слева от игрока.
			if w.X+w.W > p.X {
				continue
			}
			// Не даём левому краю игрока зайти за правый край стены.
			if limit := w.X + w.W; limit > targetX {
				targetX = limit
			}
		}
	}

	// Границы экрана.
	if targetX < 0 {
		targetX = 0
	}
	if targetX+playerSize > screenWidth {
		targetX = screenWidth - playerSize
	}

	p.X = targetX
}

func (g *Game) tryMoveY(dy float64) {
	if dy == 0 {
		return
	}
	p := &g.player
	targetY := p.Y + dy

	for _, w := range g.walls {
		// Стена не пересекает игрока по X — по Y она не мешает.
		if p.X >= w.X+w.W || p.X+playerSize <= w.X {
			continue
		}
		if dy > 0 {
			if w.Y < p.Y+playerSize {
				continue
			}
			if limit := w.Y - playerSize; limit < targetY {
				targetY = limit
			}
		} else {
			if w.Y+w.H > p.Y {
				continue
			}
			if limit := w.Y + w.H; limit > targetY {
				targetY = limit
			}
		}
	}

	// Границы экрана.
	if targetY < 0 {
		targetY = 0
	}
	if targetY+playerSize > screenHeight {
		targetY = screenHeight - playerSize
	}

	p.Y = targetY
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Чёрный фон
	screen.Fill(color.Black)

	// --- Стены ---
	wallColor := color.RGBA{90, 90, 90, 255}
	for _, w := range g.walls {
		ebitenutil.DrawRect(screen, w.X, w.Y, w.W, w.H, wallColor)
	}

	// --- Игрок ---
	p := &g.player
	var playerColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if p.DashTimer > 0 {
		playerColor = color.NRGBA{R: 180, G: 220, B: 255, A: 255}
	}
	ebitenutil.DrawRect(screen, p.X, p.Y, playerSize, playerSize, playerColor)

	// --- Линия прицела от центра игрока к курсору ---
	centerX := p.X + playerSize/2
	centerY := p.Y + playerSize/2

	dirX := g.aimX - centerX
	dirY := g.aimY - centerY
	if l := math.Hypot(dirX, dirY); l > 0 {
		dirX /= l
		dirY /= l
	}
	endX := centerX + dirX*aimLineLength
	endY := centerY + dirY*aimLineLength

	vector.StrokeLine(
		screen,
		float32(centerX), float32(centerY),
		float32(endX), float32(endY),
		2,
		color.RGBA{0, 255, 0, 255},
		false,
	)

	// --- HUD ---
	ebitenutil.DebugPrint(screen, "Sqwave — WASD: move, Space: dash")
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Sqwave")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}
