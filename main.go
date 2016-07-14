package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/brianfoshee/welltree/github"
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
)

const maxBrightness = 64

func main() {
	author := flag.String("author", "", "the author to get failing PRs for")
	token := flag.String("token", "", "Github OAUTH token")
	repo := flag.String("repo", "", "Github repo to search")
	leds := flag.Int("leds", 16, "Number of LEDs in the string")
	flag.Parse()

	if *author == "" {
		fmt.Println("Please supply an author")
		return
	}

	if *token == "" {
		fmt.Println("Please supply an oauth token")
		return
	}

	if *repo == "" {
		fmt.Println("Please supply a repo")
		return
	}

	s := github.NewSearch(*author, *token, *repo)

	ticker := time.NewTicker(10 * time.Second)
	done := make(chan struct{}, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		fmt.Println("Catching signal", sig)
		ticker.Stop()
		done <- struct{}{}
	}()

	ws2811.Init(18, *leds, maxBrightness)
	defer ws2811.Fini()

	// set all LEDs off initially
	ws2811.Clear()
	ws2811.Render()
	ws2811.Wait()

	// store a failing state to handle transitions between passing/failing
	failing := false
runloop:
	for {
		select {
		case <-done:
			fmt.Println("Done")
			break runloop
		case <-ticker.C:
			var color uint32
			f, err := s.Failing()
			if err != nil {
				// Error is always due to a network or request issue
				fmt.Println(err)
				rainbow(*leds)
				break
			}
			if f {
				// if it was passing before, and now we're failing, pulse
				if !failing {
					pulse(*leds)
					ws2811.SetBrightness(maxBrightness)
				}
				failing = true
				fmt.Println("Failing...")
				// turn them all steady red
				color = 0xff0000
			} else {
				// if it was failing before, and now we're passing, do another transition
				if failing {
					swap(*leds)
				}
				failing = false
				fmt.Println("Passing!")
				// turn them all steady green
				color = 0x00ff00
			}
			for i := 0; i < *leds; i++ {
				ws2811.SetLed(i, color)
			}
			ws2811.Render()
			ws2811.Wait()
		}
	}

	ws2811.Clear()
	ws2811.Render()
	ws2811.Wait()
}

func swap(leds int) {
	var x uint32 = 0x00ff00
	var y uint32 = 0x0000ff
	for j := 0; j < 20; j++ {
		if j%2 == 0 {
			x = 0x00ff00
			y = 0x0000ff
		} else {
			x = 0x0000ff
			y = 0x00ff00
		}
		for i := 0; i < leds; i++ {
			if i%2 == 0 {
				ws2811.SetLed(i, x)
			} else {
				ws2811.SetLed(i, y)
			}
		}
		ws2811.Render()
		ws2811.Wait()
		time.Sleep(250 * time.Millisecond)
	}
}

func pulse(leds int) {
	maxBright := maxBrightness
	minBright := 5
	k := minBright
	inc := true
	for i := 0; i < leds; i++ {
		ws2811.SetLed(i, 0xff0000)
	}
	for j := 0; j < 3; j++ {
		for {
			ws2811.SetBrightness(k)
			ws2811.Render()
			ws2811.Wait()
			time.Sleep(14 * time.Millisecond)
			if inc {
				k += 1
			} else {
				k -= 1
			}
			if k > maxBright {
				k = maxBright
				inc = false
			} else if k < minBright {
				k = minBright
				inc = true
				break // break the for loop for this 'pulse'
			}
		}
	}
}

// push a color one LED at a time to the top
func stairClimb(leds int) {
	off := uint32(0x000000)
	on := uint32(0xffff00)
	ws2811.Clear()
	ws2811.Render()
	ws2811.Wait()

	for i := 0; i < 5; i++ {
		for j := 0; j < leds; j++ {
			// set curr-1 off
			if j == 0 {
				ws2811.SetLed(leds-1, off)
			} else {
				ws2811.SetLed(j-1, off)
			}

			// set curr on
			ws2811.SetLed(j, on)

			ws2811.Render()
			ws2811.Wait()
			time.Sleep(120 * time.Millisecond)
		}
	}

	for i := 0; i < leds; i++ {
		ws2811.SetLed(i, on)
	}
	ws2811.Render()
	ws2811.Wait()
}

var pos = []int{0, 1, 2, 3, 4, 5, 6, 7}

func rainbow(leds int) {
	colors := []uint32{
		0x00200000, // red
		0x00201000, // orange
		0x00202000, // yellow
		0x00002000, // green
		0x00002020, // lightblue
		0x00000020, // blue
		0x00100010, // purple
		0x00200010, // pink
	}
	// start pushing colors one at a time into the led strip
	//offset := 0
	for k := 0; k < 145; k++ {
		for i := 0; i < len(pos); i++ {
			pos[i]++
			if pos[i] > leds {
				pos[i] = 0
			}
			ws2811.SetLed(pos[i], colors[i])
		}
		ws2811.Render()
		ws2811.Wait()
		time.Sleep(66 * time.Millisecond)
	}
}
