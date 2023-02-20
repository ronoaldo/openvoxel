// package render implements OpenGL rendering logic and wrappers.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"strings"
	"unsafe"

	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ronoaldo/openvoxel/log"
)

// Version returns the OpengGL version as reported by the driver.
func Version() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

// Time returns the time in miliseconds since the window was initialized.
func Time() float64 {
	return glfw.GetTime()
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
	// Use 3.3 Core Profile
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

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
func (s *Shader) VertexShaderFile(path string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderFile{path, gl.VERTEX_SHADER})
	return s
}

// FragmentShader appends the provided shader file to the pipeline. This method
// returns the Shader reference to allow for chaining.
func (s *Shader) FragmentShaderFile(path string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderFile{path, gl.FRAGMENT_SHADER})
	return s
}

var shaderNotLinkedError = errors.New("shader: invalid state: uniform called before Link()")

// UniformInt adds the provided int value as a GLSL uniform with the given
// name.  Returns an error if the shader was not linked.
func (s *Shader) UniformInts(name string, v ...int32) error {
	if s.program == nil {
		return shaderNotLinkedError
	}
	loc := gl.GetUniformLocation(*s.program, gl.Str(name+"\x00"))
	switch len(v) {
	case 1:
		gl.Uniform1i(loc, v[0])
	case 2:
		gl.Uniform2i(loc, v[0], v[1])
	case 3:
		gl.Uniform3i(loc, v[0], v[1], v[2])
	case 4:
		gl.Uniform4i(loc, v[0], v[1], v[2], v[3])
	default:
		return fmt.Errorf("invalid argument count: %v, expected up to 4 values", len(v))
	}
	return nil
}

// UniformFloat adds the provided float value as a GLSL uniform with the given
// name.  Returns an error if the program was not linked.
func (s *Shader) UniformFloats(name string, v ...float32) error {
	if s.program == nil {
		return shaderNotLinkedError
	}
	loc := gl.GetUniformLocation(*s.program, gl.Str(name+"\x00"))
	switch len(v) {
	case 1:
		gl.Uniform1f(loc, v[0])
	case 2:
		gl.Uniform2f(loc, v[0], v[1])
	case 3:
		gl.Uniform3f(loc, v[0], v[1], v[2])
	case 4:
		gl.Uniform4f(loc, v[0], v[1], v[2], v[3])
	default:
		return fmt.Errorf("invalid argument count: %v, expected up to 4 values", len(v))
	}
	return nil
}

func (s *Shader) UniformTransformation(name string, model mgl32.Mat4) error {
	if s.program == nil {
		return shaderNotLinkedError
	}

	modelUniform := gl.GetUniformLocation(*s.program, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
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

	ebo     *uint32
	eboSize int32

	clearColor color.Color
	wireFrames bool

	tex *Texture
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
		gl.GenVertexArrays(1, s.vao)
		gl.GenBuffers(1, s.vbo)
		gl.GenBuffers(1, s.ebo)
	}
}

// 4
var sizeOfFloat32 = int(unsafe.Sizeof(float32(0)))

// AddTriangles adds the provided vertices and indices to the current scene.
func (s *Scene) AddTriangles(vertices []float32, indices []uint32) {
	log.Infof("Float size: %v", sizeOfFloat32)
	s.allocateBuffers()

	gl.BindVertexArray(*s.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, *s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*sizeOfFloat32, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, *s.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*sizeOfFloat32, gl.Ptr(indices), gl.STATIC_DRAW)
	s.eboSize += int32(len(indices))

	// Configure the vertex array attributes
	// [0] => positions size=3, stride=8*float, offset=0
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, nil)
	gl.EnableVertexAttribArray(0)
	// [1] => color     size=3,  stride=8*float, offset=3*float
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, 8*4, 3*4)
	gl.EnableVertexAttribArray(1)
	// [2] => text coord size=2, stride=8*float, offset=6*float
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, 8*4, 6*4)
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)
}

func (s *Scene) AddTexture(tex *Texture) {
	s.tex = tex
}

// Draw calls the underlying driver to render the scene graph on the current
// buffer.
//
// If the provided Shader program is not nil, it will be registered to be used
// before rendering anything on screen.
func (s *Scene) Draw(shader *Shader) {
	s.allocateBuffers()

	// Clear
	if s.clearColor == nil {
		s.clearColor = color.Black
	}
	r, g, b, a := s.clearColor.RGBA()
	gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// TODO: use a default minimal shader program if no other shaders where specified
	// since OpenGL requires a fragment and a vertex shader at a minimum.
	if shader != nil {
		shader.Use()
	}

	gl.BindVertexArray(*s.vao)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, s.tex.tex)

	if s.wireFrames {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	gl.DrawElements(gl.TRIANGLES, s.eboSize, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

type Texture struct {
	tex    uint32
	path   string
	pixels []uint8
}

func NewTexture(path string) (t *Texture, err error) {
	// TODO(ronoaldo): check for image cache and return the same texture loaded previously
	w, h, pixels, err := readImage(path)
	if err != nil {
		return nil, err
	}
	// Load the texture into OpenGL
	t = &Texture{
		path:   path,
		pixels: pixels,
	}
	gl.GenTextures(1, &t.tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(w),
		int32(h),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(pixels))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return t, nil
}

func readImage(path string) (w, h int, px []uint8, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, 0, nil, err
	}

	img, ftype, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return 0, 0, nil, err
	}
	img = imaging.FlipV(img)
	w, h = img.Bounds().Size().X, img.Bounds().Size().Y
	log.Infof("Loaded %v image (%dx%d) from file %v", ftype, w, h, path)

	// Create pixel data from PNG
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, 0, nil, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return rgba.Rect.Size().X, rgba.Rect.Size().Y, rgba.Pix, nil
}
