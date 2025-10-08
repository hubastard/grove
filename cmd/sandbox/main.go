package main

import (
	"log"
	"time"

	"github.com/hubastard/grove/engine/assets"
	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/core"
	glbackend "github.com/hubastard/grove/engine/gfx/gl"
	"github.com/hubastard/grove/engine/gfx/renderer2d"
	"github.com/hubastard/grove/engine/platform"
	"github.com/hubastard/grove/engine/profiler"
	"github.com/hubastard/grove/engine/text"
)

type App struct {
	lastFrame  time.Time
	tick       int
	r2d        *renderer2d.Renderer2D
	stats      renderer2d.Statistics
	font       *text.Font
	layer      *Layer2D
	debugLayer *LayerDebug
}

func (a *App) OnStart(e *core.Engine) {
	profiler.Init(1 << 10) // ~1K scope samples

	// Load 2D shader
	vs, err := assets.LoadShader("renderer2d.vert")
	if err != nil {
		panic(err)
	}
	fs, err := assets.LoadShader("renderer2d.frag")
	if err != nil {
		panic(err)
	}

	a.r2d, err = renderer2d.New(e.Renderer, vs, fs, 10000)
	if err != nil {
		panic(err)
	}

	// Load default font
	a.font, err = text.LoadTTF(e.Renderer, "RobotoMono.ttf", 32)
	if err != nil {
		panic(err)
	}

	// push the 2D demo layer
	a.layer = &Layer2D{r2d: a.r2d}
	e.Layers.Push(a.layer)

	a.debugLayer = &LayerDebug{r2d: a.r2d, font: a.font, stats: &a.stats}
	e.Layers.Push(a.debugLayer)
}

func (a *App) OnUpdate(e *core.Engine, dt float64) {
	a.tick++

	// Calculate frame duration
	now := time.Now()
	if a.debugLayer != nil && !a.lastFrame.IsZero() {
		a.debugLayer.frameDuration = float32(now.Sub(a.lastFrame).Seconds() * 1000.0)
		a.debugLayer.tick = a.tick
	}
	a.lastFrame = now
}
func (a *App) OnRender(e *core.Engine, alpha float64) {
	a.stats = a.r2d.Stats()
}
func (a *App) OnEvent(e *core.Engine, ev core.Event) {}
func (a *App) OnShutdown(e *core.Engine)             {}

func main() {
	cfg := core.Config{
		Title:                "Go Engine (2D)",
		Width:                1280,
		Height:               720,
		VSync:                true,
		ClearColor:           colors.DarkGray,
		ScratchAllocCapacity: 4096, // 4 KB initial capacity
		ScratchEnableLogs:    true,
	}
	app := &App{}

	newWindow := func(cfg core.Config) (core.Window, error) {
		return platform.NewGLFWWindow(cfg, nil)
	}
	newRenderer := func(win core.Window, cfg core.Config) (core.Renderer, error) {
		return glbackend.NewRendererGL(win, cfg)
	}

	if err := core.Run(app, cfg, newWindow, newRenderer); err != nil {
		log.Fatal(err)
	}
}
