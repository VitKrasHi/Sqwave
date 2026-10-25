package logo

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	colBgCenter = color.RGBA{R: 20, G: 30, B: 55, A: 255}
	colBgEdge   = color.RGBA{R: 4, G: 6, B: 12, A: 255}
	colRim      = color.RGBA{R: 120, G: 230, B: 255, A: 220}
	colRingDark = color.RGBA{R: 30, G: 90, B: 130, A: 255}
	colRingMid  = color.RGBA{R: 60, G: 160, B: 210, A: 255}
	colRingHot  = color.RGBA{R: 120, G: 230, B: 255, A: 255}
	colCore     = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	colText     = color.RGBA{R: 120, G: 230, B: 255, A: 255}
	colSub      = color.RGBA{R: 140, G: 160, B: 190, A: 255}
	colShadow   = color.RGBA{R: 0, G: 20, B: 40, A: 200}
)

const ringCount = 8

// SaveAvatar — круглая аватарка 512×512 с прозрачными углами.
func SaveAvatar(path string) error {
	return saveAvatar(path, true)
}

// SaveAvatarSquare — тот же дизайн, но квадратный (для иконок приложений).
func SaveAvatarSquare(path string) error {
	return saveAvatar(path, false)
}

func saveAvatar(path string, circle bool) error {
	const size = 512
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	cx, cy := 256.0, 256.0
	r := 240.0

	if circle {
		drawCircleBackground(img, cx, cy, r)
		drawCircleRim(img, cx, cy, r, 4)
	} else {
		const pad = 16
		const radius = 108
		drawRoundedSquareBackground(img, pad, pad, size-2*pad, size-2*pad, radius)
		drawRoundedSquareRim(img, pad, pad, size-2*pad, size-2*pad, radius, 4)
	}

	drawPortal(img, cx, cy)
	return savePNG(path, img)
}

// drawPortal — 8 концентрических окружностей со смещением по спирали.
// Создаёт эффект тоннеля/портала.
func drawPortal(img *image.RGBA, cx, cy float64) {
	// Внешний радиус ~160, внутренний ~30.
	const (
		outerR    = 175.0
		innerR    = 32.0
		offsetR   = 22.0 // радиус вращения смещения
		ringStep  = 0.5  // толщина кольца в пикселях
		angleStep = 0.44 // ~25° между кольцами
	)

	for i := 0; i < ringCount; i++ {
		t := float64(i) / float64(ringCount-1) // 0..1 (0=снаружи, 1=внутри)
		r := outerR*(1-t) + innerR*t
		angle := float64(i) * angleStep

		ox := cx + math.Cos(angle)*offsetR*t // смещение растёт к центру
		oy := cy + math.Sin(angle)*offsetR*t

		// Цвет кольца: от тёмно-бирюзового к ярко-бирюзовому.
		col := lerpColor(colRingDark, colRingHot, t)

		// Толщина кольца растёт к центру.
		thick := 2.5 + 2.5*t

		// Альфа растёт к центру.
		alpha := 0.55 + 0.45*t

		drawRingAA(img, ox, oy, r, thick, col, alpha)
		_ = ringStep
	}

	// Центральное ядро — яркая точка с гало.
	drawGlowDot(img, cx, cy, 14, colCore)
	drawGlowDot(img, cx, cy, 24, color.RGBA{R: 120, G: 230, B: 255, A: 80})
}

// drawRingAA — кольцо (окружность заданной толщины) с антиалиасингом.
func drawRingAA(img *image.RGBA, cx, cy, r, thick float64, col color.RGBA, alpha float64) {
	inner := r - thick/2
	outer := r + thick/2

	x0 := int(cx - outer - 1)
	x1 := int(cx + outer + 1)
	y0 := int(cy - outer - 1)
	y1 := int(cy + outer + 1)

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= img.Bounds().Dx() {
		x1 = img.Bounds().Dx() - 1
	}
	if y1 >= img.Bounds().Dy() {
		y1 = img.Bounds().Dy() - 1
	}

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			if d < inner-1 || d > outer+1 {
				continue
			}

			var a float64
			switch {
			case d < inner:
				a = 1 - (inner-d)/1.0
			case d > outer:
				a = 1 - (d-outer)/1.0
			default:
				a = 1
			}
			if a < 0 {
				a = 0
			}
			if a > 1 {
				a = 1
			}
			blendPixel(img, x, y, col, a*alpha)
		}
	}
}

