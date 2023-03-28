package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"syscall/js"
	"time"

	glm "github.com/go-gl/mathgl/mgl32"
	"github.com/ronoaldo/openvoxel/log"

	"github.com/disintegration/imaging"
)

// Version returns the WebGL version as reported by the driver.
func Version() string {
	if gl.IsUndefined() || gl.IsNull() {
		return "WebGL Not Initialized."
	}

	param := func(p string) string {
		return gl.Call("getParameter", gl.Get(p).Int()).String()
	}

	version := param("VERSION") +
		"; Shading Language: " + param("SHADING_LANGUAGE_VERSION") +
		"; Vendor: " + param("VENDOR") +
		"; Renderer: " + param("RENDERER")
	return version
}

var initializedAt time.Time = time.Now()

// Time returns the time in miliseconds since the window was initialized.
func Time() float64 {
	ellapsedMills := float64(time.Since(initializedAt) / time.Millisecond)
	return ellapsedMills / 1000.0
}

type Window struct {
	canvas js.Value
	scene  *Scene

	Width  int
	Height int
}

var document js.Value
var gl js.Value

func NewWindow(width, height int, title string) (w *Window, err error) {
	w = &Window{}
	w.Width = width
	w.Height = height

	document = js.Global().Get("document")
	document.Set("title", title)
	w.canvas = document.Call("createElement", "canvas")
	document.Get("body").Call("appendChild", w.canvas)
	w.canvas.Set("width", width)
	w.canvas.Set("height", height)

	// TODO(ronoaldo) error check
	gl = w.canvas.Call("getContext", "webgl2")
	w.scene = NewScene()

	requestAnimationFrame()

	return w, nil
}

func (w *Window) ShouldClose() bool {
	return false
}

func (w *Window) Close() {}

func (w *Window) PollEvents() {}

var animationFrameLock = make(chan struct{}, 1)

func requestAnimationFrame() {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) any {
		animationFrameLock <- struct{}{}
		cb.Release()
		return nil
	})
	js.Global().Call("requestAnimationFrame", cb)
}

func (w *Window) SwapBuffers() {
	<-animationFrameLock
	requestAnimationFrame()
}

func (w *Window) Scene() *Scene {
	return w.scene
}

type shaderSource struct {
	src        string
	shaderType int
}

type Shader struct {
	shaderFiles []shaderSource
	program     js.Value
}

func (s *Shader) VertexShader(src string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderSource{src, gl.Get("VERTEX_SHADER").Int()})
	return s
}

func (s *Shader) FragmentShader(src string) *Shader {
	s.shaderFiles = append(s.shaderFiles, shaderSource{src, gl.Get("FRAGMENT_SHADER").Int()})
	return s
}

func (s *Shader) UniformInts(name string, v ...int32) error {
	if s.program.IsNull() || s.program.IsUndefined() {
		return ErrShaderNotLinked
	}
	loc := gl.Call("getUniformLocation", s.program, name)
	switch len(v) {
	case 1:
		gl.Call("uniform1i", loc, v[0])
	case 2:
		gl.Call("uniform2i", loc, v[0], v[1])
	case 3:
		gl.Call("uniform3i", loc, v[0], v[1], v[2])
	case 4:
		gl.Call("uniform4i", loc, v[0], v[1], v[2], v[3])
	default:
		return fmt.Errorf("shader: invalid argument count: %v, expected up to 4 values", len(v))
	}
	return nil
}

func (s *Shader) UniformFloats(name string, v ...float32) error {
	if s.program.IsNull() || s.program.IsUndefined() {
		return ErrShaderNotLinked
	}
	loc := gl.Call("getUniformLocation", s.program, name)
	switch len(v) {
	case 1:
		gl.Call("uniform1f", loc, v[0])
	case 2:
		gl.Call("uniform2f", loc, v[0], v[1])
	case 3:
		gl.Call("uniform3f", loc, v[0], v[1], v[2])
	case 4:
		gl.Call("uniform4f", loc, v[0], v[1], v[2], v[3])
	default:
		return fmt.Errorf("shader: invalid argument count: %v, expected up to 4 values", len(v))
	}
	return nil
}

