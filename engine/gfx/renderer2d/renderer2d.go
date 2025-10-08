package renderer2d

import (
	"math"
	"strconv"

	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/core"
)

// Max textures per batch (common GL limit is 16)
const maxTexSlots = 16

// Vertex: pos2 + color4 + uv2 + texIndex1 => 9 floats
const vStride = 9
const vertsPerQuad = 4
const indsPerQuad = 6

var quadVertexLayout = core.VertexLayout{
	Stride: vStride * 4,
	Attributes: []core.VertexAttrib{
		{Location: 0, Size: 2, Type: core.AttribFloat32, Offset: 0},     // pos
		{Location: 1, Size: 4, Type: core.AttribFloat32, Offset: 2 * 4}, // color
		{Location: 2, Size: 2, Type: core.AttribFloat32, Offset: 6 * 4}, // uv
		{Location: 3, Size: 1, Type: core.AttribFloat32, Offset: 8 * 4}, // texIndex
	},
}

// Statistics captures the counts generated during a renderer frame.
type Statistics struct {
	DrawCalls    int
	QuadCount    int
	TextureCount int
}

// TotalVertexCount reports vertices submitted this frame.
func (s Statistics) TotalVertexCount() int { return s.QuadCount * vertsPerQuad }

// TotalIndexCount reports indices submitted this frame.
func (s Statistics) TotalIndexCount() int { return s.QuadCount * indsPerQuad }

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

	mesh     core.Mesh
	samplers map[string]core.Texture
	uniforms map[string]any
	texNames [maxTexSlots]string

	_vp           [16]float32
	stats         Statistics
	extraUniforms map[string]any
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

	// Create a reusable mesh large enough for the biggest batch.
	initialVerts := make([]float32, maxQuads*vertsPerQuad*vStride)
	initialInds := make([]uint32, maxQuads*indsPerQuad)
	mesh, err := r.CreateMesh(core.MeshDesc{
		Vertices: initialVerts,
		Indices:  initialInds,
		Layout:   quadVertexLayout,
	})
	if err != nil {
		return nil, err
	}
	rd.mesh = mesh

	rd.samplers = make(map[string]core.Texture, maxTexSlots)
	rd.uniforms = make(map[string]any, 4)
	for i := 0; i < maxTexSlots; i++ {
		rd.texNames[i] = "uTex[" + strconv.Itoa(i) + "]"
	}

	return rd, nil
}

func (rd *Renderer2D) BeginScene(vp [16]float32) {
	rd._vp = vp
	rd.stats = Statistics{}
	rd.resetBatch()
}

func (rd *Renderer2D) EndScene() { rd.flush() }

// Stats returns the current frame statistics snapshot.
func (rd *Renderer2D) Stats() Statistics { return rd.stats }

// SetUniform queues an additional uniform to be sent on every draw call.
// The uniform persists until overwritten; call with nil to remove.
func (rd *Renderer2D) SetUniform(name string, value any) {
	if rd.extraUniforms == nil {
		rd.extraUniforms = make(map[string]any)
	}
	if value == nil {
		delete(rd.extraUniforms, name)
		return
	}
	rd.extraUniforms[name] = value
}

// Draw solid color quad (uses white texture in slot 0)
func (rd *Renderer2D) DrawQuad(x, y, w, h float32, color colors.Color, rotationRad float32) {
	rd.ensureQuadCapacity()
	rd.drawQuadInternal(x, y, w, h, color, rotationRad, rd.texSlot(rd.white), 0, 0, 1, 1)
}

// Draw textured quad with UVs (tint color)
func (rd *Renderer2D) DrawTexturedQuad(x, y, w, h float32, tex core.Texture, tint colors.Color, rotationRad float32) {
	rd.ensureQuadCapacity()
	slot := rd.texSlot(tex)
	rd.drawQuadInternal(x, y, w, h, tint, rotationRad, slot, 0, 0, 1, 1)
}

// Draw textured sub-rect (UV rect: u0,v0 -> u1,v1)
func (rd *Renderer2D) DrawTexturedQuadUV(x, y, w, h float32, tex core.Texture, tint colors.Color, rotationRad float32, u0, v0, u1, v1 float32) {
	rd.ensureQuadCapacity()
	slot := rd.texSlot(tex)
	rd.drawQuadInternal(x, y, w, h, tint, rotationRad, slot, u0, v0, u1, v1)
}

// DrawSubTexQuad draws a quad using a SubTexture2D (tint + rotation optional).
func (rd *Renderer2D) DrawSubTexQuad(x, y, w, h float32, sub SubTexture2D, tint colors.Color, rotationRad float32) {
	rd.ensureQuadCapacity()
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
	}
	rd.texArr[rd.texCnt] = t
	rd.texCnt++
	rd.stats.TextureCount = rd.texCnt
	return float32(rd.texCnt - 1)
}

func (rd *Renderer2D) drawQuadInternal(x, y, w, h float32, color colors.Color, rotationRad float32, texIndex float32, u0, v0, u1, v1 float32) {
	halfW := w * 0.5
	halfH := h * 0.5

	// corners (TL, TR, BL, BR) with UVs. Positive Y goes down so top is -halfH.
	corners := [4][4]float32{
		{-halfW, -halfH, u0, v0},
		{halfW, -halfH, u1, v0},
		{-halfW, halfH, u0, v1},
		{halfW, halfH, u1, v1},
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
	rd.stats.QuadCount++
}

func (rd *Renderer2D) flush() {
	if rd.quadCount == 0 {
		return
	}

	if err := rd.r.UpdateMesh(rd.mesh, rd.verts, rd.inds); err != nil {
		panic(err)
	}

	for k := range rd.samplers {
		delete(rd.samplers, k)
	}
	for i := 0; i < rd.texCnt; i++ {
		rd.samplers[rd.texNames[i]] = rd.texArr[i]
	}

	for k := range rd.uniforms {
		delete(rd.uniforms, k)
	}
	rd.uniforms["uVP"] = rd._vp
	for k, v := range rd.extraUniforms {
		rd.uniforms[k] = v
	}

	rd.r.Draw(core.DrawCmd{
		Pipe:     rd.pipe,
		Mesh:     rd.mesh,
		Uniforms: rd.uniforms,
		Samplers: rd.samplers,
	})
	rd.stats.DrawCalls++

	rd.resetBatch()
}

func (rd *Renderer2D) resetBatch() {
	rd.verts = rd.verts[:0]
	rd.inds = rd.inds[:0]
	rd.quadCount = 0
	for i := range rd.texArr {
		rd.texArr[i] = nil
	}
	rd.texArr[0] = rd.white
	rd.texCnt = 1
}

func (rd *Renderer2D) ensureQuadCapacity() {
	if rd.quadCount >= rd.maxQuads {
		rd.flush()
	}
}
