package ui

// ===== Sizing & layout props =====

type Axis int

const (
	Horizontal Axis = iota
	Vertical
)

type Align int

const (
	Start Align = iota
	Center
	End
	Stretch
)

type SizeMode int

const (
	SizeFit SizeMode = iota
	SizeFixed
	SizeExpand
)

type Sizing struct {
	WMode SizeMode
	HMode SizeMode
	WVal  float32 // for SizeFixed
	HVal  float32 // for SizeFixed
}

func Fit() Sizing            { return Sizing{WMode: SizeFit, HMode: SizeFit} }
func Expand() Sizing         { return Sizing{WMode: SizeExpand, HMode: SizeExpand} }
func Px(w, h float32) Sizing { return Sizing{WMode: SizeFixed, HMode: SizeFixed, WVal: w, HVal: h} }

type Insets4 struct{ L, T, R, B float32 }

func Insets(l, t, r, b float32) Insets4 { return Insets4{l, t, r, b} }

// Container props
type Props struct {
	ID         int // stable id for this view
	Axis       Axis
	MainAlign  Align
	CrossAlign Align
	Sizing     Sizing
	Gap        float32
	Padding    Insets4
	Bg         [4]float32 // optional background
	// Optional fixed size override (if Sizing is Px)
	// Otherwise used as constraints box when Expand
	BoundsX float32
	BoundsY float32
	BoundsW float32
	BoundsH float32
}

// ===== Internal structs =====

type viewScope struct {
	props Props
	// Children recorded during measure phase
	firstCmd int // index in ctx.cmds where children commands begin
	nCmds    int
	bgCmd    int // optional background command index

	// Measured children
	firstItem int // index in ctx.items
	nItems    int

	// Resolved rect for this view (content box: minus padding)
	x, y, w, h float32
}

type item struct {
	kind cmdKind
	iCmd int     // index into ctx.cmds
	w, h float32 // desired size
}

type cmdKind uint8

const (
	cmdLabel cmdKind = iota
	cmdButton
	cmdBgQuad
)

type cmd struct {
	kind cmdKind
	id   int

	// geom (resolved at EndView)
	x, y, w, h float32

	// visuals
	text     string
	fontSize float32
	color    [4]float32
	bg       [4]float32

	// button interaction bookkeeping
	hot, active bool
	clicked     bool
}

type widgetState struct {
	hot    bool
	active bool
}

// ===== Begin/End view =====

func BeginView(p Props) {
	ctx := current // global/thread-local; set by your engine before UI pass
	// push scope
	if len(ctx.viewStack) == cap(ctx.viewStack) {
		// hard cap to keep zero-alloc invariant. In practice: bump capacities once.
		return
	}
	scope := viewScope{
		props:     p,
		firstCmd:  len(ctx.cmds),
		firstItem: len(ctx.items),
		bgCmd:     -1,
	}
	if p.Bg[3] > 0 {
		idx := emit(ctx, cmd{
			kind: cmdBgQuad,
			bg:   p.Bg,
		})
		if idx >= 0 {
			scope.bgCmd = idx
		}
	}
	ctx.viewStack = append(ctx.viewStack, scope)
}