func (s *Shader) UniformTransformation(name string, t glm.Mat4) error {
	if s.program.IsNull() || s.program.IsUndefined() {
		return ErrShaderNotLinked
	}

	loc := gl.Call("getUniformLocation", s.program, name)

	// TODO(ronoaldo): refactor out this conversion
	m := [16]float32{}
	m[0], m[1], m[2], m[3] = t.Col(0).Elem()
	m[4], m[5], m[6], m[7] = t.Col(1).Elem()
	m[8], m[9], m[10], m[11] = t.Col(2).Elem()
	m[12], m[13], m[14], m[15] = t.Col(3).Elem()
	mat4 := js.Global().Get("Float32Array").Call("of",
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12], m[13], m[14], m[15],
	)
	gl.Call("uniformMatrix4fv", loc, false, mat4)
	return nil
}

func (s *Shader) Link() error {
	shaders := []js.Value{}
	for _, file := range s.shaderFiles {
		shader, err := s.compileShader(file.src, file.shaderType)
		if err != nil {
			return err
		}

		shaders = append(shaders, shader)
	}

	p, err := s.linkProgram(shaders...)
	if err != nil {
		return err
	}

	s.program = p
	return nil
}

func (s *Shader) compileShader(shaderSource string, shaderType int) (js.Value, error) {
	log.Infof("Compiling shader (type=%v): %#s", shaderType, shaderSource)

	shader := gl.Call("createShader", shaderType)
	gl.Call("shaderSource", shader, shaderSource)
	gl.Call("compileShader", shader)
	status := gl.Call("getShaderParameter", shader, gl.Get("COMPILE_STATUS").Int())
	if !status.Bool() {
		reason := gl.Call("getShaderInfoLog", shader).String()
		log.Warnf("Error compiling shader: %v", reason)
		return js.Undefined(), fmt.Errorf("webgl: " + reason)
	}
	log.Infof("Shader compiled (status=%v)", status)
	return shader, nil
}

func (s *Shader) linkProgram(shaders ...js.Value) (js.Value, error) {
	shaderProgram := gl.Call("createProgram")
	log.Infof("Linking shaders into program %v", shaderProgram)

	for _, shader := range shaders {
		gl.Call("attachShader", shaderProgram, shader)
	}
	gl.Call("linkProgram", shaderProgram)

	status := gl.Call("getProgramParameter", shaderProgram, gl.Get("LINK_STATUS").Int())
	log.Infof("Program linked (status=%v)", status)
	return shaderProgram, nil
}

func (s *Shader) Use() {
	if s.program.IsNull() || s.program.IsUndefined() {
		panic("shader program not linked; call Shader.Link() first")
	}
	gl.Call("useProgram", s.program)
}

// Scene represents a graph of elements to be drawn on screen by the WebGL
// driver.
type Scene struct {
	tex        *Texture
	clearColor color.Color
	wireFrames bool

	vao js.Value

	vbo     js.Value
	vboSize int
}

// NewScene initializes an empty scene with the proper memory allocations.
func NewScene() *Scene {
	return &Scene{}
}

func (s *Scene) allocateBuffers() {
	if s.vao.IsNull() || s.vao.IsUndefined() {
		log.Infof("Allocating buffers ...")
		s.vao = gl.Call("createVertexArray")
		s.vbo = gl.Call("createBuffer")
	}
}

// AddTriangles adds the provided vertices and indices to the current scene.
func (s *Scene) AddTriangles(vertices []float32, indices []float32) {
}

