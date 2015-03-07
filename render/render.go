package render

import (
	"github.com/thinkofdeath/goquake/bsp"
	"github.com/thinkofdeath/goquake/pak"
	"github.com/thinkofdeath/goquake/render/gl"
	"github.com/thinkofdeath/goquake/vmath"
	"io/ioutil"
	"math"
	"time"
)

// TODO(Think) Clean this all up
// Its basically a straight port from my renderer
// in dart

const (
	atlasSize = 1024
)

var (
	currentMap *qMap
	pakFile    *pak.File

	perspectiveMatrix = vmath.NewMatrix4()
	cameraMatrix      = vmath.NewMatrix4()
	lastScreenWidth   = -1 // Used for checking if the perspective matrix needs updating
	lastScreenHeight  = -1

	colourMap    gl.Texture
	palette      gl.Texture
	texture      gl.Texture
	textureLight gl.Texture

	gameProgram       gl.Program
	gameAPosition     gl.Attribute
	gameALight        gl.Attribute
	gameATex          gl.Attribute
	gameATexInfo      gl.Attribute
	gameALightInfo    gl.Attribute
	gameALightType    gl.Attribute
	gameUPMat         gl.Uniform
	gameUUMat         gl.Uniform
	gameUColourMap    gl.Uniform
	gameUPalette      gl.Uniform
	gameUTexture      gl.Uniform
	gameUTextureLight gl.Uniform
	gameULightStyles  gl.Uniform

	cameraX       float64 = 504
	cameraY       float64 = 401
	cameraZ       float64 = 75
	cameraRotY    float64
	cameraRotX    float64 = math.Pi
	movingForward bool
)

func Init(p *pak.File, initialMap *bsp.File) {
	gl.Init()

	pakFile = p

	// Load textures
	cm, _ := ioutil.ReadAll(pakFile.Reader("gfx/colormap.lmp"))
	colourMap = createTexture(glTexture{
		Data: cm,
		Width: 256, Height: 64,
		Format: gl.Luminance,
	})

	pm, _ := ioutil.ReadAll(pakFile.Reader("gfx/palette.lmp"))
	palette = createTexture(glTexture{
		Data: pm,
		Width: 16, Height: 16,
		Format: gl.RGB,
	})

	dummy := make([]byte, atlasSize*atlasSize)

	texture = createTexture(glTexture{
		Data: dummy,
		Width: atlasSize, Height: atlasSize,
		Format: gl.Luminance,
	})
	textureLight = createTexture(glTexture{
		Data: dummy,
		Width: atlasSize, Height: atlasSize,
		Format: gl.Luminance,
		Filter: gl.Linear,
	})

	gameProgram = compileProgram(gameVertexSource, gameFragmentSource)
	gameAPosition = gameProgram.AttributeLocation("a_Position")
	gameALight = gameProgram.AttributeLocation("a_light")
	gameATex = gameProgram.AttributeLocation("a_tex")
	gameATexInfo = gameProgram.AttributeLocation("a_texInfo")
	gameALightInfo = gameProgram.AttributeLocation("a_lightInfo")
	gameALightType = gameProgram.AttributeLocation("a_lightType")
	gameUPMat = gameProgram.UniformLocation("pMat")
	gameUUMat = gameProgram.UniformLocation("uMat")
	gameUColourMap = gameProgram.UniformLocation("colourMap")
	gameUPalette = gameProgram.UniformLocation("palette")
	gameUTexture = gameProgram.UniformLocation("texture")
	gameUTextureLight = gameProgram.UniformLocation("textureLight")
	gameULightStyles = gameProgram.UniformLocation("lightStyles")

	currentMap = newQMap(initialMap)
}

var lastFrame = time.Now()

func Draw(width, height int) {
	now := time.Now()
	delta := float64(now.Sub(lastFrame).Nanoseconds()) / float64(time.Second/60)
	lastFrame = now

	if width != lastScreenWidth || height != lastScreenHeight {
		lastScreenWidth = width
		lastScreenHeight = height

		perspectiveMatrix.Identity()
		perspectiveMatrix.Perspective(
			(math.Pi/180)*75,
			float32(width)/float32(height),
			0.1,
			10000.0,
		)
	}

	gl.Viewport(0, 0, width, height)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.ColorBufferBit | gl.DepthBufferBit)

	if movingForward {
		cameraX += 5.0 * math.Sin(cameraRotY) * delta
		cameraY += 5.0 * math.Cos(cameraRotY) * delta
		cameraZ -= 5.0 * math.Sin(-cameraRotX) * delta
	}

	cameraMatrix.Identity()
	cameraMatrix.Translate(-float32(cameraX), -float32(cameraY), -float32(cameraZ))
	cameraMatrix.RotateZ(float32(-cameraRotY))
	cameraMatrix.RotateX(float32(-cameraRotX - (math.Pi / 2.0)))

	gameProgram.Use()
	gameUPMat.Matrix4(false, perspectiveMatrix)
	gameUUMat.Matrix4(false, cameraMatrix)

	// Bind textures

	gl.ActiveTexture(0)
	palette.Bind(gl.Texture2D)
	gameUPalette.Int(0)

	gl.ActiveTexture(1)
	colourMap.Bind(gl.Texture2D)
	gameUColourMap.Int(1)

	gl.ActiveTexture(2)
	texture.Bind(gl.Texture2D)
	gameUTexture.Int(2)

	gl.ActiveTexture(3)
	textureLight.Bind(gl.Texture2D)
	gameUTextureLight.Int(3)

	// Setup and render

	gl.Enable(gl.DepthTest)
	gl.Enable(gl.CullFaceFlag)
	gl.CullFace(gl.Back)
	gl.FrontFace(gl.CounterClockWise)

	gameAPosition.Enable()
	gameALight.Enable()
	gameATex.Enable()
	gameATexInfo.Enable()
	gameALightInfo.Enable()
	gameALightType.Enable()

	currentMap.render()

	gameAPosition.Disable()
	gameALight.Disable()
	gameATex.Disable()
	gameATexInfo.Disable()
	gameALightInfo.Disable()
	gameALightType.Disable()

	gl.Disable(gl.CullFaceFlag)
	gl.Disable(gl.DepthTest)

	gl.Flush()
}

func MoveForward() {
	movingForward = true
}

func StopMove() {
	movingForward = false
}

func Rotate(x, y float64) {
	cameraRotX += y
	cameraRotY += x
}