// drawGlowDot — круг с мягким гало, антиалиасинг.
func drawGlowDot(img *image.RGBA, cx, cy, r float64, col color.RGBA) {
	x0 := int(cx - r - 1)
	x1 := int(cx + r + 1)
	y0 := int(cy - r - 1)
	y1 := int(cy + r + 1)

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 >= img.Bounds().Dx() {
		x1 = img.Bounds().Dx() - 1
	}
	if y1 >= img.Bounds().Dy() {
		y1 = img.Bounds().Dy() - 1
	}

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			if d > r+1 {
				continue
			}
			a := 1.0
			if d > r {
				a = 1 - (d - r)
			}
			blendPixel(img, x, y, col, a)
		}
	}
}

// SaveWordmark — горизонтальный логотип: иконка + "ViForGame".
func SaveWordmark(path string) error {
	const W, H = 1600, 400
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	iconSize := 260
	iconX := 40
	iconY := (H - iconSize) / 2
	drawIconInto(img, iconX, iconY, iconSize)

	text := "ViForGame"
	textScale := 11
	textImg := renderTextToImage(text, colText)
	textShadow := renderTextToImage(text, colShadow)
	scaledText := scaleNearest(textImg, textScale)
	scaledShadow := scaleNearest(textShadow, textScale)

	tw := scaledText.Bounds().Dx()
	th := scaledText.Bounds().Dy()

	textX := iconX + iconSize + 50
	textY := H/2 - th/2 - 20

	draw.Draw(img,
		image.Rect(textX+4, textY+4, textX+4+tw, textY+4+th),
		scaledShadow, image.Point{}, draw.Over)
	draw.Draw(img,
		image.Rect(textX, textY, textX+tw, textY+th),
		scaledText, image.Point{}, draw.Over)

	subScale := 3
	sub := "game studio"
	subImg := renderTextToImage(sub, colSub)
	scaledSub := scaleNearest(subImg, subScale)
	sw := scaledSub.Bounds().Dx()
	sh := scaledSub.Bounds().Dy()
	subX := textX + 4
	subY := textY + th + 16
	draw.Draw(img,
		image.Rect(subX, subY, subX+sw, subY+sh),
		scaledSub, image.Point{}, draw.Over)

	// Декоративные тире.
	dashY := textY + th/2 - 3
	dashW := 36
	dashH := 6
	gap := 20
	fillRect(img, textX-tw/8-gap-dashW, dashY, dashW, dashH, colText)
	fillRect(img, textX+tw+gap, dashY, dashW, dashH, colText)

	return savePNG(path, img)
}

// drawIconInto — рисует портал в прямоугольник (x, y, size).
func drawIconInto(img *image.RGBA, x, y, size int) {
	cx := float64(x) + float64(size)/2
	cy := float64(y) + float64(size)/2
	r := float64(size) / 2

	drawCircleBackground(img, cx, cy, r)
	drawCircleRim(img, cx, cy, r, 4)

	// Портал масштабирован.
	s := float64(size) / 512.0
	const (
		outerR    = 175.0
		innerR    = 32.0
		offsetR   = 22.0
		angleStep = 0.44
	)

	for i := 0; i < ringCount; i++ {
		t := float64(i) / float64(ringCount-1)
		rr := (outerR*(1-t) + innerR*t) * s
		angle := float64(i) * angleStep
		ox := cx + math.Cos(angle)*offsetR*t*s
		oy := cy + math.Sin(angle)*offsetR*t*s
		col := lerpColor(colRingDark, colRingHot, t)
		thick := (2.5 + 2.5*t) * s
		alpha := 0.55 + 0.45*t
		drawRingAA(img, ox, oy, rr, thick, col, alpha)
	}

	drawGlowDot(img, cx, cy, 14*s, colCore)
	drawGlowDot(img, cx, cy, 24*s, color.RGBA{R: 120, G: 230, B: 255, A: 80})
}

