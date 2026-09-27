package geometry

// ClipSegmentByWalls возвращает конечную точку отрезка (x1,y1)→(x2,y2),
// обрезанного по ближайшему пересечению с любым из прямоугольников.
// Если пересечений нет, возвращает (x2, y2) без изменений.
func ClipSegmentByWalls(x1, y1, x2, y2 float64, walls []Rect) (float64, float64) {
	t := 1.0
	for _, w := range walls {
		if tt, ok := RaySegmentIntersectsRect(x1, y1, x2, y2, w); ok {
			if tt < t {
				t = tt
			}
		}
	}
	return x1 + (x2-x1)*t, y1 + (y2-y1)*t
}
