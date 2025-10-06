package core

import "time"

// App defines the game/application hooks.
type App interface {
	OnStart(e *Engine)                 // called once after window/renderer init
	OnUpdate(e *Engine, dt float64)    // called at a fixed tick (60Hz by default)
	OnRender(e *Engine, alpha float64) // render with interpolation alpha [0..1]
	OnEvent(e *Engine, ev Event)       // input/window events
	OnShutdown(e *Engine)              // before exit
}

// Engine exposes core services to the App.
type Engine struct {
	Window   Window
	Renderer Renderer
	start    time.Time
}

func (e *Engine) Uptime() time.Duration { return time.Since(e.start) }

// Window abstraction.
type Window interface {
	PollEvents()
	SwapBuffers()
	ShouldClose() bool
	FramebufferSize() (int, int)
	SetTitle(title string)
}

// Renderer abstraction (minimal for now; grows with engine).
type Renderer interface {
	Init() error
	Resize(w, h int)
	Clear(r, g, b, a float32)
	DrawDemoTriangle() // temporary for the sandbox
	Shutdown()
}

// Event model (can expand over time).
type Event interface{ isEvent() }

type EventCloseRequested struct{}

func (EventCloseRequested) isEvent() {}

type EventResize struct{ W, H int }

func (EventResize) isEvent() {}

type EventKey struct {
	Key  Key
	Down bool
	Mods Mod
}

func (EventKey) isEvent() {}

type EventMouseMove struct{ X, Y float64 }

func (EventMouseMove) isEvent() {}

// Key/mod enums (subset; add as needed).
type Key int

const (
	KeyUnknown Key = iota
	KeyEscape
	KeySpace
	KeyW
	KeyA
	KeyS
	KeyD
)

type Mod int

const (
	ModNone  Mod = 0
	ModShift Mod = 1 << 0
	ModCtrl  Mod = 1 << 1
	ModAlt   Mod = 1 << 2
	ModSuper Mod = 1 << 3
)

// Config for the engine run.
type Config struct {
	Title      string
	Width      int
	Height     int
	VSync      bool
	ClearColor [4]float32 // RGBA
}