// ---------- Фон ----------

func drawCircleBackground(img *image.RGBA, cx, cy, r float64) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			if d > r {
				continue
			}
			t := d / r
			img.SetRGBA(x, y, lerpColor(colBgCenter, colBgEdge, t))
		}
	}
}

func drawCircleRim(img *image.RGBA, cx, cy, r, thick float64) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			if d < r-thick || d > r {
				continue
			}
			edgeDist := math.Min(d-(r-thick), r-d)
			alpha := edgeDist / 1.5
			if alpha > 1 {
				alpha = 1
			}
			blendPixel(img, x, y, colRim, alpha)
		}
	}
}

func drawRoundedSquareBackground(img *image.RGBA, x, y, w, h, radius int) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if !insideRoundedRect(i, j, x, y, w, h, radius) {
				continue
			}
			cx := float64(x) + float64(w)/2
			cy := float64(y) + float64(h)/2
			dx := float64(i) - cx
			dy := float64(j) - cy
			d := math.Sqrt(dx*dx+dy*dy) / (float64(w) / 2)
			if d > 1 {
				d = 1
			}
			img.SetRGBA(i, j, lerpColor(colBgCenter, colBgEdge, d))
		}
	}
}

func drawRoundedSquareRim(img *image.RGBA, x, y, w, h, radius, thick int) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			in := insideRoundedRect(i, j, x, y, w, h, radius)
			inInner := insideRoundedRect(i, j, x+thick, y+thick, w-2*thick, h-2*thick, radius-thick)
			if in && !inInner {
				blendPixel(img, i, j, colRim, 1.0)
			}
		}
	}
}

func insideRoundedRect(px, py, x, y, w, h, radius int) bool {
	if px < x || px >= x+w || py < y || py >= y+h {
		return false
	}
	if px >= x+radius && px <= x+w-radius-1 {
		return true
	}
	if py >= y+radius && py <= y+h-radius-1 {
		return true
	}
	corners := [][2]int{
		{x + radius, y + radius},
		{x + w - radius - 1, y + radius},
		{x + radius, y + h - radius - 1},
		{x + w - radius - 1, y + h - radius - 1},
	}
	for _, c := range corners {
		dx := px - c[0]
		dy := py - c[1]
		if dx*dx+dy*dy <= radius*radius {
			return true
		}
	}
	return false
}

// ---------- Утилиты ----------

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func blendPixel(img *image.RGBA, x, y int, col color.RGBA, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	dst := img.RGBAAt(x, y)
	a := float64(col.A) / 255.0 * alpha
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(col.R)*a + float64(dst.R)*(1-a)),
		G: uint8(float64(col.G)*a + float64(dst.G)*(1-a)),
		B: uint8(float64(col.B)*a + float64(dst.B)*(1-a)),
		A: 255,
	})
}

func fillRect(img *image.RGBA, x, y, w, h int, col color.RGBA) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if i < 0 || i >= img.Bounds().Dx() || j < 0 || j >= img.Bounds().Dy() {
				continue
			}
			img.Set(i, j, col)
		}
	}
}

func renderTextToImage(text string, col color.Color) *image.RGBA {
	w := len(text) * 7
	h := 13
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, 11),
	}
	d.DrawString(text)
	return img
}

func scaleNearest(src *image.RGBA, factor int) *image.RGBA {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, srcW*factor, srcH*factor))
	for y := 0; y < srcH*factor; y++ {
		for x := 0; x < srcW*factor; x++ {
			dst.Set(x, y, src.At(x/factor, y/factor))
		}
	}
	return dst
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
