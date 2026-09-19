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
