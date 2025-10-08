package ui

import (
	"math"

	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/gfx/renderer2d"
	"github.com/hubastard/grove/engine/text"
)

type SizeMode int

const (
	SizeModeFit SizeMode = iota
	SizeModeFixed
	SizeModeExpand
)

type Constraints struct {
	Min [2]float32
	Max [2]float32
}

type LayoutResult struct {
	Size [2]float32
}

type Context struct {
	Viewport    [4]float32
	DefaultFont *text.Font
	Renderer    *renderer2d.Renderer2D
}

type UIElement interface {
	Node() *Base
	Layout(ctx *Context, constraints Constraints) LayoutResult
	Draw(ctx *Context)
}

type Base struct {
	parent    UIElement
	children  []UIElement
	position  [2]float32
	size      [2]float32
	color     colors.Color
	widthMod  SizeMode
	heightMod SizeMode
	widthVal  float32
	heightVal float32
	padding   [4]float32 // left, top, right, bottom
}

func (b *Base) Parent() UIElement       { return b.parent }
func (b *Base) Children() []UIElement   { return b.children }
func (b *Base) Pos() (x, y float32)     { return b.position[0], b.position[1] }
func (b *Base) Size() (w, h float32)    { return b.size[0], b.size[1] }
func (b *Base) SetPos(x, y float32)     { b.position = [2]float32{x, y} }
func (b *Base) SetSize(w, h float32)    { b.size = [2]float32{w, h} }
func (b *Base) SetColor(c colors.Color) { b.color = c }
func (b *Base) Padding() [4]float32     { return b.padding }
func (b *Base) SetPadding(l, t, r, btm float32) {
	b.padding = [4]float32{l, t, r, btm}
}

func (b *Base) initDefaults() {
	if b.widthMod == 0 && b.widthVal == 0 {
		b.widthMod = SizeModeFit
	}
	if b.heightMod == 0 && b.heightVal == 0 {
		b.heightMod = SizeModeFit
	}
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func resolveConstraint(max float32) float32 {
	if max == 0 {
		return float32(math.MaxFloat32)
	}
	return max
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func (b *Base) resolveAxis(mode SizeMode, fixed, content, min, max float32) float32 {
	switch mode {
	case SizeModeFixed:
		if fixed > 0 {
			return clamp(fixed, min, resolveConstraint(max))
		}
		return clamp(content, min, resolveConstraint(max))
	case SizeModeExpand:
		return clamp(resolveConstraint(max), min, resolveConstraint(max))
	default:
		return clamp(content, min, resolveConstraint(max))
	}
}

func (b *Base) innerPosition() (float32, float32) {
	return b.position[0] + b.padding[0], b.position[1] + b.padding[1]
}

func (b *Base) innerSize() (float32, float32) {
	innerW := b.size[0] - b.padding[0] - b.padding[2]
	innerH := b.size[1] - b.padding[1] - b.padding[3]
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	return innerW, innerH
}

// ------ Helper ------

type Common[T any] struct {
	owner T
	base  Base
}

func NewCommon[T any](owner T) Common[T] {
	var b Base
	b.initDefaults()
	return Common[T]{owner: owner, base: b}
}

func (c *Common[T]) Node() *Base              { return &c.base }
func (c *Common[T]) Position(x, y float32) T  { c.base.SetPos(x, y); return c.owner }
func (c *Common[T]) Size(w, h float32) T      { c.base.SetSize(w, h); return c.owner }
func (c *Common[T]) Color(col colors.Color) T { c.base.SetColor(col); return c.owner }

func (c *Common[T]) WidthFit() T {
	c.base.widthMod = SizeModeFit
	return c.owner
}

func (c *Common[T]) WidthFixed(width float32) T {
	c.base.widthMod = SizeModeFixed
	c.base.widthVal = width
	return c.owner
}

func (c *Common[T]) WidthExpand() T {
	c.base.widthMod = SizeModeExpand
	return c.owner
}

func (c *Common[T]) HeightFit() T {
	c.base.heightMod = SizeModeFit
	return c.owner
}

func (c *Common[T]) HeightFixed(height float32) T {
	c.base.heightMod = SizeModeFixed
	c.base.heightVal = height
	return c.owner
}

func (c *Common[T]) HeightExpand() T {
	c.base.heightMod = SizeModeExpand
	return c.owner
}

func (c *Common[T]) Padding(all float32) T {
	c.base.SetPadding(all, all, all, all)
	return c.owner
}

func (c *Common[T]) Padding2(horizontal, vertical float32) T {
	c.base.SetPadding(horizontal, vertical, horizontal, vertical)
	return c.owner
}

func (c *Common[T]) Padding4(left, top, right, bottom float32) T {
	c.base.SetPadding(left, top, right, bottom)
	return c.owner
}

func (c *Common[T]) Children(kids ...UIElement) T {
	c.base.children = append(c.base.children, kids...)
	for _, k := range kids {
		k.Node().parent = any(c.owner).(UIElement)
	}
	return c.owner
}
