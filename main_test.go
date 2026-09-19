package main

import (
	"testing"
)

// ---------- Rect.Intersects ----------

func TestRectIntersects(t *testing.T) {
	tests := []struct {
		name string
		a, b Rect
		want bool
	}{
		{
			name: "identical rects",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{0, 0, 10, 10},
			want: true,
		},
		{
			name: "partial overlap",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{5, 5, 10, 10},
			want: true,
		},
		{
			name: "b inside a",
			a:    Rect{0, 0, 100, 100},
			b:    Rect{40, 40, 10, 10},
			want: true,
		},
		{
			name: "touching edges horizontally (no overlap)",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{10, 0, 10, 10},
			want: false,
		},
		{
			name: "touching edges vertically (no overlap)",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{0, 10, 10, 10},
			want: false,
		},
		{
			name: "far apart",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{100, 100, 10, 10},
			want: false,
		},
		{
			name: "a to the left of b",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{20, 0, 10, 10},
			want: false,
		},
		{
			name: "a above b",
			a:    Rect{0, 0, 10, 10},
			b:    Rect{0, 20, 10, 10},
			want: false,
		},
		{
			name: "zero-size rect inside another",
			a:    Rect{5, 5, 0, 0},
			b:    Rect{0, 0, 10, 10},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Intersects(tt.b); got != tt.want {
				t.Errorf("Intersects(%+v, %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
			// Симметричность: a∩b == b∩a
			if got := tt.b.Intersects(tt.a); got != tt.want {
				t.Errorf("Intersects not symmetric: (%+v, %+v) = %v, want %v",
					tt.b, tt.a, got, tt.want)
			}
		})
	}
}

// ---------- Player.Rect ----------

func TestPlayerRect(t *testing.T) {
	p := Player{X: 10, Y: 20}
	got := p.Rect()
	want := Rect{X: 10, Y: 20, W: playerSize, H: playerSize}

	if got != want {
		t.Errorf("Player.Rect() = %+v, want %+v", got, want)
	}
}

// ---------- tryMove: движение без препятствий ----------

func TestTryMove_FreeSpace(t *testing.T) {
	g := &Game{
		player: Player{X: 100, Y: 100},
		walls:  nil,
	}

	g.tryMove(5, 0)
	if g.player.X != 105 || g.player.Y != 100 {
		t.Errorf("after move (5,0): got (%.0f, %.0f), want (105, 100)",
			g.player.X, g.player.Y)
	}

	g.tryMove(0, -10)
	if g.player.X != 105 || g.player.Y != 90 {
		t.Errorf("after move (0,-10): got (%.0f, %.0f), want (105, 90)",
			g.player.X, g.player.Y)
	}
}

// ---------- tryMove: коллизия со стеной ----------

func TestTryMove_BlockedByWall(t *testing.T) {
	// Стена справа от игрока, вплотную.
	// Игрок: 100..124 (playerSize=24), стена начинается с x=124
	wall := Rect{X: 124, Y: 90, W: 50, H: 50}
	g := &Game{
		player: Player{X: 100, Y: 100},
		walls:  []Rect{wall},
	}

	// Попытка сдвинуться вправо на 1 — должна упереться в стену
	g.tryMove(1, 0)
	if g.player.X != 100 {
		t.Errorf("player should be blocked by wall: got X=%.0f, want 100", g.player.X)
	}

	// Попытка сдвинуться влево — стена не мешает
	g.tryMove(-5, 0)
	if g.player.X != 95 {
		t.Errorf("player should move left freely: got X=%.0f, want 95", g.player.X)
	}
}

// ---------- tryMove: скольжение вдоль стены ----------

func TestTryMove_SlideAlongWall(t *testing.T) {
	// Игрок прижат правым краем к вертикальной стене.
	// Игрок: X=100..124, Y=100..124
	// Стена:  X=124..224, Y=0..500
	// Правый край игрока в точности совпадает с левым краем стены —
	// это касание, а не перекрытие, поэтому стоять так можно.
	wall := Rect{X: 124, Y: 0, W: 100, H: 500}
	g := &Game{
		player: Player{X: 100, Y: 100},
		walls:  []Rect{wall},
	}

	// Вправо — блок: шаг на 10 даёт перекрытие с стеной.
	g.tryMove(10, 0)
	if g.player.X != 100 {
		t.Errorf("horizontal move into wall should be blocked: got X=%.0f, want 100", g.player.X)
	}

	// Вниз — проходит: это и есть скольжение вдоль стены.
	g.tryMove(0, 10)
	if g.player.Y != 110 {
		t.Errorf("vertical move should slide: got Y=%.0f, want 110", g.player.Y)
	}
	if g.player.X != 100 {
		t.Errorf("X should remain unchanged during slide: got X=%.0f, want 100", g.player.X)
	}
}

// TestTryMove_DiagonalSlide — воспроизводит то, как Update вызывает tryMove:
// сначала по X, потом по Y. Если движение вправо заблокировано, а вниз — нет,
// игрок должен "просочиться" вниз, не застряв.
func TestTryMove_DiagonalSlide(t *testing.T) {
	wall := Rect{X: 124, Y: 0, W: 100, H: 500}
	g := &Game{
		player: Player{X: 100, Y: 100},
		walls:  []Rect{wall},
	}

	// Диагональный интент: вправо-вниз.
	g.tryMove(10, 0) // блок
	g.tryMove(0, 10) // скольжение

	if g.player.X != 100 || g.player.Y != 110 {
		t.Errorf("diagonal slide: got (%.0f, %.0f), want (100, 110)",
			g.player.X, g.player.Y)
	}
}

// ---------- tryMove: границы экрана ----------

func TestTryMove_ScreenBounds(t *testing.T) {
	t.Run("left edge", func(t *testing.T) {
		g := &Game{player: Player{X: 5, Y: 100}}
		g.tryMove(-100, 0)
		if g.player.X != 0 {
			t.Errorf("X should clamp to 0, got %.0f", g.player.X)
		}
	})

	t.Run("top edge", func(t *testing.T) {
		g := &Game{player: Player{X: 100, Y: 5}}
		g.tryMove(0, -100)
		if g.player.Y != 0 {
			t.Errorf("Y should clamp to 0, got %.0f", g.player.Y)
		}
	})

	t.Run("right edge", func(t *testing.T) {
		g := &Game{player: Player{X: screenWidth - playerSize - 5, Y: 100}}
		g.tryMove(100, 0)
		wantX := float64(screenWidth - playerSize)
		if g.player.X != wantX {
			t.Errorf("X should clamp to %.0f, got %.0f", wantX, g.player.X)
		}
	})

	t.Run("bottom edge", func(t *testing.T) {
		g := &Game{player: Player{X: 100, Y: screenHeight - playerSize - 5}}
		g.tryMove(0, 100)
		wantY := float64(screenHeight - playerSize)
		if g.player.Y != wantY {
			t.Errorf("Y should clamp to %.0f, got %.0f", wantY, g.player.Y)
		}
	})
}

// ---------- tryMove: движение в угол ----------

func TestTryMove_IntoCorner(t *testing.T) {
	// Игрок упёрся в угол между двумя стенами. Ни одно движение внутрь угла
	// не должно сдвигать его.
	g := &Game{
		player: Player{X: 100, Y: 100},
		walls: []Rect{
			{X: 124, Y: 0, W: 50, H: 500}, // стена справа
			{X: 0, Y: 124, W: 500, H: 50}, // стена снизу
		},
	}

	g.tryMove(10, 0) // вправо — блок
	g.tryMove(0, 10) // вниз — блок
	if g.player.X != 100 || g.player.Y != 100 {
		t.Errorf("player should be stuck in corner: got (%.0f, %.0f), want (100, 100)",
			g.player.X, g.player.Y)
	}

	// А влево-вверх должен свободно уйти
	g.tryMove(-10, -10)
	if g.player.X != 90 || g.player.Y != 90 {
		t.Errorf("player should move away from corner: got (%.0f, %.0f), want (90, 90)",
			g.player.X, g.player.Y)
	}
}
