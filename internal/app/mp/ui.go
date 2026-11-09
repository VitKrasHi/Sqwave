package mp

import (
	"image/color"
	"net"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/world"
)

// Общая палитра кооп-экранов — в тон главному меню.
var (
	colBG      = color.NRGBA{R: 12, G: 15, B: 24, A: 255}
	colPanel   = color.NRGBA{R: 20, G: 25, B: 36, A: 245}
	colBorder  = color.NRGBA{R: 60, G: 85, B: 120, A: 255}
	colText    = color.NRGBA{R: 220, G: 228, B: 240, A: 255}
	colDim     = color.NRGBA{R: 120, G: 135, B: 160, A: 255}
	colAccent  = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	colWarning = color.NRGBA{R: 255, G: 120, B: 90, A: 255}
	colGood    = color.NRGBA{R: 120, G: 230, B: 150, A: 255}
)

func drawFrame(r game.SceneRenderer, title string) {
	r.Clear(colBG)
	w := float64(world.ScreenWidth)
	r.DrawScreenRect(0, 0, w, 80, colPanel)
	r.DrawScreenRect(0, 80, w, 2, colAccent)
	r.DrawScreenTextScaled(title, 40, 24, 2.2, colAccent)
}

// localIP — первый не-loopback IPv4-адрес для подсказки «куда подключаться».
func localIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue
			}
			return ip4.String()
		}
	}
	return "127.0.0.1"
}
