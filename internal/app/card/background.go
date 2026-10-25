package card

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

// Палитра — та же, что в particles.go.
var menuPalette = []color.RGBA{
	{R: 200, G: 60, B: 60, A: 255},   // Enemy
	{R: 160, G: 90, B: 220, A: 255},  // Shooter
	{R: 90, G: 220, B: 130, A: 255},  // Scout
	{R: 100, G: 180, B: 255, A: 255}, // Medic
	{R: 180, G: 40, B: 60, A: 255},   // Rammer
	{R: 160, G: 40, B: 90, A: 255},   // Sniper
	{R: 200, G: 140, B: 40, A: 255},  // Rocketeer
}

// SaveBackground генерирует фоновое изображение 1920×1080 из партиклов,
// визуально повторяющее анимацию главного меню.
func SaveBackground(path string) error {
	const (
		W = 1920
		H = 1080

		count    = 150
		linkDist = 250.0
	)

	rng := rand.New(rand.NewSource(42))
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	// Тёмный фон.
	fillRect(img, 0, 0, W, H, color.RGBA{R: 10, G: 12, B: 20, A: 255})

	type particle struct {
		X, Y, Size float64
		Col        color.RGBA
	}

	ps := make([]particle, count)
	for i := range ps {
		ps[i] = particle{
			X:    rng.Float64() * float64(W),
			Y:    rng.Float64() * float64(H),
			Size: 6 + rng.Float64()*8,
			Col:  menuPalette[rng.Intn(len(menuPalette))],
		}
	}

	// Соединительные линии (аддитивное наложение).
	for i := 0; i < len(ps); i++ {
		for j := i + 1; j < len(ps); j++ {
			dx := ps[i].X - ps[j].X
			dy := ps[i].Y - ps[j].Y
			d := math.Sqrt(dx*dx + dy*dy)
			if d > linkDist {
				continue
			}
			strength := 1 - d/linkDist
			intensity := uint8(120 * strength)

			drawLineAdditive(
				img,
				int(ps[i].X), int(ps[i].Y),
				int(ps[j].X), int(ps[j].Y),
				100, 160, 200, intensity,
			)
		}
	}

	// Сами частицы — мягкое свечение + тело.
	for _, p := range ps {
		// Внешнее свечение (два слоя).
		drawSquareAdditive(img,
			int(p.X), int(p.Y), int(p.Size*1.8),
			p.Col.R, p.Col.G, p.Col.B, 40)
		drawSquareAdditive(img,
			int(p.X), int(p.Y), int(p.Size*1.3),
			p.Col.R, p.Col.G, p.Col.B, 80)

		// Тело.
		half := int(p.Size / 2)
		fillRect(img,
			int(p.X)-half, int(p.Y)-half,
			int(p.Size), int(p.Size),
			color.RGBA{R: p.Col.R, G: p.Col.G, B: p.Col.B, A: 230})
	}

	// Виньетка по краям.
	applyVignette(img)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// ---------- Примитивы ----------

// addPixel — аддитивное наложение с ограничением по 255.
func addPixel(img *image.RGBA, x, y int, r, g, b uint8) {
	if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
		return
	}
	dst := img.RGBAAt(x, y)
	img.SetRGBA(x, y, color.RGBA{
		R: clampU8(int(dst.R) + int(r)),
		G: clampU8(int(dst.G) + int(g)),
		B: clampU8(int(dst.B) + int(b)),
		A: 255,
	})
}

// drawLineAdditive — линия Брезенхема с аддитивным смешиванием.
func drawLineAdditive(img *image.RGBA, x0, y0, x1, y1 int, r, g, b, a uint8) {
	// Масштабируем RGB на интенсивность a.
	rr := uint8(int(r) * int(a) / 255)
	gg := uint8(int(g) * int(a) / 255)
	bb := uint8(int(b) * int(a) / 255)

	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy

	for {
		addPixel(img, x0, y0, rr, gg, bb)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// drawSquareAdditive — квадрат с центром (cx, cy) и аддитивным наложением.
func drawSquareAdditive(img *image.RGBA, cx, cy, size int, r, g, b, a uint8) {
	rr := uint8(int(r) * int(a) / 255)
	gg := uint8(int(g) * int(a) / 255)
	bb := uint8(int(b) * int(a) / 255)
	half := size / 2
	for y := cy - half; y <= cy+half; y++ {
		for x := cx - half; x <= cx+half; x++ {
			addPixel(img, x, y, rr, gg, bb)
		}
	}
}

func applyVignette(img *image.RGBA) {
	W := img.Bounds().Dx()
	H := img.Bounds().Dy()
	cx := float64(W) / 2
	cy := float64(H) / 2
	maxD := math.Sqrt(cx*cx + cy*cy)

	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx+dy*dy) / maxD
			if d < 0.55 {
				continue
			}
			t := (d - 0.55) / 0.45
			if t > 1 {
				t = 1
			}
			factor := 1.0 - t*0.75

			pixel := img.RGBAAt(x, y)
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(pixel.R) * factor),
				G: uint8(float64(pixel.G) * factor),
				B: uint8(float64(pixel.B) * factor),
				A: 255,
			})
		}
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clampU8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
