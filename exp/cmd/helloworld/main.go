package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	winWidth  int32 = 1024
	winHeight int32 = 768
)

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Hello Triangles", 200, 200, winWidth, winHeight, sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}
	window.GLCreateContext()
	defer window.Destroy()

	gl.Init()
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL Version:", version)

	// Vertex Shader
	vertexShaderSrc :=
		`#version 330 core
		layout (location = 0) in vec3 aPos;
		void main() {
			gl_Position = vec4(aPos.x, aPos.y, aPos.z, 1.0);
		}`
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexShaderSrc)
	gl.ShaderSource(vertexShader, 1, csource, nil)
	free()
	gl.CompileShader(vertexShader)
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
		panic("Failed to compile vertex shader: \n" + log)
	}
	fmt.Printf("Vertex compiled with success (status=%v)\n", status)

	// Fragment Shader
	fragmentShaderSrc :=
		`#version 330 core
		out vec4 FragColor;
		void main() {
			FragColor = vec4(1.0f, 0.5f, 0.2f, 1.0f);
		}` + "\x00"

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentShaderSrc)
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	free()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		panic("Failed to compile fragment shader: \n" + log)
	}
	fmt.Printf("Fragment compiled with success (status=%v)\n", status)

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		panic("Failed to compile fragment shader: \n" + log)
	}
	fmt.Printf("Shader program linked properly (status=%v)\n", status)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	fmt.Println("Shaders freed to release resources.")

	vertices := []float32{
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
		0.0, 0.5, 0.0,
	}
	fmt.Println("Rendering Vertices:", vertices)

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	var VAO uint32
	gl.GenBuffers(1, &VAO)
	gl.BindVertexArray(VAO)

	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)

	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(0)

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		gl.ClearColor(0.0, 0.0, 0.0, 0.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		window.GLSwap()
	}
}
