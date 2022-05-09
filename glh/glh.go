// package glh implements some OpenGL helpers and wrappers for more idiomatic go.
package glh

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

func GetVersion() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

func CompileShader(shaderSource string, shaderType uint32) (uint32, error) {
	fmt.Printf("Compiling \n%s\n...", shaderSource)
	shader := gl.CreateShader(shaderType)
	csource, free := gl.Strs(shaderSource + "\x00")
	gl.ShaderSource(shader, 1, csource, nil)
	free()
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, errors.New("Failed to compile shader: \n" + log)
	}
	fmt.Printf("Shader compiled (status=%v)\n", status)
	return shader, nil
}

func LinkProgram(vertexShader, fragmentShader uint32) (uint32, error) {
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	var status int32
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		return 0, errors.New("Failed to create shader program: \n" + log)
	}
	fmt.Printf("Shader program linked properly (status=%v)\n", status)
	fmt.Println("Freeing shader resources ...")
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	return shaderProgram, nil
}
