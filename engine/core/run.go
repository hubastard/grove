package core

import (
	"log"
	"runtime"
	"time"
)

// Run wires the platform window + renderer and executes the main loop.
func Run(app App, cfg Config, newWindow func(Config) (Window, error), newRenderer func(Window, Config) (Renderer, error)) error {
	// Graphics contexts require the main OS thread.
	runtime.LockOSThread()

	win, err := newWindow(cfg)
	if err != nil {
		return err
	}
	defer func() {
		// window owns context; renderer shuts down first (below)
	}()

	rend, err := newRenderer(win, cfg)
	if err != nil {
		return err
	}
	defer rend.Shutdown()

	w, h := win.FramebufferSize()
	rend.Resize(w, h)

	eng := &Engine{Window: win, Renderer: rend, start: time.Now()}
	win.SetEventCallback(func(ev Event) {
		app.OnEvent(eng, ev)
		if _, ok := ev.(EventResize); ok {
			fw, fh := win.FramebufferSize()
			if fw < 1 || fh < 1 {
				return
			}
			rend.Resize(fw, fh)
		}
	})

	app.OnStart(eng)

	// Fixed-timestep (60 Hz) with interpolation
	const tick = time.Second / 60
	var (
		accum   time.Duration
		prev    = time.Now()
		clear   = cfg.ClearColor
		maxStep = 10 // prevent spiral of death
	)

	for !win.ShouldClose() {
		now := time.Now()
		frame := now.Sub(prev)
		prev = now
		accum += frame

		// Poll OS events (platform will emit via callbacks)
		win.PollEvents()

		// Run fixed updates
		steps := 0
		for accum >= tick && steps < maxStep {
			app.OnUpdate(eng, float64(tick)/float64(time.Second))
			accum -= tick
			steps++
		}
		// Interpolation factor for rendering
		alpha := float64(accum) / float64(tick)

		// Render
		rend.Clear(clear[0], clear[1], clear[2], clear[3])
		app.OnRender(eng, alpha)

		// Present
		win.SwapBuffers()
	}

	app.OnShutdown(eng)
	log.Println("Engine exit")
	return nil
}
