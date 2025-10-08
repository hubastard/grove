package ui

import (
	"math"
	"strings"

	"github.com/hubastard/grove/engine/colors"
	"github.com/hubastard/grove/engine/text"
)

type UILabel struct {
	Common[*UILabel]
	text      string
	fontSize  float32
	font      *text.Font
	wrap      bool
	maxWidth  float32
	layoutStr string
}

func Label(str string) *UILabel {
	l := &UILabel{text: str, fontSize: 16}
	l.Common = NewCommon(l)
	l.base.color = colors.White
	return l
}
func (l *UILabel) FontSize(size float32) *UILabel { l.fontSize = size; return l }
func (l *UILabel) Font(font *text.Font) *UILabel  { l.font = font; return l }
func (l *UILabel) Color(c colors.Color) *UILabel  { l.base.color = c; return l }
func (l *UILabel) Wrap(enabled bool) *UILabel     { l.wrap = enabled; return l }
func (l *UILabel) MaxWidth(width float32) *UILabel {
	l.maxWidth = width
	if width > 0 {
		l.wrap = true
	}
	return l
}

func (l *UILabel) Layout(ctx *Context, constraints Constraints) LayoutResult {
	if l.font == nil {
		l.font = ctx.DefaultFont
	}
	if l.font == nil {
		return LayoutResult{}
	}

	padding := l.base.Padding()
	effectiveMax := resolveConstraint(constraints.Max[0])
	if effectiveMax == float32(math.MaxFloat32) {
		effectiveMax = 0
	}
	if l.maxWidth > 0 && (effectiveMax == 0 || l.maxWidth < effectiveMax) {
		effectiveMax = l.maxWidth
	}
	if effectiveMax > 0 {
		effectiveMax -= padding[0] + padding[2]
		if effectiveMax < 0 {
			effectiveMax = 0
		}
	}

	contentW, contentH, laidOut := l.measureText(effectiveMax)
	l.layoutStr = laidOut

	width := l.base.resolveAxis(l.base.widthMod, l.base.widthVal, contentW+padding[0]+padding[2], constraints.Min[0], constraints.Max[0])
	height := l.base.resolveAxis(l.base.heightMod, l.base.heightVal, contentH+padding[1]+padding[3], constraints.Min[1], constraints.Max[1])

	l.base.SetSize(width, height)
	return LayoutResult{Size: [2]float32{width, height}}
}

func (l *UILabel) Draw(ctx *Context) {
	if l.layoutStr == "" {
		l.layoutStr = l.text
	}
	if l.layoutStr != "" && l.font != nil && l.base.color[3] > 0 {
		padding := l.base.Padding()
		x := l.base.position[0] + padding[0]
		y := l.base.position[1] + padding[1]
		text.DrawText(ctx.Renderer, l.font, x, y, l.layoutStr, l.base.color)
	}
}

func (l *UILabel) measureText(maxWidth float32) (float32, float32, string) {
	if l.text == "" {
		return 0, 0, ""
	}
	if !l.wrap || maxWidth <= 0 {
		w, h := text.MeasureText(l.font, l.text, l.fontSize)
		if h == 0 {
			h = text.LineHeight(l.font)
		}
		return w, h, l.text
	}

	baseLineH := text.LineHeight(l.font)
	spaceWidth, _ := text.MeasureText(l.font, " ", l.fontSize)

	var wrapped []string
	var maxLineWidth float32

	rawLines := strings.Split(l.text, "\n")
	for _, raw := range rawLines {
		words := strings.Fields(raw)
		if len(words) == 0 {
			wrapped = append(wrapped, "")
			continue
		}

		current := words[0]
		currentWidth, _ := text.MeasureText(l.font, current, l.fontSize)
		for _, word := range words[1:] {
			wordWidth, _ := text.MeasureText(l.font, word, l.fontSize)
			required := currentWidth + spaceWidth + wordWidth
			if required > maxWidth {
				wrapped = append(wrapped, current)
				if currentWidth > maxLineWidth {
					maxLineWidth = currentWidth
				}
				current = word
				currentWidth = wordWidth
			} else {
				current += " " + word
				currentWidth += spaceWidth + wordWidth
			}
		}
		wrapped = append(wrapped, current)
		if currentWidth > maxLineWidth {
			maxLineWidth = currentWidth
		}
	}

	if len(wrapped) == 0 {
		wrapped = append(wrapped, "")
	}
	if maxLineWidth == 0 {
		for _, w := range wrapped {
			width, _ := text.MeasureText(l.font, w, l.fontSize)
			if width > maxLineWidth {
				maxLineWidth = width
			}
		}
	}

	lineHeight := baseLineH
	if lineHeight == 0 {
		lineHeight = 1
	}
	height := lineHeight * float32(len(wrapped))
	joined := strings.Join(wrapped, "\n")
	return maxLineWidth, height, joined
}
