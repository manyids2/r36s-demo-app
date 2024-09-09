package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	_ "embed"
)

func getLines(s string, n int) []string {
	if len(s) == 0 {
		return []string{strings.Repeat("-", n)}
	}
	lines := strings.Split(s, "\n")
	return lines
}

func main() {
	var font *ttf.Font

	cmds := []string{"ip", "ifconfig"}
	num_cmds := len(cmds)
	outputs := [][]string{}
	errors := [][]string{}
	for _, c := range cmds {
		fmt.Println("Running: ", c)
		cmd := exec.Command(c)
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		cmd.Run()
		outs := getLines(outb.String(), 80)
		errs := getLines(errb.String(), 80)
		outputs = append(outputs, outs)
		errors = append(errors, errs)
	}

	if err := ttf.Init(); err != nil {
		return
	}
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", 0, 0, 640, 480, sdl.WINDOW_SHOWN)
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

	sdl.ShowCursor(sdl.DISABLE)

	sdl.JoystickEventState(sdl.ENABLE)

	if rinfo, err := renderer.GetInfo(); err == nil {
		log.Printf("Renderer info: %#v", rinfo)
	}

	fontOps, err := sdl.RWFromMem(latoBold)
	if err != nil {
		panic(err)
	}
	// Load the font for our text
	if font, err = ttf.OpenFontRW(fontOps, 0, 12); err != nil {
		panic(err)
	}
	defer font.Close()
	defer fontOps.Close()

	cmd_text := TextDisplay{x: 10, y: 5, color: color.NRGBA{255, 0, 255, 255}, font: font}
	out_texts := make([]TextDisplay, 10)
	err_texts := make([]TextDisplay, 10)
	for i := 0; i < 10; i++ {
		out_texts[i] = TextDisplay{x: 10, y: int32((i * 15) + 25), color: color.NRGBA{255, 0, 255, 255}, font: font}
		err_texts[i] = TextDisplay{x: 10, y: int32((i * 15) + 180), color: color.NRGBA{255, 0, 255, 255}, font: font}
	}

	current := 0
	running := true
	tick := time.Tick(time.Microsecond * 33333)

	for running {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		cmd_text.Render(renderer)
		for _, oo := range out_texts {
			oo.Render(renderer)
		}
		for _, ee := range err_texts {
			ee.Render(renderer)
		}
		renderer.Present()

		<-tick

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {

			// Clear screen
			cmd_text.SetText(fmt.Sprintf("%d: %s (%d, %d)", current, cmds[current], len(out_texts), len(err_texts)))
			for i, _ := range out_texts {
				if i < len(outputs[current]) {
					if len(outputs[current][i]) > 0 {
						out_texts[i].SetText(outputs[current][i])
					}
				} else {
					out_texts[i].SetText("---")
				}
			}
			for i, _ := range err_texts {
				if i < len(errors[current]) {
					if len(errors[current][i]) > 0 {
						err_texts[i].SetText(errors[current][i])
					}
				} else {
					err_texts[i].SetText("---")
				}
			}

			switch t := event.(type) {
			case *sdl.QuitEvent: // NOTE: Please use `*sdl.QuitEvent` for `v0.4.x` (current version).
				println("Quit")
				running = false
			case *sdl.KeyboardEvent:
				if t.Keysym.Sym == sdl.K_ESCAPE {
					println("Quit")
					running = false
				}
				if (t.State == sdl.PRESSED) && ((t.Keysym.Sym == sdl.K_LEFT) || (t.Keysym.Sym == sdl.K_RIGHT)) {
					current = (current + 1) % num_cmds
				}
			case *sdl.JoyButtonEvent:
				if t.State == sdl.PRESSED {
					current = (current + 1) % num_cmds
				}
			}
		}
	}
}
