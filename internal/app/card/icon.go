package card

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func saveIconInternal(path string) error {
	W := int(cardDesignW)
	H := int(cardDesignH)
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	// Фон.
	fillRect(img, 0, 0, W, H, toRGBA(colorBg))

	// Рамка.
	const bw = 2
	bc := toRGBA(colorBorder)
	fillRect(img, 0, 0, W, bw, bc)
	fillRect(img, 0, H-bw, W, bw, bc)
	fillRect(img, 0, 0, bw, H, bc)
	fillRect(img, W-bw, 0, bw, H, bc)

	// Заголовок.
	titleScale := int(cardTitleScale)
	titleImg := renderTextToImage(cardTitle, toRGBA(colorTitle))
	shadowImg := renderTextToImage(cardTitle, color.RGBA{R: 0, G: 60, B: 100, A: 200})

	scaledTitle := scaleNearest(titleImg, titleScale)
	scaledShadow := scaleNearest(shadowImg, titleScale)

	titleW := scaledTitle.Bounds().Dx()
	titleH := scaledTitle.Bounds().Dy()
	titleX := (W - titleW) / 2
	titleY := 55

	draw.Draw(img,
		image.Rect(titleX+3, titleY+3, titleX+3+titleW, titleY+3+titleH),
		scaledShadow, image.Point{}, draw.Over)
	draw.Draw(img,
		image.Rect(titleX, titleY, titleX+titleW, titleY+titleH),
		scaledTitle, image.Point{}, draw.Over)

	// Подчёркивание.
	underlineY := titleY + 13*titleScale + 8
	underlineW := int(float64(titleW) * 1.15)
	underlineX := (W - underlineW) / 2
	fillRect(img, underlineX, underlineY, underlineW, 2, toRGBA(colorTitle))

	// Боковые линии.
	dashY := titleY + 13*titleScale/2 - 1
	dashW := 30
	dashH := 3
	gap := 16

	leftDashX := titleX - dashW - gap
	rightDashX := titleX + titleW + gap
	fillRect(img, leftDashX, dashY, dashW, dashH, toRGBA(colorTitle))
	fillRect(img, rightDashX, dashY, dashW, dashH, toRGBA(colorTitle))

	// Игрок.
	pcx := int(cardPlayerCX)
	pcy := int(cardPlayerCY)
	ps := int(cardPlayerSize)

	fillRect(img, pcx-ps/2, pcy-ps/2, ps, ps,
		color.RGBA{R: 255, G: 255, B: 255, A: 255})

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// ---------- helpers ----------

func toRGBA(c color.NRGBA) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

func fillRect(img *image.RGBA, x, y, w, h int, col color.RGBA) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
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
