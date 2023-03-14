package render

import (
	"errors"
	"image/color"
)

var (
	errShaderNotLinked = errors.New("shader: invalid state: uniform called before Link()")

	errNotImplemented = errors.New("not implemented")
)

var (
	BgColor = color.Black
)
