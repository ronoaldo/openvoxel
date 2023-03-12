package main

import (
	"os"
	"runtime"

	"github.com/ronoaldo/openvoxel/log"
	"github.com/ronoaldo/openvoxel/render"
	"github.com/ronoaldo/openvoxel/transform"
)

var (
	winWidth  int = 800
	winHeight int = 600

	f = func(i int) float32 { return float32(i) }
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

	log.Infof("Rendering cube %v", cube)
	window.Scene().AddVertices(cube)

	tex, err := render.NewTexture("textures/dirt.png")
	if err != nil {
		log.Warnf("Error loading texture: %v", err)
		os.Exit(1)
	}
	window.Scene().AddTexture(tex)

	// Main program loop
	view := transform.Translate(0, 0, -15).Mul4(transform.Rotate(transform.DegToRad(45), 0.5, 1, 0))
	fov := transform.DegToRad(45)
	aspect := float32(winWidth) / float32(winHeight)
	projection := transform.Perspective(fov, aspect, 0.1, 100)
	for !window.ShouldClose() {
		t := render.Time()

		window.Scene().Clear()

		shader.UniformFloats("renderTime", float32(t))
		shader.UniformTransformation("projection", projection)
		shader.UniformTransformation("view", view)

		// Draw 10x10 blocks of dirt
		for x := -100; x < 100; x++ {
			for z := -100; z < 100; z++ {
				model := transform.Translate(f(x), 0, f(z))
				shader.UniformTransformation("model", model)
				window.Scene().Draw(shader)
			}
		}

		window.SwapBuffers()
		window.PollEvents()
	}
}

var (
	cube = []float32{
		// positions       // tex coords
		-0.5, -0.5, -0.5, 0.0, 0.0,
		0.5, -0.5, -0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 0.0,

		-0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 1.0,
		-0.5, 0.5, 0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,

		-0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, -0.5, 1.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,
		-0.5, 0.5, 0.5, 1.0, 0.0,

		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, 0.5, 0.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0,

		-0.5, -0.5, -0.5, 0.0, 1.0,
		0.5, -0.5, -0.5, 1.0, 1.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		0.5, -0.5, 0.5, 1.0, 0.0,
		-0.5, -0.5, 0.5, 0.0, 0.0,
		-0.5, -0.5, -0.5, 0.0, 1.0,

		-0.5, 0.5, -0.5, 0.0, 1.0,
		0.5, 0.5, -0.5, 1.0, 1.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		0.5, 0.5, 0.5, 1.0, 0.0,
		-0.5, 0.5, 0.5, 0.0, 0.0,
		-0.5, 0.5, -0.5, 0.0, 1.0,
	}
)