func EndView() {
	ctx := current
	if len(ctx.viewStack) == 0 {
		return
	}

	// pop scope
	scope := ctx.viewStack[len(ctx.viewStack)-1]
	ctx.viewStack = ctx.viewStack[:len(ctx.viewStack)-1]
	scope.nCmds = len(ctx.cmds) - scope.firstCmd
	scope.nItems = len(ctx.items) - scope.firstItem

	// measure total main/cross span
	var totalMain, maxCross float32
	gap := scope.props.Gap
	mainIsX := (scope.props.Axis == Horizontal)

	for i := 0; i < scope.nItems; i++ {
		it := ctx.items[scope.firstItem+i]
		if mainIsX {
			totalMain += it.w
			if it.h > maxCross {
				maxCross = it.h
			}
		} else {
			totalMain += it.h
			if it.w > maxCross {
				maxCross = it.w
			}
		}
	}
	if scope.nItems > 1 {
		totalMain += gap * float32(scope.nItems-1)
	}

	// resolve self size
	var availW, availH float32
	switch scope.props.Sizing.WMode {
	case SizeFixed:
		availW = scope.props.Sizing.WVal
	case SizeExpand:
		availW = scope.props.BoundsW
	default: // fit
		if mainIsX {
			availW = totalMain
		} else {
			availW = maxCross
		}
	}
	switch scope.props.Sizing.HMode {
	case SizeFixed:
		availH = scope.props.Sizing.HVal
	case SizeExpand:
		availH = scope.props.BoundsH
	default: // fit
		if mainIsX {
			availH = maxCross
		} else {
			availH = totalMain
		}
	}
	// add paddings
	availW += scope.props.Padding.L + scope.props.Padding.R
	availH += scope.props.Padding.T + scope.props.Padding.B

	// resolve own position (use provided BoundsX/Y; parent will place us)
	scope.x = scope.props.BoundsX + scope.props.Padding.L
	scope.y = scope.props.BoundsY + scope.props.Padding.T
	scope.w = maxf(0, availW-scope.props.Padding.L-scope.props.Padding.R)
	scope.h = maxf(0, availH-scope.props.Padding.T-scope.props.Padding.B)

	// optional background
	if scope.bgCmd >= 0 {
		c := &ctx.cmds[scope.bgCmd]
		c.x = scope.props.BoundsX
		c.y = scope.props.BoundsY
		c.w = availW
		c.h = availH
		c.bg = scope.props.Bg
	}

	// compute starting offset for main align
	var start float32
	free := func() float32 {
		if mainIsX {
			return scope.w - totalMain
		}
		return scope.h - totalMain
	}()
	if free < 0 {
		free = 0
	}
	switch scope.props.MainAlign {
	case Start:
		start = 0
	case Center:
		start = free * 0.5
	case End:
		start = free
	default:
		start = 0
	}

	// place children + finalize draw / hit-test
	cursor := start
	for i := 0; i < scope.nItems; i++ {
		it := ctx.items[scope.firstItem+i]
		c := &ctx.cmds[it.iCmd]

		// cross alignment
		var crossPos float32
		switch scope.props.CrossAlign {
		case Start:
			crossPos = 0
		case Center:
			if mainIsX {
				crossPos = (scope.h - it.h) * 0.5
			} else {
				crossPos = (scope.w - it.w) * 0.5
			}
		case End:
			if mainIsX {
				crossPos = scope.h - it.h
			} else {
				crossPos = scope.w - it.w
			}
		case Stretch:
			// override item size on cross axis
			if mainIsX {
				it.h = scope.h
			} else {
				it.w = scope.w
			}
			crossPos = 0
		}

		// final xywh
		if mainIsX {
			c.x, c.y, c.w, c.h = scope.x+cursor, scope.y+crossPos, it.w, it.h
			cursor += it.w
		} else {
			c.x, c.y, c.w, c.h = scope.x+crossPos, scope.y+cursor, it.w, it.h
			cursor += it.h
		}
		if i != scope.nItems-1 {
			cursor += gap
		}
	}

	// clear transient items segment (no allocs, just shrink)
	ctx.items = ctx.items[:scope.firstItem]
}

// Helper to seed a view’s child list with 0 allocs.
func addItem(ctx *Ctx, it item) {
	// Ensure capacity: we hard cap at initialization to avoid allocs;
	// bump capItems in New(...) once if you hit the ceiling.
	if len(ctx.items) == cap(ctx.items) {
		return
	}
	ctx.items = append(ctx.items, it)
}

// For Begin/End helpers, we need a pointer to the active Ctx.
// In your engine, set this before the UI pass and DO NOT change during a frame.
var current *Ctx

func Use(ctx *Ctx) { current = ctx }

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
