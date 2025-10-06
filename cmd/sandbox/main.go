package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hubastard/grove/engine/core"
	glbackend "github.com/hubastard/grove/engine/gfx/gl"
	"github.com/hubastard/grove/engine/platform"
)

type SampleApp struct {
	frames       int
	lastFPSTitle time.Time

	mesh  core.Mesh
	pipe  core.Pipeline
	tex   core.Texture
	angle float32
}

func (a *SampleApp) OnStart(e *core.Engine) {
	log.Println("App start")
	a.lastFPSTitle = time.Now()

	// --- shaders (GLSL 330 core) ---
	vs := `#version 330 core
layout(location=0) in vec2 aPos;
layout(location=1) in vec3 aColor;
layout(location=2) in vec2 aUV;
uniform mat4 uMVP;
out vec3 vColor;
out vec2 vUV;
void main(){
    vColor = aColor;
    vUV = aUV;
    gl_Position = uMVP * vec4(aPos, 0.0, 1.0);
}` + "\x00"

	fs := `#version 330 core
in vec3 vColor;
in vec2 vUV;
out vec4 FragColor;
uniform sampler2D uTex0;
void main(){
    vec4 tex = texture(uTex0, vUV);
    FragColor = tex * vec4(vColor, 1.0);
}` + "\x00"

	var err error
	a.pipe, err = e.Renderer.CreatePipeline(core.PipelineDesc{
		VertexSource: vs, FragmentSource: fs,
		DepthTest: false, Blend: true,
	})
	if err != nil {
		panic(err)
	}

	// --- quad mesh (pos2, color3, uv2) + indices ---
	//   (-0.5,0.5)  (0.5,0.5)
	//       0---------1
	//       |       / |
	//       |     /   |
	//       |   /     |
	//       2---------3
	//   (-0.5,-0.5) (0.5,-0.5)
	verts := []float32{
		//   x,    y,   r,   g,   b,   u,   v
		-0.5, 0.5, 1.0, 0.8, 0.8, 0.0, 0.0, // 0
		0.5, 0.5, 0.8, 1.0, 0.8, 1.0, 0.0, // 1
		-0.5, -0.5, 0.8, 0.8, 1.0, 0.0, 1.0, // 2
		0.5, -0.5, 1.0, 1.0, 1.0, 1.0, 1.0, // 3
	}
	indices := []uint32{0, 2, 1, 1, 2, 3}
	layout := core.VertexLayout{
		Stride: 7 * 4,
		Attributes: []core.VertexAttrib{
			{Location: 0, Size: 2, Type: core.AttribFloat32, Offset: 0},
			{Location: 1, Size: 3, Type: core.AttribFloat32, Offset: 2 * 4},
			{Location: 2, Size: 2, Type: core.AttribFloat32, Offset: 5 * 4},
		},
	}
	a.mesh, err = e.Renderer.CreateMesh(core.MeshDesc{
		Vertices: verts,
		Indices:  indices,
		Layout:   layout,
	})
	if err != nil {
		panic(err)
	}

	// --- procedural checkerboard texture (RGBA8) ---
	w, h := 64, 64
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			// 8x8 checker
			if ((x/8)+(y/8))%2 == 0 {
				pix[i+0] = 230
				pix[i+1] = 230
				pix[i+2] = 230
				pix[i+3] = 255
			} else {
				pix[i+0] = 30
				pix[i+1] = 30
				pix[i+2] = 30
				pix[i+3] = 255
			}
		}
	}
	a.tex, err = e.Renderer.CreateTexture(core.TextureDesc{
		Width: w, Height: h,
		Format:    core.TextureRGBA8,
		Pixels:    pix,
		MinFilter: "nearest",
		MagFilter: "nearest",
		WrapU:     "repeat",
		WrapV:     "repeat",
	})
	if err != nil {
		panic(err)
	}
}

func (a *SampleApp) OnShutdown(e *core.Engine) { log.Println("App shutdown") }

func (a *SampleApp) OnUpdate(e *core.Engine, dt float64) {
	a.frames++
	if time.Since(a.lastFPSTitle) >= time.Second {
		elapsed := time.Since(a.lastFPSTitle).Seconds()
		fps := float64(a.frames) / elapsed
		e.Window.SetTitle(fmt.Sprintf("Go Engine — ~%.0f FPS", fps))
		a.frames = 0
		a.lastFPSTitle = time.Now()
	}
	// spin slowly
	a.angle += float32(dt) * 1.5
}

func (a *SampleApp) OnRender(e *core.Engine, alpha float64) {
	mvp := rotZ(a.angle)
	e.Renderer.Draw(core.DrawCmd{
		Pipe:     a.pipe,
		Mesh:     a.mesh,
		Uniforms: map[string]any{"uMVP": mvp},
		Samplers: map[string]core.Texture{"uTex0": a.tex},
	})
	_ = alpha
}

func (a *SampleApp) OnEvent(e *core.Engine, ev core.Event) {
	// Resize is already handled centrally in core.Run, so we just log if desired
	switch v := ev.(type) {
	case core.EventKey:
		if v.Key == core.KeyEscape && v.Down {
			log.Println("ESC pressed — close the window to exit")
		}
	case core.EventResize:
		log.Printf("Resize: %dx%d", v.W, v.H)
	}
}

func main() {
	cfg := core.Config{
		Title: "Go Engine",
		Width: 1280, Height: 720,
		VSync:      true,
		ClearColor: [4]float32{0.08, 0.10, 0.12, 1},
	}
	app := &SampleApp{}

	newWindow := func(cfg core.Config) (core.Window, error) {
		// engine will attach event sink later
		return platform.NewGLFWWindow(cfg, nil)
	}
	newRenderer := func(win core.Window, cfg core.Config) (core.Renderer, error) {
		return glbackend.NewRendererGL(win, cfg)
	}

	if err := core.Run(app, cfg, newWindow, newRenderer); err != nil {
		panic(err)
	}
	time.Sleep(50 * time.Millisecond)
}

// Simple Z-rotation 4x4 column-major matrix suitable for GLSL
func rotZ(a float32) [16]float32 {
	c := float32(math.Cos(float64(a)))
	s := float32(math.Sin(float64(a)))
	return [16]float32{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}
