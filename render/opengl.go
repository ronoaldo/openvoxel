// package render implements OpenGL rendering logic and wrappers.
package render

import (
	"errors"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/ronoaldo/openvoxel/log"
)

// Version returns the OpengGL Version
func Version() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

// CompileShader takes a GLSL shader source string and type and compiles it.
// It returns the shader pointer or an error if the compilation failed.
func CompileShader(shaderSource string, shaderType uint32) (uint32, error) {
	log.Infof("Compiling shader (type=%v): %#s", shaderType, shaderSource)
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
		return 0, errors.New("Failed to compile shader: " + log)
	}
	log.Infof("Shader compiled (status=%v)", status)
	return shader, nil
}

// LinkProgram takes an array of compiled shaders and link them into a usable program.
func LinkProgram(shaders ...uint32) (uint32, error) {
	shaderProgram := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(shaderProgram, shader)
	}
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
	log.Infof("Shader program linked properly (status=%v)", status)
	log.Infof("Freeing shader resources ...")
	for _, shader := range shaders {
		gl.DeleteShader(shader)
	}
	return shaderProgram, nil
}
