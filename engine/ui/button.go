package ui

import "github.com/hubastard/grove/engine/text"

type UIButton struct {
	Common[*UIButton]
	text  string
	label *UILabel
}

func Button(str string) *UIButton {
	l := &UIButton{text: str}
	l.Common = NewCommon(l)
	l.label = Label(str)
	l.base.children = append(l.base.children, l.label)
	l.base.color = [4]float32{1, 1, 1, 1}
	l.base.SetPadding(10, 10, 10, 10)
	return l
}
func (l *UIButton) BgColor(color [4]float32) *UIButton   { l.base.color = color; return l }
func (l *UIButton) TextColor(color [4]float32) *UIButton { l.label.base.color = color; return l }
func (l *UIButton) FontSize(size float32) *UIButton      { l.label.fontSize = size; return l }
func (l *UIButton) Font(font *text.Font) *UIButton       { l.label.font = font; return l }

func (l *UIButton) Layout(ctx *Context, constraints Constraints) LayoutResult {
	padding := l.base.Padding()
	innerConstraints := Constraints{
		Min: [2]float32{0, 0},
		Max: [2]float32{
			maxf(0, resolveConstraint(constraints.Max[0])-padding[0]-padding[2]),
			maxf(0, resolveConstraint(constraints.Max[1])-padding[1]-padding[3]),
		},
	}

	res := l.label.Layout(ctx, innerConstraints)
	contentW := res.Size[0]
	contentH := res.Size[1]

	width := l.base.resolveAxis(l.base.widthMod, l.base.widthVal, contentW+padding[0]+padding[2], constraints.Min[0], constraints.Max[0])
	height := l.base.resolveAxis(l.base.heightMod, l.base.heightVal, contentH+padding[1]+padding[3], constraints.Min[1], constraints.Max[1])

	innerWidth := maxf(0, width-padding[0]-padding[2])
	innerHeight := maxf(0, height-padding[1]-padding[3])

	l.base.SetSize(width, height)

	child := l.label.Node()
	childWidth := clamp(contentW, 0, innerWidth)
	if child.widthMod == SizeModeExpand {
		childWidth = innerWidth
	}
	childHeight := clamp(contentH, 0, innerHeight)
	if child.heightMod == SizeModeExpand {
		childHeight = innerHeight
	}
	child.SetSize(childWidth, childHeight)

	return LayoutResult{Size: [2]float32{width, height}}
}

func (l *UIButton) Draw(ctx *Context) {
	padding := l.base.Padding()
	if l.base.color[3] > 0 {
		halfW := l.base.size[0] / 2
		halfH := l.base.size[1] / 2
		ctx.Renderer.DrawQuad(l.base.position[0]+halfW, l.base.position[1]+halfH, l.base.size[0], l.base.size[1], l.base.color, 0)
	}

	child := l.label.Node()
	child.SetPos(l.base.position[0]+padding[0], l.base.position[1]+padding[1])

	l.label.Draw(ctx)
}
