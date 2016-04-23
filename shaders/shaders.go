package shaders // import "engo.io/engo/shaders"

import (
	"fmt"

	internalcamera "engo.io/engo/internal/camera"
	internalwindow "engo.io/engo/internal/window"
	"engo.io/engo/message"
	"engo.io/engo/window"
	"engo.io/gl"
)

const bufferSize = 10000

type Shader interface {
	Initialize(width, height float32)
	Pre()
	Draw(texture *gl.Texture, buffer *gl.Buffer, x, y, rotation float32)
	Post()
}

func LoadShader(vertSrc, fragSrc string) *gl.Program {
	vertShader := internalwindow.Gl.CreateShader(internalwindow.Gl.VERTEX_SHADER)
	internalwindow.Gl.ShaderSource(vertShader, vertSrc)
	internalwindow.Gl.CompileShader(vertShader)
	defer internalwindow.Gl.DeleteShader(vertShader)

	fragShader := internalwindow.Gl.CreateShader(internalwindow.Gl.FRAGMENT_SHADER)
	internalwindow.Gl.ShaderSource(fragShader, fragSrc)
	internalwindow.Gl.CompileShader(fragShader)
	defer internalwindow.Gl.DeleteShader(fragShader)

	program := internalwindow.Gl.CreateProgram()
	internalwindow.Gl.AttachShader(program, vertShader)
	internalwindow.Gl.AttachShader(program, fragShader)
	internalwindow.Gl.LinkProgram(program)

	return program
}

type defaultShader struct {
	indices  []uint16
	indexVBO *gl.Buffer
	program  *gl.Program

	projX float32
	projY float32

	lastTexture *gl.Texture

	inPosition   int
	inTexCoords  int
	inColor      int
	ufCamera     *gl.UniformLocation
	ufPosition   *gl.UniformLocation
	ufProjection *gl.UniformLocation
}

func (s *defaultShader) Initialize(width, height float32) {
	s.program = LoadShader(`
#version 120

attribute vec2 in_Position;
attribute vec2 in_TexCoords;
attribute vec4 in_Color;

uniform vec2 uf_Position;
uniform vec3 uf_Camera;
uniform vec2 uf_Projection;

varying vec4 var_Color;
varying vec2 var_TexCoords;

void main() {
  var_Color = in_Color;
  var_TexCoords = in_TexCoords;

  gl_Position = vec4((in_Position.x + uf_Position.x - uf_Camera.x)/  uf_Projection.x,
  					 (in_Position.y + uf_Position.y - uf_Camera.y)/ -uf_Projection.y,
  					 0.0, uf_Camera.z);

}`, `
/* Fragment Shader */
#ifdef GL_ES
#define LOWP lowp
precision mediump float;
#else
#define LOWP
#endif

varying vec4 var_Color;
varying vec2 var_TexCoords;

uniform sampler2D uf_Texture;

void main (void) {
  gl_FragColor = var_Color * texture2D(uf_Texture, var_TexCoords);
}`)

	// Create and populate indices buffer
	s.indices = make([]uint16, 6*bufferSize)
	for i, j := 0, 0; i < bufferSize*6; i, j = i+6, j+4 {
		s.indices[i+0] = uint16(j + 0)
		s.indices[i+1] = uint16(j + 1)
		s.indices[i+2] = uint16(j + 2)
		s.indices[i+3] = uint16(j + 0)
		s.indices[i+4] = uint16(j + 2)
		s.indices[i+5] = uint16(j + 3)
	}
	s.indexVBO = internalwindow.Gl.CreateBuffer()
	internalwindow.Gl.BindBuffer(internalwindow.Gl.ELEMENT_ARRAY_BUFFER, s.indexVBO)
	internalwindow.Gl.BufferData(internalwindow.Gl.ELEMENT_ARRAY_BUFFER, s.indices, internalwindow.Gl.STATIC_DRAW)

	s.SetProjection(width, height)

	// Define things that should be read from the texture buffer
	s.inPosition = internalwindow.Gl.GetAttribLocation(s.program, "in_Position")
	s.inTexCoords = internalwindow.Gl.GetAttribLocation(s.program, "in_TexCoords")
	s.inColor = internalwindow.Gl.GetAttribLocation(s.program, "in_Color")

	// Define things that should be set per draw
	s.ufCamera = internalwindow.Gl.GetUniformLocation(s.program, "uf_Camera")
	s.ufPosition = internalwindow.Gl.GetUniformLocation(s.program, "uf_Position")
	s.ufProjection = internalwindow.Gl.GetUniformLocation(s.program, "uf_Projection")

	// Enable those things
	internalwindow.Gl.EnableVertexAttribArray(s.inPosition)
	internalwindow.Gl.EnableVertexAttribArray(s.inTexCoords)
	internalwindow.Gl.EnableVertexAttribArray(s.inColor)

	internalwindow.Gl.Enable(internalwindow.Gl.BLEND)
	internalwindow.Gl.BlendFunc(internalwindow.Gl.SRC_ALPHA, internalwindow.Gl.ONE_MINUS_SRC_ALPHA)

	// We sometimes have to change our projection matrix
	message.Mailbox.Listen("WindowResizeMessage", func(m message.Message) {
		wrm, ok := m.(window.WindowResizeMessage)
		if !ok {
			return
		}

		if !internalwindow.ScaleOnResize {
			s.SetProjection(float32(wrm.NewWidth), float32(wrm.NewHeight))
		}
	})
}

