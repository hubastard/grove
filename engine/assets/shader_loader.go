package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadShader reads a GLSL file into a null-terminated string for OpenGL.
func LoadShader(name string) (string, error) {
	path := filepath.Join("assets", "shaders", name)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load shader %q: %w", name, err)
	}
	// Ensure null termination for gl.Str
	if len(b) == 0 || b[len(b)-1] != 0 {
		b = append(b, 0)
	}
	return string(b), nil
}
