package renderer2d

import (
	"math"

	"github.com/hubastard/grove/engine/core"
)

// Max textures per batch (common GL limit is 16)
const maxTexSlots = 16

// Vertex: pos2 + color4 + uv2 + texIndex1 => 9 floats
const vStride = 9
const vertsPerQuad = 4
const indsPerQuad = 6

type Renderer2D struct {
	r      core.Renderer
	pipe   core.Pipeline
	white  core.Texture // 1x1 white (slot 0)
	texArr [maxTexSlots]core.Texture
	texCnt int

	verts     []float32
	inds      []uint32
	quadCount int
	maxQuads  int

	_vp [16]float32
}

// New creates renderer and compiles the shader pipeline.
func New(r core.Renderer, vertSrc, fragSrc string, maxQuads int) (*Renderer2D, error) {
	if maxQuads <= 0 {
		maxQuads = 10000
	}
	pipe, err := r.CreatePipeline(core.PipelineDesc{
		VertexSource:   vertSrc,
		FragmentSource: fragSrc,
		DepthTest:      false,
		Blend:          true,
	})
	if err != nil {
		return nil, err
	}

	// build 1x1 white texture
	whitePix := []byte{255, 255, 255, 255}
	white, err := r.CreateTexture(core.TextureDesc{
		Width: 1, Height: 1,
		Format:    core.TextureRGBA8,
		Pixels:    whitePix,
		MinFilter: "nearest", MagFilter: "nearest",
		WrapU: "clamp", WrapV: "clamp",
	})
	if err != nil {
		return nil, err
	}

	rd := &Renderer2D{
		r: r, pipe: pipe, white: white, maxQuads: maxQuads,
		verts: make([]float32, 0, maxQuads*vertsPerQuad*vStride),
		inds:  make([]uint32, 0, maxQuads*indsPerQuad),
	}
	return rd, nil
}

func (rd *Renderer2D) BeginScene(vp [16]float32) {
	rd._vp = vp
	rd.verts = rd.verts[:0]
	rd.inds = rd.inds[:0]
	rd.quadCount = 0

	// reset texture array; slot 0 reserved for white
	for i := range rd.texArr {
		rd.texArr[i] = nil
	}
	rd.texArr[0] = rd.white
	rd.texCnt = 1
}

func (rd *Renderer2D) EndScene() { rd.flush() }

// Draw solid color quad (uses white texture in slot 0)
func (rd *Renderer2D) DrawQuad(x, y, w, h float32, color [4]float32, rotationRad float32) {
	rd.drawQuadInternal(x, y, w, h, color, rotationRad, rd.texSlot(rd.white), 0, 0, 1, 1)
}

// Draw textured quad with UVs (tint color)
func (rd *Renderer2D) DrawTexturedQuad(x, y, w, h float32, tex core.Texture, tint [4]float32, rotationRad float32) {
	slot := rd.texSlot(tex)
	rd.drawQuadInternal(x, y, w, h, tint, rotationRad, slot, 0, 0, 1, 1)
}

// Draw textured sub-rect (UV rect: u0,v0 -> u1,v1)
func (rd *Renderer2D) DrawTexturedQuadUV(x, y, w, h float32, tex core.Texture, tint [4]float32, rotationRad float32, u0, v0, u1, v1 float32) {
	slot := rd.texSlot(tex)
	rd.drawQuadInternal(x, y, w, h, tint, rotationRad, slot, u0, v0, u1, v1)
}

// DrawSubTexQuad draws a quad using a SubTexture2D (tint + rotation optional).
func (rd *Renderer2D) DrawSubTexQuad(x, y, w, h float32, sub SubTexture2D, tint [4]float32, rotationRad float32) {
	slot := rd.texSlot(sub.Texture)
	rd.drawQuadInternal(x, y, w, h, tint, rotationRad, slot, sub.U0, sub.V0, sub.U1, sub.V1)
}

// --- internals ---

func (rd *Renderer2D) texSlot(t core.Texture) float32 {
	// already in array?
	for i := 0; i < rd.texCnt; i++ {
		if rd.texArr[i] == t {
			return float32(i)
		}
	}
	// need a new slot
	if rd.texCnt >= maxTexSlots {
		// flush and reset texture bindings
		rd.flush()
		rd.BeginScene(rd._vp)
	}
	rd.texArr[rd.texCnt] = t
	rd.texCnt++
	return float32(rd.texCnt - 1)
}

func (rd *Renderer2D) drawQuadInternal(x, y, w, h float32, color [4]float32, rotationRad float32, texIndex float32, u0, v0, u1, v1 float32) {
	if rd.quadCount >= rd.maxQuads {
		rd.flush()
		rd.BeginScene(rd._vp)
	}
	halfW := w * 0.5
	halfH := h * 0.5

	// corners (TL, TR, BL, BR) with UVs
	corners := [4][4]float32{
		{-halfW, halfH, u0, v0},
		{halfW, halfH, u1, v0},
		{-halfW, -halfH, u0, v1},
		{halfW, -halfH, u1, v1},
	}
	c, s := float32(math.Cos(float64(rotationRad))), float32(math.Sin(float64(rotationRad)))

	startVertex := uint32(len(rd.verts) / vStride)

	for _, p := range corners {
		rx := p[0]*c - p[1]*s + x
		ry := p[0]*s + p[1]*c + y
		u, v := p[2], p[3]
		rd.verts = append(rd.verts,
			rx, ry,
			color[0], color[1], color[2], color[3],
			u, v,
			texIndex,
		)
	}
	rd.inds = append(rd.inds,
		startVertex+0, startVertex+2, startVertex+1,
		startVertex+1, startVertex+2, startVertex+3,
	)
	rd.quadCount++
}

func (rd *Renderer2D) flush() {
	if rd.quadCount == 0 {
		return
	}

	// build mesh on the fly
	mesh, _ := rd.r.CreateMesh(core.MeshDesc{
		Vertices: rd.verts,
		Indices:  rd.inds,
		Layout: core.VertexLayout{
			Stride: vStride * 4,
			Attributes: []core.VertexAttrib{
				{Location: 0, Size: 2, Type: core.AttribFloat32, Offset: 0},     // pos
				{Location: 1, Size: 4, Type: core.AttribFloat32, Offset: 2 * 4}, // color
				{Location: 2, Size: 2, Type: core.AttribFloat32, Offset: 6 * 4}, // uv
				{Location: 3, Size: 1, Type: core.AttribFloat32, Offset: 8 * 4}, // texIndex
			},
		},
	})

	// bind samplers: uTex[0..N-1]
	sam := make(map[string]core.Texture, rd.texCnt)
	for i := 0; i < rd.texCnt; i++ {
		name := "uTex[" + itoa(i) + "]"
		sam[name] = rd.texArr[i]
	}

	rd.r.Draw(core.DrawCmd{
		Pipe: rd.pipe,
		Mesh: mesh,
		Uniforms: map[string]any{
			"uVP": rd._vp,
		},
		Samplers: sam,
	})

	// reset batch
	rd.verts = rd.verts[:0]
	rd.inds = rd.inds[:0]
	rd.quadCount = 0
}

// tiny int->string without fmt to avoid allocs in hot path
func itoa(i int) string {
	if i < 10 {
		return string('0' + byte(i))
	}
	// very small usage; fallback to simple build
	return []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"}[i]
}
