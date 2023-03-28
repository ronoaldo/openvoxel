// package transform implements several transformations using vectors and
// matrices.
package transform

import glm "github.com/go-gl/mathgl/mgl32"

// RadToDeg converts the value in radians to degrees.
func RadToDeg(rad float32) float32 {
	return glm.RadToDeg(rad)
}

// DegToRad convertes the value in degrees to radians.
func DegToRad(deg float32) float32 {
	return glm.DegToRad(deg)
}

// Rotate generates a matrix transformation by the provided angle in degrees, at
// the given axis vector described by x,y and z.
func Rotate(rad float32, x, y, z float32) glm.Mat4 {
	return glm.HomogRotate3D(rad, glm.Vec3{x, y, z})
}

// Translate creates a Mat4 that applies a translation in x, y and z.
func Translate(x, y, z float32) glm.Mat4 {
	return glm.Translate3D(x, y, z)
}

// Perspective creates a Mat4 that applies the perspective using the provided
// field of view (FOV), aspect ratio at the near/far planes.
func Perspective(fov float32, aspect float32, near, far float32) glm.Mat4 {
	return glm.Perspective(fov, aspect, near, far)
}

// Chain can be used to chain several Mat4 operations togheter. All matrices
// provided are multiplied one after the other, and the final result is
// returned.
func Chain(operations ...glm.Mat4) glm.Mat4 {
	if len(operations) == 0 {
		panic("transform.Chain: at least one operation required for chaining")
	}
	op := operations[0]
	for _, other := range operations[1:] {
		op = op.Mul4(other)
	}
	return op
}
