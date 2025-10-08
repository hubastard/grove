package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hubastard/grove/engine/assets"
	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/core"
	glbackend "github.com/hubastard/grove/engine/gfx/gl"
	"github.com/hubastard/grove/engine/gfx/renderer2d"
	"github.com/hubastard/grove/engine/platform"
	"github.com/hubastard/grove/engine/profiler"
	"github.com/hubastard/grove/engine/scene"
	"github.com/hubastard/grove/engine/text"
	"github.com/hubastard/grove/engine/ui"
)

type App struct {
	lastTitle time.Time
	frames    int
	layer     *Layer2D
}

func (a *App) OnStart(e *core.Engine) {
	profiler.Init(1 << 10) // ~1K scope samples

	// push the 2D demo layer
	l := &Layer2D{}
	e.Layers.Push(l)
	a.layer = l
}

func (a *App) OnUpdate(e *core.Engine, dt float64) {
	a.frames++
	if time.Since(a.lastTitle) >= time.Second {
		elapsed := time.Since(a.lastTitle).Seconds()
		fps := float64(a.frames) / elapsed
		if a.layer != nil {
			stats := a.layer.Stats()
			e.Window.SetTitle(fmt.Sprintf(
				"Go Engine — ~%.0f FPS | DC: %d | Quads: %d | Verts: %d | Inds: %d",
				fps,
				stats.DrawCalls,
				stats.QuadCount,
				stats.TotalVertexCount(),
				stats.TotalIndexCount(),
			))
		} else {
			e.Window.SetTitle(fmt.Sprintf("Go Engine — ~%.0f FPS", fps))
		}
		a.frames = 0
		a.lastTitle = time.Now()
	}
}
func (a *App) OnRender(e *core.Engine, alpha float64) {}
func (a *App) OnEvent(e *core.Engine, ev core.Event)  {}
func (a *App) OnShutdown(e *core.Engine)              {}

// ------- A simple 2D Layer demo -------
type Layer2D struct {
	worldCam *scene.OrthoCamera2D
	uiCam    *scene.OrthoCamera2D
	ctrl     *scene.OrthoController2D
	r2d      *renderer2d.Renderer2D
	tex      core.Texture
	font     *text.Font
	player   renderer2d.SubTexture2D
	t        float32
	stats    renderer2d.Statistics
}

func (l *Layer2D) OnAttach(e *core.Engine) {
	// Camera sized to framebuffer
	w, h := e.Window.FramebufferSize()
	l.worldCam = scene.NewOrtho2D(w, h)
	l.ctrl = scene.NewOrthoController2D(l.worldCam)

	l.uiCam = scene.NewOrtho2D(w, h)
	l.uiCam.SetPosition(float32(w/2), float32(h/2)) // origin top-left

	// Load 2D shader
	vs, err := assets.LoadShader("renderer2d.vert")
	if err != nil {
		panic(err)
	}
	fs, err := assets.LoadShader("renderer2d.frag")
	if err != nil {
		panic(err)
	}

	l.r2d, err = renderer2d.New(e.Renderer, vs, fs, 10000)
	if err != nil {
		panic(err)
	}

	l.font, err = text.LoadTTF(e.Renderer, "RobotoMono.ttf", 32)
	if err != nil {
		panic(err)
	}

	w, h, pixels, err := assets.LoadPNG("player.png")
	if err != nil {
		panic(err)
	}

	l.tex, err = e.Renderer.CreateTexture(core.TextureDesc{
		Width:     w,
		Height:    h,
		Format:    core.TextureRGBA8,
		Pixels:    pixels,
		MinFilter: "linear",
		MagFilter: "nearest",
		WrapU:     "clamp",
		WrapV:     "clamp",
	})
	if err != nil {
		panic(err)
	}

	l.player = renderer2d.FromPixels(l.tex, 0, 0, 32, 32, w, h)
}

func (l *Layer2D) OnDetach(e *core.Engine) {}

func (l *Layer2D) OnUpdate(e *core.Engine, dt float64) {
	l.ctrl.Update(e, float32(dt))
	l.t += float32(dt)

	if e.Input.IsKeyDown(core.KeyEscape) {
		e.Window.RequestClose()
	}
}

func (l *Layer2D) OnRender(e *core.Engine, alpha float64) {
	l.r2d.BeginScene(l.worldCam.VP())

	l.r2d.DrawSubTexQuad(0, 0, 32, 32, l.player, colors.White, l.t)

	// stats := l.Stats()
	// text.DrawText(l.r2d, l.font, -500, -500, fmt.Sprintf("Draw Calls: %d", stats.DrawCalls), l.white)
	l.r2d.EndScene()

	l.r2d.BeginScene(l.uiCam.VP())

	uictx := ui.Context{
		Viewport:    [4]float32{0, 0, l.uiCam.Width(), l.uiCam.Height()},
		DefaultFont: l.font,
		Renderer:    l.r2d,
	}

	ui.Canvas(
		ui.Label("Welcome to Grove UI").FontSize(32).Padding(8),
		ui.Label("This layout engine supports fit, expand, and wrapped text. Resize the viewport or tweak the sizing modes to experiment.").Wrap(true).WidthExpand().Padding(4),
		ui.Button("Primary Action").FontSize(24).BgColor(colors.Color{0.2, 0.2, 0.8, 1}).WidthExpand(),
		ui.Button("Secondary").FontSize(16).BgColor(colors.Color{0.15, 0.15, 0.4, 1}).WidthExpand(),
	).
		Padding(16).
		Gap(12).
		LayoutVertically(true).
		AlignCross(ui.AlignStretch).
		Draw(&uictx)
	l.r2d.EndScene()

	l.stats = l.r2d.Stats()
}

func (l *Layer2D) OnEvent(e *core.Engine, ev core.Event) bool {
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
		l.worldCam.SetViewportPixels(v.W, v.H)
		l.uiCam.SetViewportPixels(v.W, v.H)
		l.uiCam.SetPosition(float32(v.W/2), float32(v.H/2)) // origin top-left
	case core.EventScroll:
		if l.ctrl.HandleEvent(e, ev) {
			return true
		}
	}
	return false
}

func (l *Layer2D) Stats() renderer2d.Statistics { return l.stats }

func main() {
	cfg := core.Config{
		Title:      "Go Engine (2D)",
		Width:      1280,
		Height:     720,
		VSync:      true,
		ClearColor: colors.DarkGray,
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
