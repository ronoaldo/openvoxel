package transform

import "github.com/go-gl/mathgl/mgl32"

func RadToDeg(rad float32) float32 {
	return mgl32.RadToDeg(rad)
}
func DegToRad(deg float32) float32 {
	return mgl32.DegToRad(deg)
}

// Rotate generates a matrix transformation by the provided angle in degrees.
func Rotate(deg float32) mgl32.Mat4 {
	trans := mgl32.
		Rotate3DZ(mgl32.DegToRad(deg)).
		Mat4().Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
	return trans
}
