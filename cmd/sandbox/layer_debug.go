package main

import (
	"log"

	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/core"
	"github.com/hubastard/grove/engine/gfx/renderer2d"
	"github.com/hubastard/grove/engine/profiler"
	"github.com/hubastard/grove/engine/scene"
	"github.com/hubastard/grove/engine/scratch"
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
	ctx           *ui.Ctx
}

type UIRenderer struct {
	r2d  *renderer2d.Renderer2D
	font *text.Font
}

// Draws a solid quad centered at (cx, cy) with w,h and color RGBA [0..1]
func (u *UIRenderer) DrawQuad(cx, cy, w, h float32, color [4]float32, rotation float32) {
	u.r2d.DrawQuad(cx, cy, w, h, color, rotation)
}
func (u *UIRenderer) DrawText(x, y float32, str string, size float32, color [4]float32) {
	text.DrawText(u.r2d, u.font, x, y, str, color)
}
func (u *UIRenderer) Measure(str string, size float32) (w, h float32) {
	// TODO: support size scaling
	return text.MeasureText(u.font, str)
}

func (l *LayerDebug) OnAttach(e *core.Engine) {
	// Camera sized to framebuffer
	w, h := e.Window.FramebufferSize()
	l.cam = scene.NewOrtho2D(w, h)
	l.cam.SetPosition(float32(w/2), float32(h/2)) // origin top-left

	l.ctx = ui.New(64, 512, 512)
	l.ctx.R = &UIRenderer{r2d: l.r2d, font: l.font}
	l.ctx.I = &ui.Input{}
}

func (l *LayerDebug) OnDetach(e *core.Engine) {}

func (l *LayerDebug) OnUpdate(e *core.Engine, dt float64) {
	l.ctx.I.MouseX, l.ctx.I.MouseY = e.Input.MousePosition()
	l.ctx.I.MouseDown = e.Input.IsMouseDown(core.MouseButtonLeft)
	l.ctx.I.MousePressed = e.Input.IsMousePressed(core.MouseButtonLeft)
	l.ctx.I.MouseReleased = e.Input.IsMouseReleased(core.MouseButtonLeft)

	// Update UI
	ui.Use(l.ctx)
	ui.BeginFrame(l.ctx)

	ui.BeginView(ui.Props{
		Axis:      ui.Vertical,
		MainAlign: ui.End,
		Sizing:    ui.Fit(),
		Padding:   ui.Insets(8, 8, 8, 8),
		Gap:       8,
		ID:        1,
	})

	ui.Label(ui.LabelProps{Text: scratch.Sprintf("Frame: %d", l.tick), Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\t%.3f ms (%.2f FPS)", l.frameDuration, 1000.0/l.frameDuration)})
	ui.Label(ui.LabelProps{Text: "2D Renderer", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tDraw Calls: %d", l.stats.DrawCalls)})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tQuads: %d", l.stats.QuadCount)})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tVertices: %d", l.stats.TotalVertexCount())})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tTextures: %d", l.stats.TextureCount)})
	ui.Label(ui.LabelProps{Text: "Memory", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tUsage: %.3f MB", float32(profiler.MemoryUsage())/(1<<20))})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tAllocs: %d", profiler.MemoryAllocs())})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tGoroutines: %d", profiler.NumGoroutine())})
	ui.Label(ui.LabelProps{Text: "CPU", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tCount: %d", profiler.NumCPU())})
	ui.Label(ui.LabelProps{Text: "GPU", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tVendor: %s", e.Renderer.GPUVendor())})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tRenderer: %s", e.Renderer.GPURenderer())})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tVersion: %s", e.Renderer.GPUVersion())})

	ui.EndView()
}

func (l *LayerDebug) OnRender(e *core.Engine, alpha float64) {
	scopeRender := profiler.Start("LayerDebug.OnRender")

	l.r2d.BeginScene(l.cam.VP())
	ui.Flush(l.ctx)
	l.r2d.EndScene()

	scopeRender.End()
}

func (l *LayerDebug) OnEvent(e *core.Engine, ev core.Event) bool {
	switch v := ev.(type) {
	case core.EventKey:
		if v.Down && v.Key == core.KeyP && (v.Mods&core.ModCtrl) != 0 {
			if path, err := profiler.OpenProfilerGraph(); err == nil {
				log.Printf("speedscope dump: %s\n", path)
			} else {
				log.Printf("profiler dump error: %v\n", err)
			}
			return true
		}
	case core.EventResize:
		l.cam.SetViewportPixels(v.W, v.H)
		l.cam.SetPosition(float32(v.W/2), float32(v.H/2)) // origin top-left
	}
	return false
}
