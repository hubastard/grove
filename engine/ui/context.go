package ui

// ===== Public engine-facing bits you likely already have =====

type Renderer interface {
	// Draws a solid quad centered at (cx, cy) with w,h and color RGBA [0..1]
	DrawQuad(cx, cy, w, h float32, color [4]float32, rotation float32)
	// Draws text top-left at (x,y)
	DrawText(x, y float32, text string, size float32, color [4]float32)
	// Measures text (w,h) for a given font size
	Measure(text string, size float32) (w, h float32)
}

type Input struct {
	MouseX, MouseY float32
	MouseDown      bool
	MousePressed   bool
	MouseReleased  bool
}

// ===== Immediate-UI context =====

type Ctx struct {
	R Renderer
	I *Input

	// Fixed-capacity stacks & buffers reused every frame
	viewStack []viewScope // layout scopes
	cmds      []cmd       // drawing + hit-test commands (deferred)
	items     []item      // transient per-view child list (reused)

	// Stable widget state (hot/active) — no per-frame inserts after bootstrap
	state map[int]widgetState

	// Limits to avoid re-alloc; tweak to your game’s needs once
	capViews int
	capCmds  int
	capItems int
}

func New(capViews, capCmds, capItems int) *Ctx {
	return &Ctx{
		viewStack: make([]viewScope, 0, capViews),
		cmds:      make([]cmd, 0, capCmds),
		items:     make([]item, 0, capItems),
		state:     make(map[int]widgetState, 256), // fills once, then steady
		capViews:  capViews,
		capCmds:   capCmds,
		capItems:  capItems,
	}
}

// Reset for a new frame. No heap allocations.
func BeginFrame(ctx *Ctx) {
	ctx.cmds = ctx.cmds[:0]
	ctx.viewStack = ctx.viewStack[:0]
	// items is a transient scratch reused per view — we clear per EndView
}

// Optionally split render from layout if your engine needs that.
// Here we already draw during resolve; this can be a no-op.
func Flush(ctx *Ctx) {
	for i := range ctx.cmds {
		resolveWidget(ctx, &ctx.cmds[i])
	}
}
