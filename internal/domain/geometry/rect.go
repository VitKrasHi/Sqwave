package geometry

type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Intersects(o Rect) bool {
	return r.X < o.X+o.W &&
		r.X+r.W > o.X &&
		r.Y < o.Y+o.H &&
		r.Y+r.H > o.Y
}

func RaySegmentIntersectsRectRaw(x1, y1, x2, y2, rx, ry, rw, rh float64) (float64, bool) {
	dx := x2 - x1
	dy := y2 - y1

	tmin, tmax := 0.0, 1.0

	if dx != 0 {
		tx1 := (rx - x1) / dx
		tx2 := (rx + rw - x1) / dx
		if tx1 > tx2 {
			tx1, tx2 = tx2, tx1
		}
		if tx1 > tmin {
			tmin = tx1
		}
		if tx2 < tmax {
			tmax = tx2
		}
	} else if x1 < rx || x1 > rx+rw {
		return 0, false
	}

	if dy != 0 {
		ty1 := (ry - y1) / dy
		ty2 := (ry + rh - y1) / dy
		if ty1 > ty2 {
			ty1, ty2 = ty2, ty1
		}
		if ty1 > tmin {
			tmin = ty1
		}
		if ty2 < tmax {
			tmax = ty2
		}
	} else if y1 < ry || y1 > ry+rh {
		return 0, false
	}

	if tmin > tmax {
		return 0, false
	}
	return tmin, true
}

func RaySegmentIntersectsRect(x1, y1, x2, y2 float64, r Rect) (float64, bool) {
	return RaySegmentIntersectsRectRaw(x1, y1, x2, y2, r.X, r.Y, r.W, r.H)
}
