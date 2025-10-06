package core

type Input struct {
	keys           map[Key]bool
	mouseX, mouseY float64
}

func NewInput() *Input { return &Input{keys: map[Key]bool{}} }

func (in *Input) Handle(ev Event) {
	switch e := ev.(type) {
	case EventKey:
		in.keys[e.Key] = e.Down
	case EventMouseMove:
		in.mouseX, in.mouseY = e.X, e.Y
	}
}

func (in *Input) IsKeyDown(k Key) bool      { return in.keys[k] }
func (in *Input) Mouse() (float64, float64) { return in.mouseX, in.mouseY }
