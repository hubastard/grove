package scene

import "github.com/hubastard/grove/engine/core"

// OrthoController2D: WASD move, Q/E rotate, Z/X zoom in/out.
type OrthoController2D struct {
	MoveSpeed float32
	RotSpeed  float32
	ZoomSpeed float32
	Camera    *OrthoCamera2D
}

func NewOrthoController2D(cam *OrthoCamera2D) *OrthoController2D {
	return &OrthoController2D{
		MoveSpeed: 1,
		RotSpeed:  2.0,
		ZoomSpeed: 1.2,
		Camera:    cam,
	}
}

func (cc *OrthoController2D) Update(e *core.Engine, dt float32) {
	in := e.Input
	speed := cc.MoveSpeed * dt
	// rotSpeed := cc.RotSpeed * dt

	if in.IsKeyDown(core.KeyW) {
		cc.Camera.Move(0, speed)
	}
	if in.IsKeyDown(core.KeyS) {
		cc.Camera.Move(0, -speed)
	}
	if in.IsKeyDown(core.KeyA) {
		cc.Camera.Move(-speed, 0)
	}
	if in.IsKeyDown(core.KeyD) {
		cc.Camera.Move(speed, 0)
	}

	// Q/E rotate (optional)
	// map Q/E to your Key enum as needed; if not present, omit rotation controls
	// if in.IsKeyDown(core.KeyQ) { cc.Camera.Rotate(rotSpeed) }
	// if in.IsKeyDown(core.KeyE) { cc.Camera.Rotate(-rotSpeed) }

	// Zoom via Z/X (discrete)
	// if in.IsKeyDown(core.KeyZ) { cc.Camera.SetZoom(cc.Camera.Zoom / cc.ZoomSpeed) }
	// if in.IsKeyDown(core.KeyX) { cc.Camera.SetZoom(cc.Camera.Zoom * cc.ZoomSpeed) }
}
