package text

import "github.com/hubastard/grove/engine/gfx/renderer2d"

// DrawText draws s with baseline origin (x,y). Positive Y goes downward (matching the 2D projection).
func DrawText(r2d *renderer2d.Renderer2D, font *Font, x, y float32, s string, color [4]float32) {
	penX := x
	baseY := y + font.Ascent // move origin to top left
	var prev rune = -1

	for _, r := range s {
		if r == '\n' {
			penX = x
			// move baseline *down* for next line
			lineH := font.Ascent - font.Descent + font.LineGap
			baseY += lineH
			prev = -1
			continue
		}

		g, ok := font.Glyphs[r]
		if !ok {
			if sp, ok2 := font.Glyphs[' ']; ok2 {
				penX += sp.Advance
			}
			prev = r
			continue
		}

		// Apply kerning between prev and current
		if prev >= 0 && font.Face != nil {
			penX += float32(font.Face.Kern(prev, r)) / 64.0
		}

		// Baseline-aligned quad center (Y-down system)
		// top = baseline - BearingY
		left := penX + g.BearingX
		top := baseY - g.BearingY
		cx := left + float32(g.W)*0.5
		cy := top + float32(g.H)*0.5

		r2d.DrawTexturedQuadUV(
			cx, cy,
			float32(g.W), float32(g.H),
			font.Texture, color, 0,
			g.U0, g.V0, g.U1, g.V1,
		)

		penX += g.Advance
		prev = r
	}
}

func MeasureText(font *Font, s string, size float32) (width, height float32) {
	var lineW float32
	var prev rune = -1
	lineH := font.Ascent - font.Descent + font.LineGap
	height = lineH

	scale := size / font.SizePx

	for _, r := range s {
		if r == '\n' {
			if lineW > width {
				width = lineW
			}
			lineW = 0
			height += lineH
			prev = -1
			continue
		}

		g, ok := font.Glyphs[r]
		if !ok {
			if sp, ok2 := font.Glyphs[' ']; ok2 {
				lineW += sp.Advance
			}
			prev = r
			continue
		}

		if prev >= 0 && font.Face != nil {
			lineW += float32(font.Face.Kern(prev, r)) / 64.0
		}

		lineW += g.Advance
		prev = r
	}

	if lineW > width {
		width = lineW
	}
	return width * scale, height * scale
}

// Baseline-to-top distance (useful to position text by top-left).
func BaselineToTop(font *Font) float32    { return font.Ascent }
func BaselineToBottom(font *Font) float32 { return -font.Descent }
func LineHeight(font *Font) float32       { return font.Ascent - font.Descent + font.LineGap }
