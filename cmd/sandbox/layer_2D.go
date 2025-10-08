package main

import (
	"fmt"

	"github.com/hubastard/grove/engine/assets"
	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/core"
	"github.com/hubastard/grove/engine/gfx/renderer2d"
	"github.com/hubastard/grove/engine/profiler"
	"github.com/hubastard/grove/engine/scene"
)

// ------- A simple 2D Layer demo -------
type Layer2D struct {
	cam    *scene.OrthoCamera2D
	ctrl   *scene.OrthoController2D
	r2d    *renderer2d.Renderer2D
	tex    core.Texture
	player renderer2d.SubTexture2D
	t      float32
}

func (l *Layer2D) OnAttach(e *core.Engine) {
	// Camera sized to framebuffer
	w, h := e.Window.FramebufferSize()
	l.cam = scene.NewOrtho2D(w, h)
	l.cam.SetZoom(4)
	l.ctrl = scene.NewOrthoController2D(l.cam)

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
	renderEnd := profiler.Start("Layer2D.OnRender")

	l.r2d.BeginScene(l.cam.VP())
	{
		l.r2d.DrawSubTexQuad(0, 0, 32, 32, l.player, colors.White, l.t)
	}
	l.r2d.EndScene()

	renderEnd()
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
		l.cam.SetViewportPixels(v.W, v.H)
	case core.EventScroll:
		if l.ctrl.HandleEvent(e, ev) {
			return true
		}
	}
	return false
}
