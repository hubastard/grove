package platform

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hubastard/grove/engine/core"
)

// GLFWWindow implements core.Window and pushes events to the app via a handler.
type GLFWWindow struct {
	w    *glfw.Window
	onEv func(core.Event)
}

// Must be called on main thread before any GL calls.
func NewGLFWWindow(cfg core.Config, onEvent func(core.Event)) (*GLFWWindow, error) {
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	// GL 3.2+ core profile (Mac requires forward-compatible flag).
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 0)

	win, err := glfw.CreateWindow(cfg.Width, cfg.Height, cfg.Title, nil, nil)
	if err != nil {
		return nil, err
	}
	win.MakeContextCurrent()
	if cfg.VSync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}

	if err := gl.Init(); err != nil {
		return nil, err
	}
	log.Printf("GL: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))

	gw := &GLFWWindow{w: win, onEv: onEvent}

	// Callbacks -> translate to core.Event
	win.SetCloseCallback(func(*glfw.Window) { gw.emit(core.EventCloseRequested{}) })
	win.SetFramebufferSizeCallback(func(_ *glfw.Window, w, h int) {
		gw.emit(core.EventResize{W: w, H: h})
	})
	win.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		gw.emit(core.EventMouseMove{X: x, Y: y})
	})
	win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		k := translateKey(key)
		if k == core.KeyUnknown {
			return
		}
		gw.emit(core.EventKey{Key: k, Down: action != glfw.Release, Mods: translateMods(mods)})
	})
	win.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		gw.emit(core.EventScroll{Xoff: xoff, Yoff: yoff})
	})

	return gw, nil
}

func (g *GLFWWindow) emit(ev core.Event) {
	if g.onEv != nil {
		g.onEv(ev)
	}
}

// core.Window impl
func (g *GLFWWindow) PollEvents()                          { glfw.PollEvents() }
func (g *GLFWWindow) SwapBuffers()                         { g.w.SwapBuffers() }
func (g *GLFWWindow) ShouldClose() bool                    { return g.w.ShouldClose() }
func (g *GLFWWindow) FramebufferSize() (int, int)          { return g.w.GetFramebufferSize() }
func (g *GLFWWindow) SetTitle(t string)                    { g.w.SetTitle(t) }
func (g *GLFWWindow) SetEventCallback(cb func(core.Event)) { g.onEv = cb }

func translateKey(k glfw.Key) core.Key {
	switch k {
	case glfw.KeyEscape:
		return core.KeyEscape
	case glfw.KeySpace:
		return core.KeySpace
	case glfw.KeyW:
		return core.KeyW
	case glfw.KeyA:
		return core.KeyA
	case glfw.KeyS:
		return core.KeyS
	case glfw.KeyD:
		return core.KeyD
	default:
		return core.KeyUnknown
	}
}

func translateMods(m glfw.ModifierKey) core.Mod {
	var out core.Mod
	if m&glfw.ModShift != 0 {
		out |= core.ModShift
	}
	if m&glfw.ModControl != 0 {
		out |= core.ModCtrl
	}
	if m&glfw.ModAlt != 0 {
		out |= core.ModAlt
	}
	if m&glfw.ModSuper != 0 {
		out |= core.ModSuper
	}
	return out
}
