package glbackend

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/hubastard/grove/engine/core"
)

type RendererGL struct {
	win     core.Window
	program uint32
	vao     uint32
	vbo     uint32
}

func NewRendererGL(win core.Window, _ core.Config) (*RendererGL, error) {
	r := &RendererGL{win: win}
	if err := r.Init(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RendererGL) Init() error {
	// Create simple shader program
	var err error
	r.program, err = makeProgram(vertexSource, fragmentSource)
	if err != nil {
		return err
	}

	// Triangle vertices: pos (x,y), color (r,g,b)
	verts := []float32{
		//  X,     Y,     R,   G,   B
		0.0, 0.6, 1.0, 0.2, 0.2,
		-0.6, -0.6, 0.2, 1.0, 0.2,
		0.6, -0.6, 0.2, 0.2, 1.0,
	}

	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, int(len(verts))*4, gl.Ptr(verts), gl.STATIC_DRAW)

	// layout(location = 0) in vec2 aPos;
	// layout(location = 1) in vec3 aColor;
	const stride = 5 * 4 // bytes
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(0)))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(2*4)))

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// Enable depth (not necessary for the triangle, but good default)
	gl.Enable(gl.DEPTH_TEST)
	return nil
}

func (r *RendererGL) Shutdown() {
	if r.vbo != 0 {
		gl.DeleteBuffers(1, &r.vbo)
	}
	if r.vao != 0 {
		gl.DeleteVertexArrays(1, &r.vao)
	}
	if r.program != 0 {
		gl.DeleteProgram(r.program)
	}
}

func (r *RendererGL) Resize(w, h int) {
	gl.Viewport(0, 0, int32(w), int32(h))
}

func (r *RendererGL) Clear(rf, gf, bf, af float32) {
	gl.ClearColor(rf, gf, bf, af)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *RendererGL) DrawDemoTriangle() {
	gl.UseProgram(r.program)
	gl.BindVertexArray(r.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

// --- Shader utilities ---

const vertexSource = `
#version 330 core
layout(location=0) in vec2 aPos;
layout(location=1) in vec3 aColor;
out vec3 vColor;
void main() {
    vColor = aColor;
    gl_Position = vec4(aPos, 0.0, 1.0);
}
` + "\x00"

const fragmentSource = `
#version 330 core
in vec3 vColor;
out vec4 FragColor;
void main() {
    FragColor = vec4(vColor, 1.0);
}
` + "\x00"

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
