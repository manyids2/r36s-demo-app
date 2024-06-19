package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

type JoystickDisplay struct {
	x, y       int32
	radius     int32
	joyX, joyY float32
}

func (j *JoystickDisplay) Render(renderer *sdl.Renderer) {
	col := sdl.Color(color.NRGBA{255, 0, 0, 255})
	gfx.CircleColor(renderer, j.x, j.y, j.radius, col)

	joyposx := j.x + int32(float32(j.radius)*j.joyX)
	joyposy := j.y + int32(float32(j.radius)*j.joyY)
	col = sdl.Color(color.NRGBA{0, 255, 0, 255})
	gfx.CircleColor(renderer, joyposx, joyposy, j.radius/5, col)
}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 480, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

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

	log.Printf("Renderer: %#v", renderer)
	log.Printf("Window: %#v", window)
	log.Printf("num joysticks: %d", sdl.NumJoysticks())

	var joysticks [16]*sdl.Joystick

	j1 := &JoystickDisplay{x: 100, y: 100, radius: 50}
	j2 := &JoystickDisplay{x: 200, y: 100, radius: 50}

	running := true
	tick := time.Tick(time.Microsecond * 33333)

	for running {

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		j1.Render(renderer)
		j2.Render(renderer)
		/*
			renderer.SetDrawColor(255, 255, 255, 255)
			renderer.DrawPoint(150, 300)

			renderer.SetDrawColor(0, 0, 255, 255)
			renderer.DrawLine(0, 0, 200, 200)

			points := []sdl.Point{{0, 0}, {100, 300}, {100, 300}, {200, 0}}
			renderer.SetDrawColor(255, 255, 0, 255)
			renderer.DrawLines(points)

			rect := sdl.Rect{300, 0, 200, 200}
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&rect)

			rects := []sdl.Rect{{400, 400, 100, 100}, {550, 350, 200, 200}}
			renderer.SetDrawColor(0, 255, 255, 255)
			renderer.DrawRects(rects)

			rect = sdl.Rect{250, 250, 200, 200}
			renderer.SetDrawColor(0, 255, 0, 255)
			renderer.FillRect(&rect)

			rects = []sdl.Rect{{500, 300, 100, 100}, {200, 300, 200, 200}}
			renderer.SetDrawColor(255, 0, 255, 255)
			renderer.FillRects(rects)
		*/
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
					fmt.Println("Joystick", t.Which, "button", t.Button, "pressed")
				} else {
					fmt.Println("Joystick", t.Which, "button", t.Button, "released")
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
