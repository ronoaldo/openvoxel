//go:build !js

package render

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strings"
	"unsafe"

	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	glm "github.com/go-gl/mathgl/mgl32"
	"github.com/ronoaldo/openvoxel/log"
	"github.com/ronoaldo/openvoxel/transform"
)

// f is a syntax suggar to cast any number to float32
func f[X int | int32 | int64 | uint | uint32 | uint64 | float64](i X) float32 {
	return float32(i)
}

// f is a syntax suggar to cast any number to float64
func f6[X int | int32 | int64 | uint | uint32 | uint64 | float32](i X) float64 {
	return float64(i)
}

// Version returns the OpengGL version as reported by the driver.
func Version() string {
	param := func(p uint32) string {
		return gl.GoStr(gl.GetString(p))
	}
	version := "OpenGL Version " + param(gl.VERSION) +
		"; Shading Language: " + param(gl.SHADING_LANGUAGE_VERSION) +
		"; Vendor: " + param(gl.VENDOR) +
		"; Renderer: " + param(gl.RENDERER)
	return version
}

// Time returns the time in miliseconds since the window was initialized.
func Time() float64 {
	return glfw.GetTime()
}

type Camera struct {
	pos   glm.Vec3
	front glm.Vec3
	up    glm.Vec3
}

func NewCamera() (c *Camera) {
	c = &Camera{
		pos:   glm.Vec3{-20, 4, 3},
		front: glm.Vec3{0, 0, -1},
		up:    glm.Vec3{0, 1, 0},
	}
	return
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

	Width  int
	Height int

	pressedKeys  map[glfw.Key]struct{}
	firstMouse   bool
	lastX, lastY float64
	yaw, pitch   float64
	sensitivity  float64
}

// NewWindow initializes the program window and OpenGL backend.
func NewWindow(width, height int, title string) (*Window, error) {
	// Create our wrapper object and retain settings
	w := &Window{}
	w.Width = width
	w.Height = height

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
	w.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Register GLFW callbacks
	w.window.SetFramebufferSizeCallback(w.onWindowGeometryChanged)
	w.window.SetKeyCallback(w.onKeyPressed)
	w.window.SetCursorPosCallback(w.onCursorPosChange)

	// Initialize OpenGL
	gl.Init()
	w.scene = NewScene()
	w.pressedKeys = make(map[glfw.Key]struct{})
	w.sensitivity = 0.05

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
	w.Width = width
	w.Height = height
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

	if key == glfw.KeyF1 && action == glfw.Press {
		w.sensitivity = w.sensitivity + 0.1
		log.Infof("F1 key pressed, increasing sensitivity to: %v", w.sensitivity)
	}
	if key == glfw.KeyF2 && action == glfw.Press {
		w.sensitivity = w.sensitivity - 0.1
		log.Infof("F1 key pressed, decreasing sensitivity to: %v", w.sensitivity)
	}
	if w.sensitivity > 5 || w.sensitivity < 0 {
		w.sensitivity = 0.05
		log.Infof("FIX sensitivity too crazy, adjusted to: %v", w.sensitivity)
	}

	switch action {
	case glfw.Press:
		w.pressedKeys[key] = struct{}{}
	case glfw.Release:
		delete(w.pressedKeys, key)
	}

	cam := w.scene.cam
	cameraSpeed := f(0.5)

	// Movement handling
	if _, ok := w.pressedKeys[glfw.KeyW]; ok {
		cam.pos = cam.pos.Add(cam.front.Mul(cameraSpeed))
		log.Infof("Key W => Moving forward: cam=%#v", cam)
	}
	if _, ok := w.pressedKeys[glfw.KeyS]; ok {
		cam.pos = cam.pos.Sub(cam.front.Mul(cameraSpeed))
		log.Infof("Key S => Moving backward: cam=%#v", w.scene.cam)
	}
	if _, ok := w.pressedKeys[glfw.KeyA]; ok {
		cam.pos = cam.pos.Sub(
			cam.front.Cross(cam.up).Normalize().Mul(cameraSpeed),
		)
		log.Infof("Key A => Moving left: cam=%#v", w.scene.cam)
	}
	if _, ok := w.pressedKeys[glfw.KeyD]; ok {
		cam.pos = cam.pos.Add(
			cam.front.Cross(cam.up).Normalize().Mul(cameraSpeed),
		)
		log.Infof("Key D => Moving right: cam=%#v", w.scene.cam)
	}
}

func (w *Window) onCursorPosChange(wd *glfw.Window, xpos, ypos float64) {
	if w.firstMouse {
		w.lastX = xpos
		w.lastY = ypos
		w.firstMouse = false
	}

	xoffset := xpos - w.lastX
	yoffset := w.lastY - ypos
	w.lastX = xpos
	w.lastY = ypos

	xoffset *= w.sensitivity
	yoffset *= w.sensitivity

	w.yaw += xoffset
	w.pitch += yoffset

	if w.pitch > 89.0 {
		w.pitch = 89.0
	}
	if w.pitch < -89.0 {
		w.pitch = -89.0
	}

	yaw := f6(glm.DegToRad(f(w.yaw)))
	pitch := f6(glm.DegToRad(f(w.pitch)))

	direction := glm.Vec3{
		f(math.Cos(yaw) * math.Cos(pitch)),
		f(math.Sin(pitch)),
		f(math.Sin(yaw) * math.Cos(pitch)),
	}
	w.scene.cam.front = direction.Normalize()
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

// ShaderFile is a helper struct that represents a shader file and it's type.
type shaderSource struct {
	src        string
	shaderType uint32
}

// Shader abstracts raw GLSL shader compilation, linking and usage.
type Shader struct {
	shaderFiles []shaderSource
	program     *uint32
}

// VertexShader appends the provider shader file to the pipeline. This method
// returns the Shader reference to allow for chaining.
func (s *Shader) VertexShader(src string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderSource{src, gl.VERTEX_SHADER})
	return s
}