func (s *defaultShader) Pre() {
	internalwindow.Gl.UseProgram(s.program)
	internalwindow.Gl.Uniform2f(s.ufProjection, s.projX, s.projY)
	internalwindow.Gl.Uniform3f(s.ufCamera, internalcamera.Cam.X(), internalcamera.Cam.Y(), internalcamera.Cam.Z())
}

func (s *defaultShader) Draw(texture *gl.Texture, buffer *gl.Buffer, x, y, rotation float32) {
	if s.lastTexture != texture {
		internalwindow.Gl.BindTexture(internalwindow.Gl.TEXTURE_2D, texture)
		internalwindow.Gl.BindBuffer(internalwindow.Gl.ARRAY_BUFFER, buffer)

		internalwindow.Gl.VertexAttribPointer(s.inPosition, 2, internalwindow.Gl.FLOAT, false, 20, 0)
		internalwindow.Gl.VertexAttribPointer(s.inTexCoords, 2, internalwindow.Gl.FLOAT, false, 20, 8)
		internalwindow.Gl.VertexAttribPointer(s.inColor, 4, internalwindow.Gl.UNSIGNED_BYTE, true, 20, 16)

		s.lastTexture = texture
	}

	// TODO: add rotation
	internalwindow.Gl.Uniform2f(s.ufPosition, x, y)
	internalwindow.Gl.DrawElements(internalwindow.Gl.TRIANGLES, 6, internalwindow.Gl.UNSIGNED_SHORT, 0)
}

func (s *defaultShader) Post() {
	s.lastTexture = nil
}

func (s *defaultShader) SetProjection(width, height float32) {
	s.projX = width / 2
	s.projY = height / 2
}

type hudShader struct {
	indices  []uint16
	indexVBO *gl.Buffer
	program  *gl.Program

	projX float32
	projY float32

	lastTexture *gl.Texture

	inPosition   int
	inTexCoords  int
	inColor      int
	ufPosition   *gl.UniformLocation
	ufProjection *gl.UniformLocation
}

