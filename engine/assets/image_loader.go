package assets

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

// LoadPNG returns width, height, and tightly packed RGBA8 pixels (row-major, top-left origin).
// We then flip vertically to match OpenGL's bottom-left origin.
func LoadPNG(relPath string) (w, h int, rgba []byte, err error) {
	path := filepath.Join("assets", "textures", relPath)
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("decode png %q: %w", path, err)
	}

	// Ensure RGBA
	rgbaImg := imageToRGBA(img)
	w, h = rgbaImg.Bounds().Dx(), rgbaImg.Bounds().Dy()

	// Repack in tight rows (stride == 4*w)
	out := make([]byte, w*h*4)
	src := rgbaImg.Pix
	srcStride := rgbaImg.Stride

	// Copy row by row
	for y := 0; y < h; y++ {
		copy(out[y*w*4:(y+1)*w*4], src[y*srcStride:y*srcStride+w*4])
	}

	return w, h, out, nil
}

func imageToRGBA(img image.Image) *image.RGBA {
	if m, ok := img.(*image.RGBA); ok && m.Stride == m.Rect.Dx()*4 {
		return m
	}
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Src)
	return dst
}
