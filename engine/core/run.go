package core

import (
	"log"
	"runtime"
	"time"

	"github.com/hubastard/grove/engine/profiler"
)

// Run wires the platform window + renderer and executes the main loop.
func Run(app App, cfg Config, newWindow func(Config) (Window, error), newRenderer func(Window, Config) (Renderer, error)) error {
	// Graphics contexts require the main OS thread.
	runtime.LockOSThread()

	win, err := newWindow(cfg)
	if err != nil {
		return err
	}

	rend, err := newRenderer(win, cfg)
	if err != nil {
		return err
	}
	defer rend.Shutdown()

	eng := &Engine{Window: win, Renderer: rend, Input: NewInput(), start: time.Now()}

	// authoritative initial size
	w, h := win.FramebufferSize()
	rend.Resize(w, h)

	win.SetEventCallback(func(ev Event) {
		eng.Input.Handle(ev)
		eng.Layers.ForEachReverse(func(l Layer) bool { return l.OnEvent(eng, ev) })
		app.OnEvent(eng, ev)

		// engine-level resize
		if _, ok := ev.(EventResize); ok {
			fw, fh := win.FramebufferSize()
			if fw < 1 || fh < 1 {
				return
			}
			rend.Resize(fw, fh)
		}
	})

	app.OnStart(eng)
	eng.Layers.ForEach(func(l Layer) { l.OnAttach(eng) })

	// Fixed-timestep (default 60 Hz) with interpolation
	tps := time.Duration(60)
	if cfg.TickPerSec > 0 {
		tps = time.Duration(cfg.TickPerSec)
	}

	tick := time.Second / tps
	var (
		accum   time.Duration
		prev    = time.Now()
		clear   = cfg.ClearColor
		maxStep = 10 // prevent spiral of death
	)

	for !win.ShouldClose() {
		pFrameEnd := profiler.Start("Frame")

		now := time.Now()
		frame := now.Sub(prev)
		prev = now
		accum += frame

		// Poll OS events (platform will emit via callbacks)
		pPollEvents := profiler.Start("PollEvents")
		win.PollEvents()
		pPollEvents()

		// Run fixed updates
		steps := 0
		for accum >= tick && steps < maxStep {
			pUpdateEnd := profiler.Start("Update")
			app.OnUpdate(eng, float64(tick)/float64(time.Second))
			eng.Layers.ForEach(func(l Layer) { l.OnUpdate(eng, float64(tick)/float64(time.Second)) })
			accum -= tick
			steps++
			pUpdateEnd()
		}

		// Interpolation factor for rendering
		alpha := float64(accum) / float64(tick)

		// Render
		pRenderEnd := profiler.Start("Render")
		rend.Clear(clear[0], clear[1], clear[2], clear[3])
		app.OnRender(eng, alpha)
		eng.Layers.ForEach(func(l Layer) { l.OnRender(eng, alpha) })
		pRenderEnd()

		// Frame end (we don't include SwapBuffers in profiling)

		// Present
		pSwapEnd := profiler.Start("SwapBuffers")
		win.SwapBuffers()
		pSwapEnd()

		pFrameEnd()
	}

	eng.Layers.ForEach(func(l Layer) { l.OnDetach(eng) })
	app.OnShutdown(eng)
	log.Println("Engine exit")
	return nil
}
