package render

import (
	"errors"
	"fmt"
	"image/color"
)

var (
	// ErrShaderNotLinked is returned when the program attempts to use a shader
	// program that was not properly compiled and linked.
	ErrShaderNotLinked = errors.New("shader: invalid state: uniform called before Link()")

	// ErrNotImplemented indicates that the rendering layer in use has
	// functionality that is not yet implemented.
	ErrNotImplemented = errors.New("not implemented")
)

var (
	// BgColor is the default rendering background color used by OpenGL.
	BgColor = Color("#87CEEB")
)

// Color parses the provided web color string and returns a color.RGBA value.
// It panics if the provided string is not in the format #RRGGBB or #RGB.
func Color(s string) (c color.RGBA) {
	var err error
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	if err != nil {
		panic(err)
	}
	return
}
