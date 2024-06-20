package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	_ "embed"
)

//go:embed Lato-Bold.ttf
var latoBold []byte

//go:embed CantinaBand3.ogg
var soundData []byte

type JoystickDisplay struct {
	x, y       int32
	radius     int32
	joyX, joyY float32
	pressed    bool
}

type TextDisplay struct {
	x, y    int32
	color   color.NRGBA
	font    *ttf.Font
	text    string
	texture *sdl.Texture
	surface *sdl.Surface
}

func (t *TextDisplay) SetText(text string) error {
	t.Close()
	t.text = text
	surface, err := t.font.RenderUTF8Blended(t.text, sdl.Color(t.color))
	if err != nil {
		log.Printf("Cannot set text: %s", err)
		return err
	}
	t.surface = surface
	return nil
}

func (t *TextDisplay) Render(renderer *sdl.Renderer) {
	if t.surface == nil {
		return
	}
	if t.texture == nil {
		texture, err := renderer.CreateTextureFromSurface(t.surface)
		if err != nil {
			log.Printf("cannot render text: %s", err)
			return
		}
		t.texture = texture
	}
	r := sdl.Rect{X: t.x, Y: t.y, W: t.surface.W, H: t.surface.H}
	renderer.Copy(t.texture, nil, &r)
}

func (t *TextDisplay) Close() error {
	if t.texture != nil {
		t.texture.Destroy()
		t.texture = nil
	}
	if t.surface != nil {
		t.surface.Free()
		t.surface = nil
	}
	return nil
}
func (j *JoystickDisplay) Render(renderer *sdl.Renderer) {
	col := sdl.Color(color.NRGBA{255, 0, 0, 255})
	if j.pressed {
		col = sdl.Color(color.NRGBA{255, 255, 0, 255})
	}
	gfx.FilledCircleColor(renderer, j.x, j.y, j.radius, col)

	joyposx := j.x + int32(float32(j.radius)*j.joyX)
	joyposy := j.y + int32(float32(j.radius)*j.joyY)
	col = sdl.Color(color.NRGBA{0, 255, 0, 255})
	gfx.FilledCircleColor(renderer, joyposx, joyposy, j.radius/5, col)
}

func main() {
	var font *ttf.Font

	if err := ttf.Init(); err != nil {
		return
	}
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 480, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		panic(err)
	}
	defer mix.CloseAudio()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Printf("Couldn't get accelerated renderer: %s", err)
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
		if err != nil {
			panic(err)
		}
	}
	defer renderer.Destroy()

	sdl.JoystickEventState(sdl.ENABLE)

	if rinfo, err := renderer.GetInfo(); err == nil {
		log.Printf("Renderer info: %#v", rinfo)
	}
	log.Printf("Renderer: %#v", renderer)
	log.Printf("Window: %#v", window)
	log.Printf("num joysticks: %d", sdl.NumJoysticks())

	var joysticks [16]*sdl.Joystick

	j1 := &JoystickDisplay{x: 100, y: 300, radius: 50}
	j2 := &JoystickDisplay{x: 540, y: 300, radius: 50}

	fontOps, err := sdl.RWFromMem(latoBold)
	if err != nil {
		panic(err)
	}
	// Load the font for our text
	if font, err = ttf.OpenFontRW(fontOps, 0, 48); err != nil {
		panic(err)
	}
	defer font.Close()
	defer fontOps.Close()

	text := TextDisplay{x: 10, y: 30, color: color.NRGBA{255, 0, 255, 255}, font: font}

	mix.VolumeMusic(48) // Turn the volume down a bit
	mixOps, err := sdl.RWFromMem(soundData)
	if err != nil {
		panic(err)
	}

	mus, err := mix.LoadMUSRW(mixOps, 0)
	if err != nil {
		panic(err)
	}
	defer mus.Free()
	defer mixOps.Close()

	running := true
	tick := time.Tick(time.Microsecond * 33333)

	for running {

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		j1.Render(renderer)
		j2.Render(renderer)
		text.Render(renderer)

		renderer.Present()

		<-tick
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent: // NOTE: Please use `*sdl.QuitEvent` for `v0.4.x` (current version).
				println("Quit")
				running = false
			case *sdl.JoyAxisEvent:
				// Convert the value to a -1.0 - 1.0 range
				value := float32(t.Value) / 32768.0
				if t.Axis == 0 {
					j1.joyX = value
				} else if t.Axis == 1 {
					j1.joyY = value
				} else if t.Axis == 2 {
					j2.joyX = value
				} else if t.Axis == 3 {
					j2.joyY = value
				}
				//fmt.Printf("[%d ms] JoyAxis\ttype:%d\twhich:%c\taxis:%d\tvalue:%f\n",
				//t.Timestamp, t.Type, t.Which, t.Axis, value)
			case *sdl.JoyBallEvent:
				fmt.Println("Joystick", t.Which, "trackball moved by", t.XRel, t.YRel)
			case *sdl.JoyButtonEvent:
				if t.State == sdl.PRESSED {
					text.SetText(fmt.Sprintf("Button %d/%d pressed", t.Which, t.Button))
				} else {
					text.SetText(fmt.Sprintf("Button %d/%d released", t.Which, t.Button))
				}
				if t.Button == 14 {
					j1.pressed = t.State == sdl.PRESSED
				} else if t.Button == 15 {
					j2.pressed = t.State == sdl.PRESSED
				} else if t.Button == 2 && mix.Playing(-1) == 0 {
					mus.Play(0)
				}
			case *sdl.JoyDeviceAddedEvent:
				// Open joystick for use
				joysticks[int(t.Which)] = sdl.JoystickOpen(int(t.Which))
				if joysticks[int(t.Which)] != nil {
					fmt.Println("Joystick", t.Which, "connected")
				}
			case *sdl.JoyDeviceRemovedEvent:
				if joystick := joysticks[int(t.Which)]; joystick != nil {
					joystick.Close()
				}
				fmt.Println("Joystick", t.Which, "disconnected")
			}
		}
	}
}
