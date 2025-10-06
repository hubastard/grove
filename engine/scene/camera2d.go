package scene

import "math"

// OrthoCamera2D provides an orthographic camera with position, rotation, zoom.
type OrthoCamera2D struct {
	Left, Right, Bottom, Top float32
	Near, Far                float32
	X, Y                     float32
	RotationRad              float32
	Zoom                     float32 // 1 = no zoom
	vp                       [16]float32
	dirty                    bool
}

func NewOrtho2D(width, height int) *OrthoCamera2D {
	halfW := float32(width) * 0.5
	halfH := float32(height) * 0.5
	c := &OrthoCamera2D{
		Left: -halfW, Right: halfW,
		Bottom: -halfH, Top: halfH,
		Near: -1, Far: 1,
		Zoom: 1,
	}
	c.Recalculate()
	return c
}

func (c *OrthoCamera2D) SetViewportPixels(w, h int) {
	halfW := float32(w) * 0.5
	halfH := float32(h) * 0.5
	c.Left, c.Right = -halfW, halfW
	c.Bottom, c.Top = -halfH, halfH
	c.dirty = true
}

func (c *OrthoCamera2D) Move(dx, dy float32) { c.X += dx; c.Y += dy; c.dirty = true }
func (c *OrthoCamera2D) Rotate(dRad float32) { c.RotationRad += dRad; c.dirty = true }
func (c *OrthoCamera2D) SetZoom(z float32) {
	if z < 0.05 {
		z = 0.05
	}
	c.Zoom = z
	c.dirty = true
}

func (c *OrthoCamera2D) VP() [16]float32 {
	if c.dirty {
		c.Recalculate()
	}
	return c.vp
}

func (c *OrthoCamera2D) Recalculate() {
	// Ortho scaled by Zoom
	z := c.Zoom
	proj := ortho(c.Left/z, c.Right/z, c.Bottom/z, c.Top/z, c.Near, c.Far)

	// Correct view for column-vector math:
	// view = R(-rot) · T(-pos)
	view := mul(
		rotateZ(-c.RotationRad),
		translate(-c.X, -c.Y, 0),
	)

	c.vp = mul(proj, view)
	c.dirty = false
}

// ---- tiny mat helpers (column-major, GLSL-style) ----

func translate(x, y, z float32) [16]float32 {
	return [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func rotateZ(a float32) [16]float32 {
	c := float32(math.Cos(float64(a)))
	s := float32(math.Sin(float64(a)))
	return [16]float32{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func ortho(l, r, b, t, n, f float32) [16]float32 {
	rl := 1 / (r - l)
	tb := 1 / (t - b)
	fn := 1 / (f - n)
	return [16]float32{
		2 * rl, 0, 0, 0,
		0, 2 * tb, 0, 0,
		0, 0, -2 * fn, 0,
		-(r + l) * rl, -(t + b) * tb, -(f + n) * fn, 1,
	}
}

func mul(a, b [16]float32) [16]float32 {
	var out [16]float32
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			out[i+4*j] = a[0+4*j]*b[i+0] + a[1+4*j]*b[i+4] + a[2+4*j]*b[i+8] + a[3+4*j]*b[i+12]
		}
	}
	return out
}
