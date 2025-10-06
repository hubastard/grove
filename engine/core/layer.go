package core

type Layer interface {
	OnAttach(e *Engine)
	OnDetach(e *Engine)
	OnUpdate(e *Engine, dt float64)
	OnRender(e *Engine, alpha float64)
	OnEvent(e *Engine, ev Event) bool // return true if handled; propagation stops
}

type LayerStack struct{ list []Layer }

func (ls *LayerStack) Push(l Layer) { ls.list = append(ls.list, l) }
func (ls *LayerStack) Pop() (Layer, bool) {
	if len(ls.list) == 0 {
		return nil, false
	}
	i := len(ls.list) - 1
	l := ls.list[i]
	ls.list = ls.list[:i]
	return l, true
}

func (ls *LayerStack) ForEach(f func(Layer)) {
	for _, l := range ls.list {
		f(l)
	}
}

func (ls *LayerStack) ForEachReverse(f func(Layer) bool) {
	for i := len(ls.list) - 1; i >= 0; i-- {
		if stop := f(ls.list[i]); stop {
			break
		}
	}
}
