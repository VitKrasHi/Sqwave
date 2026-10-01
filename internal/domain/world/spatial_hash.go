package world

const spatialCellSize = 100.0

type SpatialHash struct {
	cols, rows int
	buckets    [][]int // индексы врагов
}

func NewSpatialHash(w, h float64) *SpatialHash {
	cols := int(w/spatialCellSize) + 1
	rows := int(h/spatialCellSize) + 1
	return &SpatialHash{
		cols:    cols,
		rows:    rows,
		buckets: make([][]int, cols*rows),
	}
}

func (s *SpatialHash) Clear() {
	for i := range s.buckets {
		s.buckets[i] = s.buckets[i][:0]
	}
}

func (s *SpatialHash) Insert(idx int, x, y float64) {
	cx := int(x / spatialCellSize)
	cy := int(y / spatialCellSize)
	if cx < 0 || cy < 0 || cx >= s.cols || cy >= s.rows {
		return
	}
	key := cy*s.cols + cx
	s.buckets[key] = append(s.buckets[key], idx)
}

// ForEachNear вызывает fn для каждого индекса в клетке и её 8 соседях.
func (s *SpatialHash) ForEachNear(x, y float64, fn func(idx int)) {
	cx := int(x / spatialCellSize)
	cy := int(y / spatialCellSize)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := cx+dx, cy+dy
			if nx < 0 || ny < 0 || nx >= s.cols || ny >= s.rows {
				continue
			}
			for _, idx := range s.buckets[ny*s.cols+nx] {
				fn(idx)
			}
		}
	}
}
