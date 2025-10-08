package glbackend

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/hubastard/grove/engine/core"
)

// ---------- GL handles implementing core.Mesh / core.Pipeline / core.Texture ----------

type meshGL struct {
	vao  uint32
	vbo  uint32
	ebo  uint32 // 0 if none
	nIdx int
	nVtx int
}

func (meshGL) IsMesh() {}

type pipeGL struct {
	prog      uint32
	depthTest bool
	blend     bool
}

func (pipeGL) IsPipeline() {}

type texGL struct {
	id   uint32
	unit int // texture unit to bind to (we'll assign on the fly)
	w, h int
}

func (texGL) IsTexture() {}

type RendererGL struct {
	win core.Window
}

func NewRendererGL(win core.Window, _ core.Config) (*RendererGL, error) {
	r := &RendererGL{win: win}
	if err := r.Init(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RendererGL) Init() error {
	gl.Enable(gl.DEPTH_TEST) // default on
	return nil
}
func (r *RendererGL) Shutdown()       {}
func (r *RendererGL) Resize(w, h int) { gl.Viewport(0, 0, int32(w), int32(h)) }

func (r *RendererGL) Clear(rf, gf, bf, af float32) {
	gl.ClearColor(rf, gf, bf, af)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *RendererGL) CreateMesh(desc core.MeshDesc) (core.Mesh, error) {
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(desc.Vertices)*4, gl.Ptr(desc.Vertices), gl.STATIC_DRAW)

	if len(desc.Indices) > 0 {
		gl.GenBuffers(1, &ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(desc.Indices)*4, gl.Ptr(desc.Indices), gl.STATIC_DRAW)
	}

	for _, a := range desc.Layout.Attributes {
		if a.Type != core.AttribFloat32 {
			return nil, fmt.Errorf("unsupported attrib type")
		}
		gl.EnableVertexAttribArray(uint32(a.Location))
		gl.VertexAttribPointer(uint32(a.Location), int32(a.Size), gl.FLOAT, false, int32(desc.Layout.Stride), unsafe.Pointer(uintptr(a.Offset)))
	}

	// Keep EBO bound to VAO association
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	nVtx := 0
	if desc.Layout.Stride > 0 && len(desc.Vertices) > 0 && len(desc.Indices) == 0 {
		nVtx = len(desc.Vertices) * 4 / desc.Layout.Stride
	}
	m := &meshGL{vao: vao, vbo: vbo, ebo: ebo, nIdx: len(desc.Indices), nVtx: nVtx}
	return m, nil
}

func (r *RendererGL) CreatePipeline(desc core.PipelineDesc) (core.Pipeline, error) {
	prog, err := makeProgram(desc.VertexSource, desc.FragmentSource)
	if err != nil {
		return nil, err
	}
	return &pipeGL{prog: prog, depthTest: desc.DepthTest, blend: desc.Blend}, nil
}

func (r *RendererGL) CreateTexture(desc core.TextureDesc) (core.Texture, error) {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)

	// Params
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, toGLFilter(desc.MinFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, toGLFilter(desc.MagFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, toGLWrap(desc.WrapU))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, toGLWrap(desc.WrapV))

	// Data
	if desc.Format != core.TextureRGBA8 {
		return nil, fmt.Errorf("only RGBA8 supported for now")
	}
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(desc.Width), int32(desc.Height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(desc.Pixels))

	gl.BindTexture(gl.TEXTURE_2D, 0)
	return &texGL{id: id, w: desc.Width, h: desc.Height}, nil
}

func (r *RendererGL) GPUVendor() string   { return gl.GoStr(gl.GetString(gl.VENDOR)) }
func (r *RendererGL) GPURenderer() string { return gl.GoStr(gl.GetString(gl.RENDERER)) }
func (r *RendererGL) GPUVersion() string  { return gl.GoStr(gl.GetString(gl.VERSION)) }

// ------- Helpers -------

func toGLFilter(f string) int32 {
	switch f {
	case "linear":
		return gl.LINEAR
	default:
		return gl.NEAREST
	}
}
func toGLWrap(w string) int32 {
	switch w {
	case "repeat":
		return gl.REPEAT
	default:
		return gl.CLAMP_TO_EDGE
	}
}

func (r *RendererGL) Draw(cmd core.DrawCmd) {
	p := cmd.Pipe.(*pipeGL)
	m := cmd.Mesh.(*meshGL)

	// state
	if p.depthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}
	if p.blend {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	} else {
		gl.Disable(gl.BLEND)
	}

	gl.UseProgram(p.prog)

	// uniforms (minimal)
	for name, v := range cmd.Uniforms {
		switch name {
		case "uMVP":
			if mat, ok := v.([16]float32); ok {
				loc := gl.GetUniformLocation(p.prog, gl.Str("uMVP\x00"))
				if loc >= 0 {
					gl.UniformMatrix4fv(loc, 1, false, &mat[0])
				}
			}
		case "uVP":
			if mat, ok := v.([16]float32); ok {
				loc := gl.GetUniformLocation(p.prog, gl.Str("uVP\x00"))
				if loc >= 0 {
					gl.UniformMatrix4fv(loc, 1, false, &mat[0])
				}
			}
		}
	}

	// samplers: bind textures to units 0..N in stable order of iteration
	unit := 0
	for name, t := range cmd.Samplers {
		tx := t.(*texGL)
		gl.ActiveTexture(uint32(gl.TEXTURE0 + unit))
		gl.BindTexture(gl.TEXTURE_2D, tx.id)

		// assign sampler uniform to this unit
		loc := gl.GetUniformLocation(p.prog, gl.Str((name + "\x00")))
		if loc >= 0 {
			gl.Uniform1i(loc, int32(unit))
		}
		unit++
	}
	// NOTE: we don't unbind here; next draw will overwrite bindings

	gl.BindVertexArray(m.vao)
	if m.ebo != 0 && m.nIdx > 0 {
		gl.DrawElements(gl.TRIANGLES, int32(m.nIdx), gl.UNSIGNED_INT, unsafe.Pointer(uintptr(0)))
	} else {
		count := m.nVtx
		if cmd.Count > 0 {
			count = cmd.Count
		}
		gl.DrawArrays(gl.TRIANGLES, 0, int32(count))
	}
	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

// ------------ shader helpers ------------

func makeShader(src string, shaderType uint32) (uint32, error) {
	sh := gl.CreateShader(shaderType)
	csrc, free := gl.Strs(src)
	defer free()
	gl.ShaderSource(sh, 1, csrc, nil)
	gl.CompileShader(sh)

	var status int32
	gl.GetShaderiv(sh, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(sh, gl.INFO_LOG_LENGTH, &logLen)
		log := strings.Repeat("\x00", int(logLen))
		gl.GetShaderInfoLog(sh, logLen, nil, gl.Str(log))
		return 0, fmt.Errorf("shader compile error: %s", log)
	}
	return sh, nil
}

func makeProgram(vsSrc, fsSrc string) (uint32, error) {
	vs, err := makeShader(vsSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fs, err := makeShader(fsSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		gl.DeleteShader(vs)
		return 0, err
	}
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLen)
		log := strings.Repeat("\x00", int(logLen))
		gl.GetProgramInfoLog(prog, logLen, nil, gl.Str(log))
		return 0, fmt.Errorf("program link error: %s", log)
	}
	return prog, nil
}