// AddVertices adds the provided vertices array to the scene.  The vertices
// array is expected to be have 5 elements per vertice, where the first three
// elements represent the x,y,z coordinate and the other two vertices represent
// the texture coordinate for it.
func (s *Scene) AddVertices(vertices []float32) {
	s.allocateBuffers()

	ARRAY_BUFFER := gl.Get("ARRAY_BUFFER").Int()
	STATIC_DRAW := gl.Get("STATIC_DRAW").Int()
	GLFLOAT := gl.Get("FLOAT")

	gl.Call("bindVertexArray", s.vao)

	gl.Call("bindBuffer", ARRAY_BUFFER, s.vbo)

	v := toFloat32Array(vertices)
	s.vboSize += len(vertices) / 5
	log.Infof("s.vboSize %d/%d [%d bytes/item]", len(vertices), v.Length(), v.Get("BYTES_PER_ELEMENT").Int())
	gl.Call("bufferData", ARRAY_BUFFER, v, STATIC_DRAW)

	gl.Call("vertexAttribPointer", 0, 3, GLFLOAT, false, 5*4, 0)
	gl.Call("enableVertexAttribArray", 0)

	gl.Call("vertexAttribPointer", 1, 2, GLFLOAT, false, 5*4, 3*4)
	gl.Call("enableVertexAttribArray", 1)

	gl.Call("bindVertexArray", nil)
}

func toFloat32Array(in []float32) (out js.Value) {
	out = js.Global().Get("Float32Array").New(len(in))
	for k, v := range in {
		out.SetIndex(k, v)
	}
	return
}

func (s *Scene) AddTexture(tex *Texture) {
	s.tex = tex
}

func (s *Scene) Clear() {
	gl.Call("enable", gl.Get("DEPTH_TEST").Int())
	if s.clearColor == nil {
		s.clearColor = BgColor
	}
	r, g, b, a := s.clearColor.RGBA()
	gl.Call("clearColor", float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff)
	gl.Call("clear", gl.Get("COLOR_BUFFER_BIT").Int()|gl.Get("DEPTH_BUFFER_BIT").Int())
}

func (s *Scene) Draw(shader *Shader) {
	s.allocateBuffers()

	if shader != nil {
		shader.Use()
	}

	if s.tex != nil {
		gl.Call("activeTexture", gl.Get("TEXTURE0").Int())
		gl.Call("bindTexture", gl.Get("TEXTURE_2D").Int(), s.tex.tex)
	}

	gl.Call("bindVertexArray", s.vao)
	gl.Call("drawArrays", gl.Get("TRIANGLES").Int(), 0, s.vboSize)
	gl.Call("bindVertexArray", nil)
}

type Texture struct {
	tex    js.Value
	pixels []uint8
}

func NewTexture(path string) (t *Texture, err error) {
	return nil, ErrNotImplemented
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
	t.tex = gl.Call("createTexture")
	gl.Call("activeTexture", gl.Get("TEXTURE0").Int())
	gl.Call("bindTexture", gl.Get("TEXTURE_2D").Int(), t.tex)

	gl.Call("texParameteri", gl.Get("TEXTURE_2D").Int(),
		gl.Get("TEXTURE_WRAP_S").Int(), gl.Get("REPEAT").Int())
	gl.Call("texParameteri", gl.Get("TEXTURE_2D").Int(),
		gl.Get("TEXTURE_WRAP_T").Int(), gl.Get("REPEAT").Int())
	gl.Call("texParameteri", gl.Get("TEXTURE_2D").Int(),
		gl.Get("TEXTURE_MIN_FILTER").Int(), gl.Get("NEAREST").Int())
	gl.Call("texParameteri", gl.Get("TEXTURE_2D").Int(),
		gl.Get("TEXTURE_MAG_FILTER").Int(), gl.Get("NEAREST").Int())

	jsPix := js.Global().Call("eval", fmt.Sprintf("new Uint8Array(%d)", len(pixels)))
	log.Infof("js.CopyBytesToJS copied %d/%d bytes", js.CopyBytesToJS(jsPix, pixels), len(pixels))
	gl.Call("texImage2D",
		gl.Get("TEXTURE_2D").Int(),
		0,
		gl.Get("RGBA").Int(),
		int32(w),
		int32(h),
		0,
		gl.Get("RGBA").Int(),
		gl.Get("UNSIGNED_BYTE").Int(),
		jsPix)
	gl.Call("generateMipmap", gl.Get("TEXTURE_2D").Int())
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
