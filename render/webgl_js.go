package render

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"syscall/js"
	"time"

	glm "github.com/go-gl/mathgl/mgl32"
	"github.com/ronoaldo/openvoxel/log"

	"github.com/disintegration/imaging"
)

func Version() string {
	if gl.IsUndefined() {
		return "WebGL Not Initialized"
	}

	params := []string{"VERSION", "SHADING_LANGUAGE_VERSION", "VENDOR", "RENDERER"}
	version := ""
	for _, p := range params {
		version = version + "\n" + gl.Call("getParameter", gl.Get(p).Int()).String()
	}
	return version
}

var initializedAt time.Time

func Time() float64 {
	return float64(time.Since(initializedAt) / time.Millisecond)
}

type Window struct {
	canvas js.Value
	scene  *Scene

	width  int
	height int
}

var document js.Value
var gl js.Value

func NewWindow(width, height int, title string) (w *Window, err error) {
	w = &Window{}
	w.width = width
	w.height = height

	document = js.Global().Get("document")
	document.Set("title", title)
	w.canvas = document.Call("createElement", "canvas")
	document.Get("body").Call("appendChild", w.canvas)
	w.canvas.Set("width", width)
	w.canvas.Set("height", height)

	// TODO(ronoaldo) error check
	gl = w.canvas.Call("getContext", "webgl2")
	w.scene = NewScene()
	return w, nil
}

func (w *Window) ShouldClose() bool {
	return false
}

func (w *Window) Close() {}

func (w *Window) PollEvents() {}

func (w *Window) SwapBuffers() {}

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
	if s.program.IsNull() {
		return errShaderNotLinked
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
	if s.program.IsNull() {
		return errShaderNotLinked
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
	if s.program.IsNull() {
		return errShaderNotLinked
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
	if s.program.IsNull() {
		panic("shader program not linked; call Shader.Link() first")
	}
	gl.Call("useProgram", s.program)
}

type Scene struct {
	tex *Texture
}

func NewScene() *Scene {
	return &Scene{}
}

func (s *Scene) AddTriangles(vertices []float32, indices []float32) {}

func (s *Scene) AddVertices(vertices []float32) {}

func (s *Scene) AddTexture(tex *Texture) {
	s.tex = tex
}

func (s *Scene) Clear() {}

func (s *Scene) Draw(shader *Shader) {}

type Texture struct {
	tex    js.Value
	pixels []uint8
}

func NewTexture(path string) (t *Texture, err error) {
	return nil, errNotImplemented
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
