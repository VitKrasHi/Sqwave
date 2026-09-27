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

// RaySegmentIntersectsRect возвращает параметр t ∈ [0,1] первой точки
// пересечения отрезка (x1,y1)→(x2,y2) с прямоугольником r.
// Если пересечения нет — ok = false.
func RaySegmentIntersectsRect(x1, y1, x2, y2 float64, r Rect) (t float64, ok bool) {
	dx := x2 - x1
	dy := y2 - y1

	tmin, tmax := 0.0, 1.0

	// Ось X.
	if dx != 0 {
		tx1 := (r.X - x1) / dx
		tx2 := (r.X + r.W - x1) / dx
		if tx1 > tx2 {
			tx1, tx2 = tx2, tx1
		}
		if tx1 > tmin {
			tmin = tx1
		}
		if tx2 < tmax {
			tmax = tx2
		}
	} else if x1 < r.X || x1 > r.X+r.W {
		return 0, false
	}

	// Ось Y.
	if dy != 0 {
		ty1 := (r.Y - y1) / dy
		ty2 := (r.Y + r.H - y1) / dy
		if ty1 > ty2 {
			ty1, ty2 = ty2, ty1
		}
		if ty1 > tmin {
			tmin = ty1
		}
		if ty2 < tmax {
			tmax = ty2
		}
	} else if y1 < r.Y || y1 > r.Y+r.H {
		return 0, false
	}

	if tmin > tmax {
		return 0, false
	}
	return tmin, true
}
