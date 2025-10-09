package main

import (
	"fmt"
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
	lastAllocs    uint64
	allocs        uint64
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
		Padding:   ui.Insets(24, 24, 24, 24),
		Gap:       8,
		Bg:        colors.Black.WithAlpha(0.5),
	})

	ui.Label(ui.LabelProps{Text: scratch.Sprintf("Frame: %d", l.tick), Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\t%.3f ms (%.2f FPS)", l.frameDuration, 1000.0/l.frameDuration)})
	ui.Label(ui.LabelProps{Text: "2D Renderer", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tDraw Calls: %d\n\tQuads: %d\n\tVertices: %d\n\tTextures: %d", l.stats.DrawCalls, l.stats.QuadCount, l.stats.TotalVertexCount(), l.stats.TextureCount)})
	ui.Label(ui.LabelProps{Text: "Memory", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tUsage: %.3f MB\n\tTotal Allocs: %d\n\tFrame Allocs: %d\n\tGoroutines: %d", float32(profiler.MemoryUsage())/(1<<20), l.allocs, l.allocs-l.lastAllocs, profiler.NumGoroutine())})
	ui.Label(ui.LabelProps{Text: "Hardware", Color: colors.Yellow})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tCPU Cores: %d", profiler.NumCPU())})
	ui.Label(ui.LabelProps{Text: scratch.Sprintf("\tGPU: %s - v%s", e.Renderer.GPURenderer(), e.Renderer.GPUVersion())})

	if ui.Button(ui.ButtonProps{ID: 2, Text: "Click Me!", Padding: ui.Insets(16, 8, 16, 8), Bg: colors.Blue}) {
		fmt.Println("Button clicked!")
	}

	ui.EndView()
}

func (l *LayerDebug) OnRender(e *core.Engine, alpha float64) {
	scopeRender := profiler.Start("LayerDebug.OnRender")

	l.r2d.BeginScene(l.cam.VP())
	ui.Flush(l.ctx)
	l.r2d.EndScene()

	scopeRender.End()

	l.lastAllocs = l.allocs
	l.allocs = profiler.MemoryAllocs()
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
