package main

import (
	"os"
	"runtime"
	"time"

	"github.com/ronoaldo/openvoxel/log"
	"github.com/ronoaldo/openvoxel/render"
	"github.com/ronoaldo/openvoxel/transform"

	_ "embed"
)

var (
	winWidth  int = 1920
	winHeight int = 768
)

// f is a syntax suggar to cast any number to float32
func f[X int | int32 | int64 | uint | uint32 | uint64 | float64](i X) float32 {
	return float32(i)
}

var (
	//go:embed shaders/vertex.glsl
	vertexShaderSrc string

	//go:embed shaders/fragment.glsl
	fragmentShaderSrc string

	//go:embed textures/dirt.png
	texDirt []byte
)

func init() {
	runtime.LockOSThread()
}

func main() {
	log.Infof("Initializing main window")
	window, err := render.NewWindow(winWidth, winHeight, "openvoxel.net [Demo]")
	if err != nil {
		log.Warnf("Unable to open new window: %v", err)
	}
	defer window.Close()
	log.Infof("Rendering Backend: %v", render.Version())

	shader := &render.Shader{}
	shader.VertexShader(vertexShaderSrc).FragmentShader(fragmentShaderSrc)
	if err := shader.Link(); err != nil {
		log.Warnf("error linking shader program: %v", err)
		os.Exit(1)
	}

	log.Infof("Rendering cube %v", cube)
	window.Scene().AddVertices(cube)

	tex, err := render.NewTextureFromBytes(texDirt)
	if err != nil {
		log.Warnf("Error loading texture: %v", err)
		os.Exit(1)
	}
	window.Scene().AddTexture(tex)

	// Main program loop
	fov := transform.DegToRad(45)
	frameCount := int32(0)
	start := time.Now()
	lastLog := 0
	for !window.ShouldClose() {
		t := render.Time()

		window.Scene().Clear()

		aspect := f(window.Width) / f(window.Height)
		projection := transform.Perspective(fov, aspect, 0.1, 100)

		shader.Use()
		shader.UniformInts("frameCount", frameCount)
		shader.UniformFloats("renderTime", f(t))
		shader.UniformTransformation("projection", projection)

		// Draw 10x10 blocks of dirt at bottom
		for x := -10; x < 10; x++ {
			for z := -10; z < 10; z++ {
				model := transform.Translate(f(x), 0, f(z))
				shader.UniformTransformation("model", model)
				window.Scene().Draw(shader)
			}
		}

		// Draw a rotating cube above them
		ang := transform.DegToRad(45) * f(t)
		model := transform.Chain(
			transform.Translate(0, 3, 0),
			transform.Rotate(ang, 0, 1, 0),
		)
		shader.UniformTransformation("model", model)
		window.Scene().Draw(shader)

		window.SwapBuffers()
		window.PollEvents()

		frameCount++
		elapsedMs := int(time.Since(start).Milliseconds())
		elapsedSec := elapsedMs / 1000
		fps := f(frameCount) / f(elapsedSec)
		if lastLog != int(elapsedSec) {
			log.Infof("At %d sec, avg FPS %.02f; t=%v", elapsedSec, fps, t)
			lastLog = int(elapsedSec)
		}
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
