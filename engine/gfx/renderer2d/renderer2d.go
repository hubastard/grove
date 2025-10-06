package renderer2d

import (
	"math"

	"github.com/hubastard/grove/engine/core"
)

type Renderer2D struct {
	r core.Renderer

	// pipeline and per-frame mesh
	pipe core.Pipeline

	// batch buffers
	verts     []float32
	inds      []uint32
	quadCount int

	maxQuads int

	_vp [16]float32
}

// Vertex: pos2 + color4  => 6 floats per vertex
const vStride = 6
const vertsPerQuad = 4
const indsPerQuad = 6

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
	return &Renderer2D{
		r: r, pipe: pipe, maxQuads: maxQuads,
		verts: make([]float32, 0, maxQuads*vertsPerQuad*vStride),
		inds:  make([]uint32, 0, maxQuads*indsPerQuad),
	}, nil
}

func (rd *Renderer2D) BeginScene(vp [16]float32) {
	// store VP as a uniform on draw; we just clear batches here
	rd.verts = rd.verts[:0]
	rd.inds = rd.inds[:0]
	rd.quadCount = 0
	rd._vp = vp
}

func (rd *Renderer2D) EndScene() { rd.flush() }

func (rd *Renderer2D) DrawQuad(x, y, w, h float32, color [4]float32, rotationRad float32) {
	if rd.quadCount >= rd.maxQuads {
		rd.flush()
	}
	halfW := w * 0.5
	halfH := h * 0.5

	// corners around origin
	corners := [][2]float32{
		{-halfW, halfH},  // tl
		{halfW, halfH},   // tr
		{-halfW, -halfH}, // bl
		{halfW, -halfH},  // br
	}

	c, s := float32(math.Cos(float64(rotationRad))), float32(math.Sin(float64(rotationRad)))

	startVertex := uint32(len(rd.verts) / vStride)

	// transform and append 4 verts: (pos2, color4)
	for _, p := range corners {
		// rotate then translate
		rx := p[0]*c - p[1]*s + x
		ry := p[0]*s + p[1]*c + y
		rd.verts = append(rd.verts,
			rx, ry,
			color[0], color[1], color[2], color[3],
		)
	}
	// indices (two triangles)
	rd.inds = append(rd.inds,
		startVertex+0, startVertex+2, startVertex+1,
		startVertex+1, startVertex+2, startVertex+3,
	)
	rd.quadCount++
}

// ---- internals ----
var quadVPUniform = "uVP"

func (rd *Renderer2D) flush() {
	if rd.quadCount == 0 {
		return
	}
	// Build a temporary mesh and draw once
	mesh, _ := rd.r.CreateMesh(core.MeshDesc{
		Vertices: rd.verts,
		Indices:  rd.inds,
		Layout: core.VertexLayout{
			Stride: vStride * 4,
			Attributes: []core.VertexAttrib{
				{Location: 0, Size: 2, Type: core.AttribFloat32, Offset: 0},
				{Location: 1, Size: 4, Type: core.AttribFloat32, Offset: 2 * 4},
			},
		},
	})
	rd.r.Draw(core.DrawCmd{
		Pipe: rd.pipe,
		Mesh: mesh,
		Uniforms: map[string]any{
			quadVPUniform: rd._vp,
		},
	})
	// reset batch
	rd.verts = rd.verts[:0]
	rd.inds = rd.inds[:0]
	rd.quadCount = 0
}

var _vp [16]float32 // stored per-batch (assigned in BeginScene)
