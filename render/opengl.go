// package render implements OpenGL rendering logic and wrappers.
package render

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/ronoaldo/openvoxel/log"
)

// Version returns the OpengGL version as reported by the driver.
func Version() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

// ShaderFile is a helper struct that represents a shader file and it's type.
type shaderFile struct {
	path       string
	shaderType uint32
}

// Shader abstracts raw GLSL shader compilation, linking and usage.
type Shader struct {
	shaderFiles []shaderFile
	program     *uint32
}

// VertexShader appends the provider shader file to the pipeline. This method
// returns the Shader reference to allow for chaining.
func (s *Shader) VertexShader(path string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderFile{path, gl.VERTEX_SHADER})
	return s
}

// FragmentShader appends the provided shader file to the pipeline. This method
// returns the Shader reference to allow for chaining.
func (s *Shader) FragmentShader(path string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderFile{path, gl.FRAGMENT_SHADER})
	return s
}

// UniformInt adds the provided int value as a GLSL uniform with the given
// name.  Returns an error if the shader was not linked.
func (s *Shader) UniformInt(name string, value int32) error {
	if s.program == nil {
		return errors.New("shader: invalid state: uniform called before Link()")
	}
	gl.Uniform1i(gl.GetUniformLocation(*s.program, gl.Str(name)), value)
	return nil
}

// UniformFloat adds the provided float value as a GLSL uniform with the given
// name.  Returns an error if the program was not linked.
func (s *Shader) UniformFloat(name string, value float32) error {
	if s.program == nil {
		return errors.New("shader: invalid state: uniform called before Link()")
	}
	gl.Uniform1f(gl.GetUniformLocation(*s.program, gl.Str(name)), value)
	return nil
}

// Link creates an OpenGL shader program linking all previously compiled
// shaders. It reports an error if no shaders where compiled, or if there were
// an error linking them.
func (s *Shader) Link() error {
	shaders := []uint32{}
	for _, file := range s.shaderFiles {
		b, err := ioutil.ReadFile(file.path)
		if err != nil {
			return err
		}

		shaderId, err := s.compileShader(string(b), file.shaderType)
		if err != nil {
			return err
		}

		shaders = append(shaders, shaderId)
	}

	p, err := s.linkProgram(shaders...)
	if err != nil {
		return err
	}

	s.program = new(uint32)
	*s.program = p
	return nil
}

// Use attempt to use the linked program by calling gl.UseProgram.  It will
// panic if no shaders were compiled and linked previously.
func (s *Shader) Use() {
	if s.program == nil {
		panic("shader program not linked; call Shader.Link() first")
	}
	gl.UseProgram(*s.program)
}

// compileShader takes a GLSL shader source string and type and compiles it.
// It returns the shader pointer or an error if the compilation failed.
func (s *Shader) compileShader(shaderSource string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	log.Infof("Compiling shader (type=%v): %#s", shaderType, shaderSource)
	csource, free := gl.Strs(shaderSource + "\x00")
	defer free()
	gl.ShaderSource(shader, 1, csource, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		// If status is FALSE, compilation failed so we report any issues
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, errors.New("Failed to compile shader: " + log)
	}

	log.Infof("Shader compiled (status=%v)", status)
	return shader, nil
}

// linkProgram takes an array of compiled shaders and link them into a usable program.
func (s *Shader) linkProgram(shaders ...uint32) (uint32, error) {
	shaderProgram := gl.CreateProgram()

	log.Infof("Linking shaders into program ...")
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

	for _, shader := range shaders {
		gl.DeleteShader(shader)
	}
	return shaderProgram, nil
}