func (s *hudShader) Initialize(width, height float32) {
	s.program = LoadShader(`
#version 120

attribute vec2 in_Position;
attribute vec2 in_TexCoords;
attribute vec4 in_Color;

uniform vec2 uf_Position;
uniform vec2 uf_Projection;

varying vec4 var_Color;
varying vec2 var_TexCoords;

void main() {
  var_Color = in_Color;
  var_TexCoords = in_TexCoords;

  gl_Position = vec4((in_Position.x + uf_Position.x)/  uf_Projection.x - 1.0,
  					 (in_Position.y + uf_Position.y)/ -uf_Projection.y + 1.0,
  					 0.0, 1.0);

}`, `
#ifdef GL_ES
#define LOWP lowp
precision mediump float;
#else
#define LOWP
#endif

varying vec4 var_Color;
varying vec2 var_TexCoords;

uniform sampler2D uf_Texture;

void main (void) {
  gl_FragColor = var_Color * texture2D(uf_Texture, var_TexCoords);
}`)

	// Create and populate indices buffer
	s.indices = make([]uint16, 6*bufferSize)
	for i, j := 0, 0; i < bufferSize*6; i, j = i+6, j+4 {
		s.indices[i+0] = uint16(j + 0)
		s.indices[i+1] = uint16(j + 1)
		s.indices[i+2] = uint16(j + 2)
		s.indices[i+3] = uint16(j + 0)
		s.indices[i+4] = uint16(j + 2)
		s.indices[i+5] = uint16(j + 3)
	}
	s.indexVBO = internalwindow.Gl.CreateBuffer()
	internalwindow.Gl.BindBuffer(internalwindow.Gl.ELEMENT_ARRAY_BUFFER, s.indexVBO)
	internalwindow.Gl.BufferData(internalwindow.Gl.ELEMENT_ARRAY_BUFFER, s.indices, internalwindow.Gl.STATIC_DRAW)

	s.SetProjection(width, height)

	// Define things that should be read from the texture buffer
	s.inPosition = internalwindow.Gl.GetAttribLocation(s.program, "in_Position")
	s.inTexCoords = internalwindow.Gl.GetAttribLocation(s.program, "in_TexCoords")
	s.inColor = internalwindow.Gl.GetAttribLocation(s.program, "in_Color")

	// Define things that should be set per draw
	s.ufPosition = internalwindow.Gl.GetUniformLocation(s.program, "uf_Position")
	s.ufProjection = internalwindow.Gl.GetUniformLocation(s.program, "uf_Projection")

	// Enable those things
	internalwindow.Gl.EnableVertexAttribArray(s.inPosition)
	internalwindow.Gl.EnableVertexAttribArray(s.inTexCoords)
	internalwindow.Gl.EnableVertexAttribArray(s.inColor)

	internalwindow.Gl.Enable(internalwindow.Gl.BLEND)
	internalwindow.Gl.BlendFunc(internalwindow.Gl.SRC_ALPHA, internalwindow.Gl.ONE_MINUS_SRC_ALPHA)

	// TODO: listen for Projection changes
}

func (s *hudShader) Pre() {
	internalwindow.Gl.UseProgram(s.program)
	internalwindow.Gl.Uniform2f(s.ufProjection, s.projX, s.projY)
}

func (s *hudShader) Draw(texture *gl.Texture, buffer *gl.Buffer, x, y, rotation float32) {
	if s.lastTexture != texture {
		internalwindow.Gl.BindTexture(internalwindow.Gl.TEXTURE_2D, texture)
		internalwindow.Gl.BindBuffer(internalwindow.Gl.ARRAY_BUFFER, buffer)

		internalwindow.Gl.VertexAttribPointer(s.inPosition, 2, internalwindow.Gl.FLOAT, false, 20, 0)
		internalwindow.Gl.VertexAttribPointer(s.inTexCoords, 2, internalwindow.Gl.FLOAT, false, 20, 8)
		internalwindow.Gl.VertexAttribPointer(s.inColor, 4, internalwindow.Gl.UNSIGNED_BYTE, true, 20, 16)

		s.lastTexture = texture
	}

	internalwindow.Gl.Uniform2f(s.ufPosition, x, y)
	internalwindow.Gl.DrawElements(internalwindow.Gl.TRIANGLES, 6, internalwindow.Gl.UNSIGNED_SHORT, 0)
}

func (s *hudShader) Post() {
	s.lastTexture = nil
}

func (s *hudShader) SetProjection(width, height float32) {
	s.projX = width / 2
	s.projY = height / 2
}

var (
	DefaultShader = &defaultShader{}
	HUDShader     = &hudShader{}
	shadersSet    bool
)

// TODO(u): Unexport InitShaders.
func InitShaders(width, height float32) {
	if !shadersSet {
		fmt.Println("Initialized shaders", width, height)
		DefaultShader.Initialize(width, height)
		HUDShader.Initialize(width, height)

		shadersSet = true
	}
}
