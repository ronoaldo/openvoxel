package render

import (
	"errors"
	"image/color"
)

var (
	ErrShaderNotLinked = errors.New("shader: invalid state: uniform called before Link()")

	ErrNotImplemented = errors.New("not implemented")
)

var (
	BgColor = color.Black
)
