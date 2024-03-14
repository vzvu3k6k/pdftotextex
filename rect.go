package hyperpaper

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func IsOverlapping(a, b *Rect) bool {
	if a.X+a.Width < b.X || b.X+b.Width < a.X {
		return false
	}
	if a.Y+a.Height < b.Y || b.Y+b.Height < a.Y {
		return false
	}
	return true
}
