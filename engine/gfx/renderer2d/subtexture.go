package renderer2d

import "github.com/hubastard/grove/engine/core"

// SubTexture2D describes a UV sub-rect of a full texture.
type SubTexture2D struct {
	Texture core.Texture
	U0, V0  float32 // top-left (after the shader’s V flip)
	U1, V1  float32 // bottom-right
}

// FromPixels builds a subtexture from pixel coordinates within an atlas.
func FromPixels(tex core.Texture, x, y, w, h, atlasW, atlasH int) SubTexture2D {
	// Convert to normalized UVs. We assume vertex shader flips V:
	// vUV = vec2(aUV.x, 1.0 - aUV.y)
	u0 := float32(x) / float32(atlasW)
	v0 := float32(y) / float32(atlasH)
	u1 := float32(x+w) / float32(atlasW)
	v1 := float32(y+h) / float32(atlasH)
	return SubTexture2D{Texture: tex, U0: u0, V0: v0, U1: u1, V1: v1}
}

// FromGrid builds a subtexture from tile grid coordinates (cx,cy) of cell size (cw,ch).
func FromGrid(tex core.Texture, cx, cy, cw, ch, atlasW, atlasH int) SubTexture2D {
	return FromPixels(tex, cx*cw, cy*ch, cw, ch, atlasW, atlasH)
}
