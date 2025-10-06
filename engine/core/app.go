package core

import "time"

// App defines the game/application hooks.
type App interface {
	OnStart(e *Engine)                 // once after init
	OnUpdate(e *Engine, dt float64)    // fixed tick (60Hz)
	OnRender(e *Engine, alpha float64) // render with interpolation
	OnEvent(e *Engine, ev Event)       // input/window events
	OnShutdown(e *Engine)              // before exit
}

type Engine struct {
	Window   Window
	Renderer Renderer
	Input    *Input
	Layers   LayerStack
	start    time.Time
}

func (e *Engine) Uptime() time.Duration { return time.Since(e.start) }

// Window abstraction.
type Window interface {
	PollEvents()
	SwapBuffers()
	ShouldClose() bool
	RequestClose()
	FramebufferSize() (int, int)
	SetTitle(title string)
	SetEventCallback(cb func(Event))
}

// -------- Renderer abstraction (generic, no GL types) --------

type AttribType int

const (
	AttribFloat32 AttribType = iota
)

type VertexAttrib struct {
	Location int // layout(location = X)
	Size     int // number of components (e.g., 2 for vec2, 3 for vec3)
	Type     AttribType
	Offset   int // byte offset
}

type VertexLayout struct {
	Stride     int // bytes per vertex
	Attributes []VertexAttrib
}

type MeshDesc struct {
	Vertices []float32
	Indices  []uint32 // optional; empty => draw arrays
	Layout   VertexLayout
}

type PipelineDesc struct {
	VertexSource   string // GLSL
	FragmentSource string // GLSL
	DepthTest      bool
	Blend          bool
}

type TextureFormat int

const (
	TextureRGBA8 TextureFormat = iota
)

type TextureDesc struct {
	Width, Height int
	Format        TextureFormat
	Pixels        []byte // expected to match Format (e.g., RGBA8 = 4 * w * h)
	MinFilter     string // "nearest" | "linear"
	MagFilter     string // "nearest" | "linear"
	WrapU         string // "clamp" | "repeat"
	WrapV         string // "clamp" | "repeat"
}

type Mesh interface{ IsMesh() }
type Pipeline interface{ IsPipeline() }
type Texture interface{ IsTexture() }

// Uniforms: simple map; for textures use Samplers keyed by uniform name
type DrawCmd struct {
	Pipe     Pipeline
	Mesh     Mesh
	Count    int                // vertex count if no indices; else ignored
	Uniforms map[string]any     // e.g. "uMVP": [16]float32
	Samplers map[string]Texture // e.g. "uTex0": Texture
}

type Renderer interface {
	Init() error
	Resize(w, h int)
	Clear(r, g, b, a float32)
	CreateMesh(desc MeshDesc) (Mesh, error)
	CreatePipeline(desc PipelineDesc) (Pipeline, error)
	CreateTexture(desc TextureDesc) (Texture, error)
	Draw(cmd DrawCmd)
	Shutdown()
}

// -------- Events --------

type Event interface{ isEvent() }

type EventCloseRequested struct{}

func (EventCloseRequested) isEvent() {}

type EventResize struct{ W, H int }

func (EventResize) isEvent() {}

type EventScroll struct{ Xoff, Yoff float64 }

func (EventScroll) isEvent() {}

type EventKey struct {
	Key  Key
	Down bool
	Mods Mod
}

func (EventKey) isEvent() {}

type EventMouseButton struct {
	Button MouseButton
	Down   bool
}

func (EventMouseButton) isEvent() {}

type EventMouseMove struct{ X, Y float64 }

func (EventMouseMove) isEvent() {}

// Config for the engine run.
type Config struct {
	Title      string
	Width      int
	Height     int
	VSync      bool
	ClearColor [4]float32 // RGBA
}
