package render

import (
	"errors"
)

var (
	errShaderNotLinked = errors.New("shader: invalid state: uniform called before Link()")

	errNotImplemented = errors.New("not implemented")
)
