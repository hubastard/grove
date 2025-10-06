package text

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/hubastard/grove/engine/core"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type Glyph struct {
	Rune     rune
	Advance  float32 // pixels
	BearingX float32 // left bearing in pixels
	BearingY float32 // top bearing in pixels (distance from baseline to glyph top)
	W, H     int     // glyph bitmap size
	U0, V0   float32 // UVs in atlas
	U1, V1   float32
}

type FontAtlas struct {
	SizePx                   float32
	Ascent, Descent, LineGap float32
	Glyphs                   map[rune]Glyph
	Texture                  core.Texture
	AtlasW, AtlasH           int
	Face                     font.Face
	closeFace                func()
}

func (fa *FontAtlas) Close() {
	if fa != nil && fa.closeFace != nil {
		fa.closeFace()
		fa.closeFace = nil
	}
}

// LoadTTF builds a monochrome (white) glyph atlas (alpha coverage) and uploads it as RGBA texture.
func LoadTTF(r core.Renderer, ttfRelPath string, sizePx float32) (*FontAtlas, error) {
	path := filepath.Join("assets", "fonts", ttfRelPath)
	ttfData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read font: %w", err)
	}

	ft, err := opentype.Parse(ttfData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size: float64(sizePx), DPI: 72, Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("new face: %w", err)
	}

	// Metrics in pixels
	m := face.Metrics()
	ascent := float32(m.Ascent.Round())
	descent := float32(-m.Descent.Round())
	lineGap := float32(m.Height.Round()) - ascent + descent

	// Target rune set (ASCII 32..126). Expand later as needed.
	var runes []rune
	for r := rune(32); r <= rune(255); r++ {
		runes = append(runes, r)
	}

	// Measure all glyph bounds/advances to pack a simple shelf atlas
	type meas struct {
		r      rune
		w, h   int
		adv    float32
		bx, by float32
		bounds image.Rectangle
	}
	measure := make([]meas, 0, len(runes))
	for _, rr := range runes {
		br, adv, ok := face.GlyphBounds(rr)
		if !ok {
			continue
		}
		gb := image.Rect(0, 0, (br.Max.X - br.Min.X).Round(), (br.Max.Y - br.Min.Y).Round())
		measure = append(measure, meas{
			r: rr,
			w: gb.Dx(), h: gb.Dy(),
			adv:    float32(adv.Round()),
			bx:     float32(br.Min.X.Round()),
			by:     float32(-br.Min.Y.Round()), // distance from baseline to top
			bounds: gb,
		})
	}

	// Very simple shelf packer (rows). Start with 512^2 atlas and grow until everything fits.
	const padding = 20
	atlasSize := 512
	var pos map[rune]image.Point
	for {
		x, y, rowH := padding, padding, 0
		fits := true
		pos = make(map[rune]image.Point, len(measure))

		for _, g := range measure {
			if g.w == 0 || g.h == 0 {
				continue
			}
			// If the glyph itself plus padding cannot fit into the atlas dimension, grow it.
			if g.w+padding*2 > atlasSize || g.h+padding*2 > atlasSize {
				fits = false
				break
			}
			if x+g.w+padding > atlasSize {
				x = padding
				y += rowH + padding
				rowH = 0
			}
			if y+g.h+padding > atlasSize {
				fits = false
				break
			}
			pos[g.r] = image.Pt(x, y)
			x += g.w + padding
			if g.h > rowH {
				rowH = g.h
			}
		}

		if fits {
			break
		}
		atlasSize *= 2
		if atlasSize > 4096 {
			return nil, fmt.Errorf("font atlas too large (>%d)", 4096)
		}
	}

	// Build atlas RGBA: white glyphs with alpha coverage
	dst := image.NewRGBA(image.Rect(0, 0, atlasSize, atlasSize))
	// Transparent background
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)

	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.White,
		Face: face,
	}

	glyphs := make(map[rune]Glyph, len(measure))
	for _, g := range measure {
		p := pos[g.r]
		if g.w == 0 || g.h == 0 {
			glyphs[g.r] = Glyph{
				Rune: g.r, Advance: g.adv,
				BearingX: g.bx, BearingY: g.by,
				W: g.w, H: g.h,
			}
			continue
		}

		// Drawer expects a dot at the baseline; compute baseline y within the glyph rect.
		baseline := p.Y + int(g.by)

		// Render by drawing the single rune using DrawString
		drawer.Dot = fixed.P(p.X-int(g.bx), baseline) // shift left by bearingX
		drawer.DrawString(string(g.r))

		u0 := float32(p.X) / float32(atlasSize)
		v0 := float32(p.Y) / float32(atlasSize)
		u1 := float32(p.X+g.w) / float32(atlasSize)
		v1 := float32(p.Y+g.h) / float32(atlasSize)

		glyphs[g.r] = Glyph{
			Rune: g.r, Advance: g.adv,
			BearingX: g.bx, BearingY: g.by,
			W: g.w, H: g.h,
			U0: u0, V0: v0, U1: u1, V1: v1,
		}
	}

	kerning := make(map[rune]map[rune]float32)
	for _, a := range measure {
		for _, b := range measure {
			if dx := face.Kern(a.r, b.r); dx != 0 {
				if kerning[a.r] == nil {
					kerning[a.r] = make(map[rune]float32)
				}
				kerning[a.r][b.r] = float32(dx.Round())
			}
		}
	}

	// Upload atlas
	tex, err := r.CreateTexture(core.TextureDesc{
		Width: atlasSize, Height: atlasSize,
		Format: core.TextureRGBA8,
		// raw pixels
		Pixels:    dst.Pix,
		MinFilter: "nearest",
		MagFilter: "nearest",
		WrapU:     "clamp",
		WrapV:     "clamp",
	})
	if err != nil {
		return nil, err
	}

	return &FontAtlas{
		SizePx: sizePx,
		Ascent: ascent, Descent: descent, LineGap: lineGap,
		Glyphs:  glyphs,
		Texture: tex,
		AtlasW:  atlasSize, AtlasH: atlasSize,
		Face:      face,
		closeFace: func() { _ = face.Close() },
	}, nil
}