// FragmentShader appends the provided shader file to the pipeline. This method
// returns the Shader reference to allow for chaining.
func (s *Shader) FragmentShader(src string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderSource{src, gl.FRAGMENT_SHADER})
	return s
}

// UniformInt adds the provided int value as a GLSL uniform with the given
// name.  Returns an error if the shader was not linked.
func (s *Shader) UniformInts(name string, v ...int32) error {
	if s.program == nil {
		return ErrShaderNotLinked
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
		return ErrShaderNotLinked
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

func (s *Shader) UniformTransformation(name string, model glm.Mat4) error {
	if s.program == nil {
		return ErrShaderNotLinked
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
		shaderId, err := s.compileShader(file.src, file.shaderType)
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
	cam *Camera

	vao *uint32

	vbo     *uint32
	vboSize int32

	ebo     *uint32
	eboSize int32

	clearColor color.Color
	wireFrames bool

	tex *Texture
}

// NewScene initializes an empty scene with the proper memory allocations.
func NewScene() *Scene {
	s := &Scene{
		cam: NewCamera(),
	}
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

var sizeOfFloat32 = int(unsafe.Sizeof(float32(0)))

func (s *Scene) BgColor(c color.Color) {
	s.clearColor = c
}

// AddTriangles adds the provided vertices and indices to the current scene.
func (s *Scene) AddTriangles(vertices []float32, indices []uint32) {
	log.Infof("Float size: %v", sizeOfFloat32)
	s.allocateBuffers()

	gl.BindVertexArray(*s.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, *s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*sizeOfFloat32, gl.Ptr(vertices), gl.STATIC_DRAW)
	s.vboSize += int32(len(vertices))

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

func (s *Scene) AddVertices(vertices []float32) {
	s.allocateBuffers()

	gl.BindVertexArray(*s.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, *s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*sizeOfFloat32, gl.Ptr(vertices), gl.STATIC_DRAW)
	s.vboSize += int32(len(vertices)) / 5
	log.Infof("Adding vertices to scene: vboSize=%v ", s.vboSize)

	// Configure the vertex array attributes
	// [0] => positions size=3, stride=5*float, offset=0
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)

	// [1] => text coord size=2, stride=5*float, offset=3*float
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 5*4, 3*4)
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (s *Scene) AddTexture(tex *Texture) {
	s.tex = tex
}

func (s *Scene) Clear() {
	gl.Enable(gl.DEPTH_TEST)
	if s.clearColor == nil {
		s.clearColor = BgColor
	}
	r, g, b, a := s.clearColor.RGBA()
	gl.ClearColor(float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// Draw calls the underlying driver to render the scene graph on the current
// buffer.
//
// If the provided Shader program is not nil, it will be registered to be used
// before rendering anything on screen.
func (s *Scene) Draw(shader *Shader) {
	s.allocateBuffers()

	// TODO: use a default minimal shader program if no other shaders where specified
	// since OpenGL requires a fragment and a vertex shader at a minimum.
	if shader != nil {
		// Camera position changing
		view := transform.LookAt(
			s.cam.pos, s.cam.pos.Add(s.cam.front), s.cam.up,
		)
		shader.UniformTransformation("view", view)
	}

	if s.tex != nil {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, s.tex.tex)
	}

	if s.wireFrames {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	gl.BindVertexArray(*s.vao)
	if s.eboSize > 0 {
		gl.DrawElements(gl.TRIANGLES, s.eboSize, gl.UNSIGNED_INT, nil)
	} else {
		gl.DrawArrays(gl.TRIANGLES, 0, s.vboSize)
	}
	gl.BindVertexArray(0)
}

type Texture struct {
	tex    uint32
	pixels []uint8
}

func NewTexture(path string) (t *Texture, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewTextureFromBytes(b)
}

func NewTextureFromBytes(b []byte) (t *Texture, err error) {
	// TODO(ronoaldo): check for image cache and return the same texture loaded previously
	w, h, pixels, err := decodeImage(b)
	if err != nil {
		return nil, err
	}
	// Load the texture into OpenGL
	t = &Texture{
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

func decodeImage(b []byte) (w, h int, px []uint8, err error) {
	img, ftype, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return 0, 0, nil, err
	}
	img = imaging.FlipV(img)
	w, h = img.Bounds().Size().X, img.Bounds().Size().Y
	log.Infof("Loaded %v image (%dx%d) from %v bytes", ftype, w, h, len(b))

	// Create pixel data from PNG
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, 0, nil, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return rgba.Rect.Size().X, rgba.Rect.Size().Y, rgba.Pix, nil
}
