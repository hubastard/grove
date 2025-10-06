package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hubastard/grove/engine/core"
	glbackend "github.com/hubastard/grove/engine/gfx/gl"
	"github.com/hubastard/grove/engine/platform"
)

type SampleApp struct {
	frames       int
	lastFPSTitle time.Time
}

func (a *SampleApp) OnStart(e *core.Engine) {
	log.Println("App start")
	a.lastFPSTitle = time.Now()
}

func (a *SampleApp) OnShutdown(e *core.Engine) {
	log.Println("App shutdown")
}

func (a *SampleApp) OnUpdate(e *core.Engine, dt float64) {
	// Count frames for FPS tracking
	a.frames++

	// Update title once per second
	if time.Since(a.lastFPSTitle) >= time.Second {
		// 1/dt is per-frame FPS, but since dt is fixed (1/60s) it’s ~60 FPS anyway.
		// To make it meaningful, you can estimate the FPS by frames / elapsed.
		elapsed := time.Since(a.lastFPSTitle).Seconds()
		fps := float64(a.frames) / elapsed

		e.Window.SetTitle(fmt.Sprintf("Go Engine — %.0f FPS", fps))

		a.frames = 0
		a.lastFPSTitle = time.Now()
	}
}

func (a *SampleApp) OnRender(e *core.Engine, alpha float64) {
	// Draw the demo triangle via renderer
	e.Renderer.DrawDemoTriangle()
	_ = alpha // use for interpolated rendering later
}

func (a *SampleApp) OnEvent(e *core.Engine, ev core.Event) {
	switch v := ev.(type) {
	case core.EventResize:
		log.Printf("Resize: %dx%d", v.W, v.H)
		e.Renderer.Resize(v.W, v.H)
	case core.EventKey:
		if v.Key == core.KeyEscape && v.Down {
			log.Println("ESC pressed — exiting soon")
			// GLFW will set ShouldClose() via close or you can signal it here with platform-specific call if desired
		}
	case core.EventMouseMove:
		_ = v
	case core.EventCloseRequested:
		// nothing to do; loop will end when ShouldClose() returns true
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
		// Wire events directly to the app
		return platform.NewGLFWWindow(cfg, func(ev core.Event) { app.OnEvent(nil, ev) })
	}
	newRenderer := func(win core.Window, cfg core.Config) (core.Renderer, error) {
		return glbackend.NewRendererGL(win, cfg)
	}

	if err := core.Run(app, cfg, newWindow, newRenderer); err != nil {
		log.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
}
