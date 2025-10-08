package colors

type Color [4]float32

var (
	White    = Color{1, 1, 1, 1}
	Red      = Color{1, 0, 0, 1}
	Green    = Color{0, 1, 0, 1}
	Blue     = Color{0, 0, 1, 1}
	Black    = Color{0, 0, 0, 1}
	Magenta  = Color{1, 0, 1, 1}
	Cyan     = Color{0, 1, 1, 1}
	Yellow   = Color{1, 1, 0, 1}
	Gray     = Color{0.5, 0.5, 0.5, 1}
	DarkGray = Color{0.08, 0.10, 0.12, 1}
)

func (c Color) WithAlpha(a float32) Color {
	c[3] = a
	return c
}
