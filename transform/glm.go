package transform

import glm "github.com/go-gl/mathgl/mgl32"

func RadToDeg(rad float32) float32 {
	return glm.RadToDeg(rad)
}
func DegToRad(deg float32) float32 {
	return glm.DegToRad(deg)
}

// Rotate generates a matrix transformation by the provided angle in degrees, at
// the given axis vector described by x,y and z.
func Rotate(rad float32, x, y, z float32) glm.Mat4 {
	return glm.HomogRotate3D(rad, glm.Vec3{x, y, z})
}

func Translate(x, y, z float32) glm.Mat4 {
	return glm.Translate3D(x, y, z)
}

func Perspective(fov float32, aspect float32, near, far float32) glm.Mat4 {
	return glm.Perspective(fov, aspect, near, far)
}
