package ui

import "github.com/hubastard/grove/engine/colors"

// ===== Label =====

type LabelProps struct {
	ID       int
	Text     string
	FontSize float32
	Color    [4]float32
	Sizing   Sizing
}

func Label(p LabelProps) {
	ctx := current
	// Measure during "record" phase
	w, h := ctx.R.Measure(p.Text, p.FontSize)

	sz := p.Sizing
	switch sz.WMode {
	case SizeFixed:
		w = sz.WVal
	case SizeExpand: /* fill at resolve via Stretch or container size */
	}
	switch sz.HMode {
	case SizeFixed:
		h = sz.HVal
	case SizeExpand: /* same as above */
	}

	if p.Color[0] == 0 && p.Color[1] == 0 && p.Color[2] == 0 && p.Color[3] == 0 {
		p.Color = colors.White
	}

	iCmd := emit(ctx, cmd{
		kind:     cmdLabel,
		id:       p.ID,
		text:     p.Text,
		fontSize: p.FontSize,
		color:    p.Color,
	})

	addItem(ctx, item{kind: cmdLabel, iCmd: iCmd, w: w, h: h})
}

// ===== Button =====

type ButtonProps struct {
	ID       int
	Text     string
	FontSize float32
	TextCol  [4]float32
	Bg       [4]float32
	Padding  Insets4
	// Sizing modes: Fit (by default), Px, or Expand
	Sizing *Sizing
}

func Button(p ButtonProps) (clicked bool) {
	ctx := current
	tw, th := ctx.R.Measure(p.Text, p.FontSize)
	w := tw + p.Padding.L + p.Padding.R
	h := th + p.Padding.T + p.Padding.B

	sz := Fit()
	if p.Sizing != nil {
		sz = *p.Sizing
	}
	if sz.WMode == SizeFixed {
		w = sz.WVal
	}
	if sz.HMode == SizeFixed {
		h = sz.HVal
	}

	if p.TextCol[0] == 0 && p.TextCol[1] == 0 && p.TextCol[2] == 0 && p.TextCol[3] == 0 {
		p.TextCol = colors.White
	}

	iCmd := emit(ctx, cmd{
		kind:     cmdButton,
		id:       p.ID,
		text:     p.Text,
		fontSize: p.FontSize,
		color:    p.TextCol,
		bg:       p.Bg,
	})

	addItem(ctx, item{kind: cmdButton, iCmd: iCmd, w: w, h: h})

	// Click result is returned after resolve; in immediate mode we can peek
	// the last command’s .clicked only after EndView. For ergonomic usage in
	// a single pass, we return the last-known click state from the map,
	// which we set during resolve. This works frame-to-frame.
	st := ctx.state[p.ID]
	return !st.active && !st.hot && ctx.I.MouseReleased && pointInCmd(ctx, &ctx.cmds[iCmd], ctx.I.MouseX, ctx.I.MouseY)
}

// ===== Internal: record & resolve =====

func emit(ctx *Ctx, c cmd) int {
	if len(ctx.cmds) == cap(ctx.cmds) {
		return -1
	}
	ctx.cmds = append(ctx.cmds, c)
	return len(ctx.cmds) - 1
}

func resolveWidget(ctx *Ctx, c *cmd) {
	switch c.kind {
	case cmdBgQuad:
		drawQuad(ctx, c)
	case cmdLabel:
		drawLabel(ctx, c)
	case cmdButton:
		resolveButton(ctx, c)
	}
}

func drawQuad(ctx *Ctx, c *cmd) {
	cx := c.x + c.w*0.5
	cy := c.y + c.h*0.5
	ctx.R.DrawQuad(cx, cy, c.w, c.h, c.bg, 0)
}

func drawLabel(ctx *Ctx, c *cmd) {
	// left baseline
	ctx.R.DrawText(c.x, c.y, c.text, c.fontSize, c.color)
}

func resolveButton(ctx *Ctx, c *cmd) {
	// hit-test
	hot := pointInCmd(ctx, c, ctx.I.MouseX, ctx.I.MouseY)
	st := ctx.state[c.id]

	// active = mouse down started inside
	if ctx.I.MousePressed && hot {
		st.active = true
	}
	if ctx.I.MouseReleased {
		// clicked when released while still hot and was active
		if st.active && hot {
			c.clicked = true
		}
		st.active = false
	}
	st.hot = hot
	ctx.state[c.id] = st // stable map mutation, no alloc after first insert

	// draw bg (simple visual feedback)
	bg := c.bg
	if st.active {
		bg[0] *= 0.85
		bg[1] *= 0.85
		bg[2] *= 0.85
	} else if st.hot {
		bg[0] *= 1.05
		bg[1] *= 1.05
		bg[2] *= 1.05
	}
	if bg[3] > 0 {
		cx := c.x + c.w*0.5
		cy := c.y + c.h*0.5
		ctx.R.DrawQuad(cx, cy, c.w, c.h, bg, 0)
	}

	// draw label centered inside
	tw, th := ctx.R.Measure(c.text, c.fontSize)
	tx := c.x + (c.w-tw)*0.5
	ty := c.y + (c.h-th)*0.5 + th*0.8*0.1
	ctx.R.DrawText(tx, ty, c.text, c.fontSize, c.color)
}

func pointInCmd(ctx *Ctx, c *cmd, x, y float32) bool {
	return x >= c.x && x <= c.x+c.w && y >= c.y && y <= c.y+c.h
}
