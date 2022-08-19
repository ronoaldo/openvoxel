// package render implements OpenGL rendering logic and wrappers.
package render

import (
	"errors"
	"image/color"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ronoaldo/openvoxel/log"
)

// Version returns the OpengGL version as reported by the driver.
func Version() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

// Window handles the basic GUI and Input event handling.
//
// Window must be created using NewWindow, which will load all the required
// components of OpenGL and GLFW, as well as initializing a Scene object that
// can be used to draw elements on. To add elements call the Scene() method.
//
// During the app main event loop, callers must make sure to call the
// PollEvents() method to receive any window/input changes from the Operating
// System.
type Window struct {
	window *glfw.Window
	scene  *Scene

	width  int
	height int
}

// NewWindow initializes the program window and OpenGL backend.
func NewWindow(width, height int, title string) (*Window, error) {
	// Create our wrapper object and retain settings
	w := &Window{}
	w.width = width
	w.height = height

	// Initialize the GLFW window/context
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	// Use glfw to create a new window
	window, err := glfw.CreateWindow(int(width), int(height), title, nil, nil)
	if err != nil {
		return nil, err
	}
	w.window = window
	w.window.MakeContextCurrent()
	// Register GLFW callbacks
	w.window.SetFramebufferSizeCallback(w.onWindowGeometryChanged)
	w.window.SetKeyCallback(w.onKeyPressed)

	// Initialize OpenGL
	gl.Init()
	w.scene = NewScene()

	return w, nil
}

// ShouldClose returns true when the window must be closed. Before it should be
// closed, it is safe to call any drawing operations. Once the window should be
// closed is flipped to true, then callers must call Close() method to ensure
// resources are properly freed.
func (w *Window) ShouldClose() bool {
	return w.window.ShouldClose()
}

// Close frees any used resources and close the underlying GLFW window.
func (w *Window) Close() {
	glfw.Terminate()
}

func (w *Window) onWindowGeometryChanged(wd *glfw.Window, width, height int) {
	w.width = width
	w.height = height
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (w *Window) onKeyPressed(wd *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	log.Infof("Key event received: key: %v, scancode: %v, action: %v, mods: %v", key, scancode, action, mods)

	if key == glfw.KeyEscape && action == glfw.Press {
		log.Infof("ESC key pressed. Exiting...")
		w.window.SetShouldClose(true)
	}

	if key == glfw.KeyF10 && action == glfw.Press {
		log.Infof("F10 key pressed. Flipping wireframe mode...")
		w.scene.wireFrames = !w.scene.wireFrames
	}
}

// PoolEvents listen to any window/input events to be passed to the input callbacks.
func (w *Window) PollEvents() {
	glfw.PollEvents()

}

// SwapBuffers will flip the drawing buffer to the visible buffer on the display.
func (w *Window) SwapBuffers() {
	w.window.SwapBuffers()
}

// Scene returns the Scene Graph used to draw on screen.
func (w *Window) Scene() *Scene {
	return w.scene
}

// shaderFile is a helper struct that represents a shader file and it's type.
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
	gl.Uniform1i(gl.GetUniformLocation(*s.program, gl.Str(name+"\x00")), value)
	return nil
}

// UniformFloat adds the provided float value as a GLSL uniform with the given
// name.  Returns an error if the program was not linked.
func (s *Shader) UniformFloat(name string, value float32) error {
	if s.program == nil {
		return errors.New("shader: invalid state: uniform called before Link()")
	}
	gl.Uniform1f(gl.GetUniformLocation(*s.program, gl.Str(name+"\x00")), value)
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

// Scene represents a graph of elements to be drawn on screen by the OpenGL
// driver.
type Scene struct {
	vao *uint32
	vbo *uint32
	ebo *uint32

	clearColor color.Color
	wireFrames bool
}

// NewScene initializes an empty scene with the proper memory allocations.
func NewScene() *Scene {
	s := &Scene{}
	s.allocateBuffers()
	return s
}

func (s *Scene) allocateBuffers() {
	if s.vao == nil {
		s.vao = new(uint32)
		s.vbo = new(uint32)
		s.ebo = new(uint32)
		gl.GenBuffers(1, s.vao)
		gl.GenBuffers(1, s.vbo)
		gl.GenBuffers(1, s.ebo)
	}
}

// AddTriangles adds the provided vertices and indices to the current scene.
func (s *Scene) AddTriangles(vertices []float32, indices []uint32) {
	s.allocateBuffers()

	gl.BindVertexArray(*s.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, *s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, *s.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// TODO(ronoaldo): 3 should be ... what???
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)

	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(0)
}

// Draw calls the underlying driver to render the scene graph on the current
// buffer.
//
// If the provided Shader program is not nil, it will be registered to be used
// before rendering anything on screen.
func (s *Scene) Draw(shader *Shader) {
	s.allocateBuffers()

	if s.clearColor == nil {
		s.clearColor = color.Black
	}
	r, g, b, a := s.clearColor.RGBA()
	gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.BindVertexArray(*s.vao)
	// TODO: use a default minimal shader program if no other shaders where specified
	// since OpenGL requires a fragment and a vertex shader at a minimum.
	if shader != nil {
		shader.Use()
	}

	if s.wireFrames {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
}
