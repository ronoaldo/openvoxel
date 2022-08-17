package main

import (
	"os"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ronoaldo/openvoxel/log"
	"github.com/ronoaldo/openvoxel/render"
)

var (
	winWidth  int32 = 1024
	winHeight int32 = 768

	wireFrames bool
)

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(int(winWidth), int(winHeight), "Hello World", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetInputMode(glfw.StickyKeysMode, glfw.True)
	window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		gl.Viewport(0, 0, int32(width), int32(height))
	})

	// Initialize
	gl.Init()
	log.Infof("OpenGL Version: %v", render.Version())

	// Vertex Shader
	shader := &render.Shader{}
	shader.VertexShader("shaders/vertex.glsl").FragmentShader("shaders/fragment.glsl")
	if err := shader.Link(); err != nil {
		log.Warnf("error linking shader program: %v", err)
		os.Exit(1)
	}

	vertices := []float32{
		+0.5, +0.5, 0.0, // top right
		+0.5, -0.5, 0.0, // bottom right
		-0.5, -0.5, 0.0, // bottom left
		-0.5, +0.5, 0.0, // top left
	}
	indices := []uint32{
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}
	log.Infof("Rendering Vertices: %#v", vertices)

	var VAO uint32
	gl.GenBuffers(1, &VAO)
	gl.BindVertexArray(VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	var EBO uint32
	gl.GenBuffers(1, &EBO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)

	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(0)

	for !window.ShouldClose() {
		processInput(window)

		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		shader.Use()
		gl.BindVertexArray(VAO)
		if wireFrames {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		} else {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		}
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func processInput(window *glfw.Window) {
	if window.GetKey(glfw.KeyEscape) == glfw.Press {
		log.Infof("ESC key pressed. Exiting...")
		window.SetShouldClose(true)
	}

	if window.GetKey(glfw.KeyF10) == glfw.Press {
		log.Infof("F10 key pressed. Flipping wireframe mode...")
		wireFrames = !wireFrames
	}
}
