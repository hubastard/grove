package ui

type Align int

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
	AlignStretch
)

type UICanvas struct {
	Common[*UICanvas]
	vertical   bool
	gap        float32
	mainAlign  Align
	crossAlign Align
}

func Canvas(children ...UIElement) *UICanvas {
	c := &UICanvas{
		gap:        10,
		mainAlign:  AlignStart,
		crossAlign: AlignStart,
	}
	c.Common = NewCommon(c)
	c.base.children = children
	return c
}

func (l *UICanvas) LayoutVertically(vertical bool) *UICanvas {
	l.vertical = vertical
	return l
}

func (l *UICanvas) Gap(g float32) *UICanvas {
	l.gap = g
	return l
}

func (l *UICanvas) AlignMain(a Align) *UICanvas {
	l.mainAlign = a
	return l
}

func (l *UICanvas) AlignCross(a Align) *UICanvas {
	l.crossAlign = a
	return l
}

func (l *UICanvas) Layout(ctx *Context, constraints Constraints) LayoutResult {
	padding := l.base.Padding()
	maxWidth := resolveConstraint(constraints.Max[0])
	maxHeight := resolveConstraint(constraints.Max[1])
	minWidth := constraints.Min[0]
	minHeight := constraints.Min[1]

	innerMaxWidth := maxf(0, maxWidth-padding[0]-padding[2])
	innerMaxHeight := maxf(0, maxHeight-padding[1]-padding[3])
	innerMinWidth := maxf(0, minWidth-padding[0]-padding[2])
	innerMinHeight := maxf(0, minHeight-padding[1]-padding[3])

	children := l.base.children
	childSizes := make([][2]float32, len(children))

	var fixedMainSum float32
	var expandMainSum float32
	var maxCross float32
	var expandCount int

	childConstraints := Constraints{
		Min: [2]float32{0, 0},
		Max: [2]float32{innerMaxWidth, innerMaxHeight},
	}

	for i, child := range children {
		res := child.Layout(ctx, childConstraints)
		size := res.Size
		childSizes[i] = size
		if l.vertical {
			if child.Node().heightMod == SizeModeExpand {
				expandMainSum += size[1]
				expandCount++
			} else {
				fixedMainSum += size[1]
			}
			maxCross = maxf(maxCross, size[0])
		} else {
			if child.Node().widthMod == SizeModeExpand {
				expandMainSum += size[0]
				expandCount++
			} else {
				fixedMainSum += size[0]
			}
			maxCross = maxf(maxCross, size[1])
		}
	}

	gapTotal := float32(0)
	if len(children) > 1 {
		gapTotal = l.gap * float32(len(children)-1)
	}

	var innerMainTarget float32
	var innerCrossTarget float32

	if l.vertical {
		contentMain := fixedMainSum + expandMainSum + gapTotal
		outerHeight := l.base.resolveAxis(l.base.heightMod, l.base.heightVal, contentMain+padding[1]+padding[3], minHeight, constraints.Max[1])
		innerMainTarget = maxf(0, outerHeight-padding[1]-padding[3])
		innerMainTarget = maxf(innerMainTarget, innerMinHeight)
		outerWidth := l.base.resolveAxis(l.base.widthMod, l.base.widthVal, maxCross+padding[0]+padding[2], minWidth, constraints.Max[0])
		innerCrossTarget = maxf(0, outerWidth-padding[0]-padding[2])
		innerCrossTarget = maxf(innerCrossTarget, innerMinWidth)
		l.base.SetSize(outerWidth, outerHeight)
	} else {
		contentMain := fixedMainSum + expandMainSum + gapTotal
		outerWidth := l.base.resolveAxis(l.base.widthMod, l.base.widthVal, contentMain+padding[0]+padding[2], minWidth, constraints.Max[0])
		innerMainTarget = maxf(0, outerWidth-padding[0]-padding[2])
		innerMainTarget = maxf(innerMainTarget, innerMinWidth)
		outerHeight := l.base.resolveAxis(l.base.heightMod, l.base.heightVal, maxCross+padding[1]+padding[3], minHeight, constraints.Max[1])
		innerCrossTarget = maxf(0, outerHeight-padding[1]-padding[3])
		innerCrossTarget = maxf(innerCrossTarget, innerMinHeight)
		l.base.SetSize(outerWidth, outerHeight)
	}

	// Distribute extra space along main axis to expanding children.
	mainMinTotal := fixedMainSum + expandMainSum
	if expandCount > 0 {
		extra := innerMainTarget - (mainMinTotal + gapTotal)
		if extra < 0 {
			extra = 0
		}
		share := extra / float32(expandCount)
		for i, child := range children {
			if l.vertical {
				if child.Node().heightMod == SizeModeExpand {
					childSizes[i][1] += share
				}
			} else {
				if child.Node().widthMod == SizeModeExpand {
					childSizes[i][0] += share
				}
			}
		}
	}

	innerOriginX, innerOriginY := l.base.innerPosition()
	mainCursor := float32(0)

	// Calculate total space used after potential expansion for alignment.
	mainUsed := float32(0)
	for i := range children {
		if l.vertical {
			mainUsed += childSizes[i][1]
		} else {
			mainUsed += childSizes[i][0]
		}
	}
	mainUsed += gapTotal

	startOffset := float32(0)
	remaining := innerMainTarget - mainUsed
	if remaining < 0 {
		remaining = 0
	}
	switch l.mainAlign {
	case AlignCenter:
		startOffset = remaining * 0.5
	case AlignEnd:
		startOffset = remaining
	default:
		startOffset = 0
	}
	mainCursor = startOffset

	for i, child := range children {
		childSize := childSizes[i]
		childBase := child.Node()
		if l.vertical {
			width := childSize[0]
			if l.crossAlign == AlignStretch || childBase.widthMod == SizeModeExpand {
				width = innerCrossTarget
			}
			width = clamp(width, 0, innerCrossTarget)

			var x float32
			switch l.crossAlign {
			case AlignCenter:
				x = innerOriginX + (innerCrossTarget-width)/2
			case AlignEnd:
				x = innerOriginX + (innerCrossTarget - width)
			default:
				x = innerOriginX
			}
			y := innerOriginY + mainCursor
			height := childSize[1]
			childBase.SetPos(x, y)
			childBase.SetSize(width, height)
			mainCursor += height
		} else {
			height := childSize[1]
			if l.crossAlign == AlignStretch || childBase.heightMod == SizeModeExpand {
				height = innerCrossTarget
			}
			height = clamp(height, 0, innerCrossTarget)

			var y float32
			switch l.crossAlign {
			case AlignCenter:
				y = innerOriginY + (innerCrossTarget-height)/2
			case AlignEnd:
				y = innerOriginY + (innerCrossTarget - height)
			default:
				y = innerOriginY
			}
			x := innerOriginX + mainCursor
			width := childSize[0]
			childBase.SetPos(x, y)
			childBase.SetSize(width, height)
			mainCursor += width
		}
		if i < len(children)-1 {
			mainCursor += l.gap
		}
	}

	return LayoutResult{Size: l.base.size}
}

func (l *UICanvas) Draw(ctx *Context) {
	if l.base.parent == nil {
		l.base.SetPos(ctx.Viewport[0], ctx.Viewport[1])
		constraints := Constraints{
			Min: [2]float32{0, 0},
			Max: [2]float32{ctx.Viewport[2], ctx.Viewport[3]},
		}
		l.Layout(ctx, constraints)
	}

	if l.base.color[3] > 0 {
		halfW := l.base.size[0] / 2
		halfH := l.base.size[1] / 2
		ctx.Renderer.DrawQuad(l.base.position[0]+halfW, l.base.position[1]+halfH, l.base.size[0], l.base.size[1], l.base.color, 0)
	}

	for _, c := range l.base.children {
		c.Draw(ctx)
	}
}
