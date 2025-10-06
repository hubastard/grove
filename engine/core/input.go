package core

type Input struct {
	keys             map[Key]bool
	mods             Mod
	mouse            map[MouseButton]bool
	mouseX, mouseY   float64
	scrollX, scrollY float64
}

func NewInput() *Input {
	return &Input{
		keys:  map[Key]bool{},
		mods:  ModNone,
		mouse: map[MouseButton]bool{},
	}
}

func (in *Input) Handle(ev Event) {
	switch e := ev.(type) {
	case EventKey:
		in.keys[e.Key] = e.Down
		in.mods = e.Mods
	case EventMouseButton:
		in.mouse[e.Button] = e.Down
	case EventMouseMove:
		in.mouseX, in.mouseY = e.X, e.Y
	case EventScroll:
		in.scrollX, in.scrollY = e.Xoff, e.Yoff
	}
}

func (in *Input) IsKeyDown(k Key) bool                 { return in.keys[k] }
func (in *Input) IsModActive(m Mod) bool               { return in.mods&m != 0 }
func (in *Input) IsMouseButtonDown(b MouseButton) bool { return in.mouse[b] }
func (in *Input) Mouse() (float64, float64)            { return in.mouseX, in.mouseY }
func (in *Input) Scroll() (float64, float64)           { return in.scrollX, in.scrollY }

// -------- KeyCodes, Mods, MouseButtons --------

type Key int

const (
	KeyUnknown Key = -1

	// Printable keys
	KeySpace        Key = 32
	KeyApostrophe   Key = 39 /* ' */
	KeyComma        Key = 44 /* , */
	KeyMinus        Key = 45 /* - */
	KeyPeriod       Key = 46 /* . */
	KeySlash        Key = 47 /* / */
	Key0            Key = 48
	Key1            Key = 49
	Key2            Key = 50
	Key3            Key = 51
	Key4            Key = 52
	Key5            Key = 53
	Key6            Key = 54
	Key7            Key = 55
	Key8            Key = 56
	Key9            Key = 57
	KeySemicolon    Key = 59 /* ; */
	KeyEqual        Key = 61 /* = */
	KeyA            Key = 65
	KeyB            Key = 66
	KeyC            Key = 67
	KeyD            Key = 68
	KeyE            Key = 69
	KeyF            Key = 70
	KeyG            Key = 71
	KeyH            Key = 72
	KeyI            Key = 73
	KeyJ            Key = 74
	KeyK            Key = 75
	KeyL            Key = 76
	KeyM            Key = 77
	KeyN            Key = 78
	KeyO            Key = 79
	KeyP            Key = 80
	KeyQ            Key = 81
	KeyR            Key = 82
	KeyS            Key = 83
	KeyT            Key = 84
	KeyU            Key = 85
	KeyV            Key = 86
	KeyW            Key = 87
	KeyX            Key = 88
	KeyY            Key = 89
	KeyZ            Key = 90
	KeyLeftBracket  Key = 91 /* [ */
	KeyBackslash    Key = 92 /* \ */
	KeyRightBracket Key = 93 /* ] */
	KeyGraveAccent  Key = 96 /* ` */

	// Function keys
	KeyEscape      Key = 256
	KeyEnter       Key = 257
	KeyTab         Key = 258
	KeyBackspace   Key = 259
	KeyInsert      Key = 260
	KeyDelete      Key = 261
	KeyRight       Key = 262
	KeyLeft        Key = 263
	KeyDown        Key = 264
	KeyUp          Key = 265
	KeyPageUp      Key = 266
	KeyPageDown    Key = 267
	KeyHome        Key = 268
	KeyEnd         Key = 269
	KeyCapsLock    Key = 280
	KeyScrollLock  Key = 281
	KeyNumLock     Key = 282
	KeyPrintScreen Key = 283
	KeyPause       Key = 284
	KeyF1          Key = 290
	KeyF2          Key = 291
	KeyF3          Key = 292
	KeyF4          Key = 293
	KeyF5          Key = 294
	KeyF6          Key = 295
	KeyF7          Key = 296
	KeyF8          Key = 297
	KeyF9          Key = 298
	KeyF10         Key = 299
	KeyF11         Key = 300
	KeyF12         Key = 301
	KeyF13         Key = 302
	KeyF14         Key = 303
	KeyF15         Key = 304
	KeyF16         Key = 305
	KeyF17         Key = 306
	KeyF18         Key = 307
	KeyF19         Key = 308
	KeyF20         Key = 309
	KeyF21         Key = 310
	KeyF22         Key = 311
	KeyF23         Key = 312
	KeyF24         Key = 313
	KeyF25         Key = 314
	KeyPad0        Key = 320
	KeyPad1        Key = 321
	KeyPad2        Key = 322
	KeyPad3        Key = 323
	KeyPad4        Key = 324
	KeyPad5        Key = 325
	KeyPad6        Key = 326
	KeyPad7        Key = 327
	KeyPad8        Key = 328
	KeyPad9        Key = 329
	KeyPadDecimal  Key = 330
	KeyPadDivide   Key = 331
	KeyPadMultiply Key = 332
	KeyPadSubtract Key = 333
	KeyPadAdd      Key = 334
	KeyPadEnter    Key = 335
	KeyPadEqual    Key = 336
	KeyLeftShift   Key = 340
	KeyLeftCtrl    Key = 341
	KeyLeftAlt     Key = 342
	KeyLeftSuper   Key = 343
	KeyRightShift  Key = 344
	KeyRightCtrl   Key = 345
	KeyRightAlt    Key = 346
	KeyRightSuper  Key = 347
)

type Mod int

const (
	ModNone     Mod = 0
	ModShift    Mod = 1 << 0
	ModCtrl     Mod = 1 << 1
	ModAlt      Mod = 1 << 2
	ModSuper    Mod = 1 << 3
	ModCapsLock Mod = 1 << 4
	ModNumLock  Mod = 1 << 5
)

type MouseButton int

const (
	MouseButtonLeft   MouseButton = 0
	MouseButtonRight  MouseButton = 1
	MouseButtonMiddle MouseButton = 2
	MouseButton4      MouseButton = 3
	MouseButton5      MouseButton = 4
	MouseButton6      MouseButton = 5
	MouseButton7      MouseButton = 6
	MouseButton8      MouseButton = 7
)
