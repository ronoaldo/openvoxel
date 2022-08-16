package main

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ronoaldo/openvoxel/log"
	"github.com/ronoaldo/openvoxel/render"
)

var (
	winWidth  int32 = 1024
	winHeight int32 = 768
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

	// Initialize
	gl.Init()
	log.Infof("OpenGL Version: %v", render.Version())

	// Vertex Shader
	vertexShaderSrc :=
		`#version 330 core
		layout (location = 0) in vec3 aPos;
		void main() {
			gl_Position = vec4(aPos.x, aPos.y, aPos.z, 1.0f);
		}`
	vertexShader, err := render.CompileShader(vertexShaderSrc, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	// Fragment Shader
	fragmentShaderSrc :=
		`#version 330 core
		out vec4 FragColor;
		void main() {
			FragColor = vec4(1.0f, 0.5f, 0.2f, 1.0f);
		}`

	fragmentShader, err := render.CompileShader(fragmentShaderSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	shaderProgram, err := render.LinkProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
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

		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(VAO)
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
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
}
