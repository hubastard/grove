package main

import (
	"fmt"

	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/core"
	"github.com/hubastard/grove/engine/gfx/renderer2d"
	"github.com/hubastard/grove/engine/profiler"
	"github.com/hubastard/grove/engine/scene"
	"github.com/hubastard/grove/engine/text"
	"github.com/hubastard/grove/engine/ui"
)

// ------- A simple 2D Layer demo -------
type LayerDebug struct {
	cam           *scene.OrthoCamera2D
	r2d           *renderer2d.Renderer2D
	font          *text.Font
	stats         *renderer2d.Statistics
	frameDuration float32
	tick          int
}

func (l *LayerDebug) OnAttach(e *core.Engine) {
	// Camera sized to framebuffer
	w, h := e.Window.FramebufferSize()
	l.cam = scene.NewOrtho2D(w, h)
	l.cam.SetPosition(float32(w/2), float32(h/2)) // origin top-left
}

func (l *LayerDebug) OnDetach(e *core.Engine) {}

func (l *LayerDebug) OnUpdate(e *core.Engine, dt float64) {}

func (l *LayerDebug) OnRender(e *core.Engine, alpha float64) {
	l.r2d.BeginScene(l.cam.VP())
	{
		ui.View(
			ui.View(
				ui.Label(fmt.Sprintf("%2.3f ms (%.2f FPS)", l.frameDuration, 1000.0/l.frameDuration)),
				ui.Label(fmt.Sprintf("Tick: %d", l.tick)),
				ui.Label(fmt.Sprintf("Draw Calls: %d", l.stats.DrawCalls)),
				ui.Label(fmt.Sprintf("Quad Count: %d", l.stats.QuadCount)),
				ui.Label(fmt.Sprintf("Vertex Count: %d", l.stats.TotalVertexCount())),
			).
				FlowDirection(ui.LayoutVertical).
				Padding(24).
				BgColor(colors.Black.WithAlpha(0.5)),
		).
			Padding(16).
			Gap(12).
			FlowDirection(ui.LayoutVertical).
			AlignCross(ui.AlignStretch).
			Draw(&ui.Context{
				Viewport:    [4]float32{0, 0, l.cam.Width(), l.cam.Height()},
				DefaultFont: l.font,
				Renderer:    l.r2d,
			})
	}
	l.r2d.EndScene()
}

func (l *LayerDebug) OnEvent(e *core.Engine, ev core.Event) bool {
	switch v := ev.(type) {
	case core.EventKey:
		if v.Down && v.Key == core.KeyP && (v.Mods&core.ModCtrl) != 0 {
			if path, err := profiler.OpenProfilerGraph(); err == nil {
				fmt.Println("speedscope dump:", path)
			} else {
				fmt.Println("profiler dump error:", err)
			}
			return true
		}
	case core.EventResize:
		l.cam.SetViewportPixels(v.W, v.H)
		l.cam.SetPosition(float32(v.W/2), float32(v.H/2)) // origin top-left
	}
	return false
}
