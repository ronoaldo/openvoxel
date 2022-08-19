package main

import (
	"math"
	"os"
	"runtime"

	"github.com/ronoaldo/openvoxel/log"
	"github.com/ronoaldo/openvoxel/render"
)

var (
	winWidth  int = 1024
	winHeight int = 768

	wireFrames bool
)

func init() {
	runtime.LockOSThread()
}

func main() {
	log.Infof("Initializing main window")
	window, err := render.NewWindow(winWidth, winHeight, "LearnOpenGL.com")
	if err != nil {
		log.Warnf("Unable to open new window: %v", err)
	}
	defer window.Close()
	log.Infof("Rendering Backend: %v", render.Version())

	shader := &render.Shader{}
	shader.VertexShaderFile("shaders/vertex.glsl").FragmentShaderFile("shaders/fragment.glsl")
	if err := shader.Link(); err != nil {
		log.Warnf("error linking shader program: %v", err)
		os.Exit(1)
	}

	vertices := []float32{
		// positions     // colors      // tex coords
		+0.5, +0.5, 0.0, 1.0, 0.0, 0.0, 1.0, 1.0, // top right
		+0.5, -0.5, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom right
		-0.5, -0.5, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom left
		-0.5, +0.5, 0.0, 1.0, 1.0, 0.0, 0.0, 1.0, // top left
	}
	indices := []uint32{
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}
	log.Infof("Rendering Vertices: %#v (indices: %#v)", vertices, indices)
	window.Scene().AddTriangles(vertices, indices)

	tex, err := render.NewTexture("textures/dirt.png")
	if err != nil {
		log.Warnf("Error loading texture: %v", err)
		os.Exit(1)
	}
	window.Scene().AddTexture(tex)

	// Main program loop
	for !window.ShouldClose() {
		t := render.Time()
		greenValue := math.Sin(t)/2.0 + 0.5
		shader.UniformFloats("progressiveColor", 0.0, float32(greenValue), 0.0, 1.0)

		window.Scene().Draw(shader)
		window.SwapBuffers()
		window.PollEvents()
	}
}
