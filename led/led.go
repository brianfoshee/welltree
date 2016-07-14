package led

import (
	"sync"
	"time"

	"github.com/jgarff/rpi_ws281x/golang/ws2811"
)

const maxBrightness = 64
const pwmPin = 18

type Strand struct {
	mu   *sync.Mutex
	leds int
	done chan struct{}
	// TODO some sort of function channel to run during Run loop
	// interface looper
	// func Setup()
	// func Teardown()
	// func Loop(), take in a context to handle cancelation
	// pop off a looper, run setup(),
	// run loop() in a goroutine until another call comes in then cancel its context
	// and repeat
}

func NewStrand(leds int) (*Strand, error) {
	s := &Strand{
		leds: leds,
		done: make(chan struct{}, 1),
	}
	s.Clear()
	return s, ws2811.Init(pwmPin, leds, maxBrightness)
}

func (s *Strand) Run() {
	go func() {
		for {
			select {
			case <-done:
				s.Clear()
				break
			default:
				// run the loop here
				// pop off the queue and run it
			}
		}
	}()
}

func (s *Strand) Stop() {
	s.mu.Lock()
	done <- struct{}{}
	s.mu.Unlock()
}

func (s *Strand) Close() {
	s.mu.Lock()
	ws2811.Fini()
	s.mu.Unlock()
}

func (s *Strand) Clear() {
	s.mu.Lock()
	ws2811.Clear()
	s.render()
	s.mu.Unlock()
}

// render is a faster way of calling Render() and Wait()
func (s *Strand) render() {
	ws2811.Render()
	ws2811.Wait()
}

func (s *Strand) SetAll(color uint32) {
	s.mu.Lock()
	for i := 0; i < s.leds; i++ {
		ws2811.SetLed(i, color)
	}
	s.render()
	s.mu.Unlock()
}

func (s *Strand) Swap() {
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
		for i := 0; i < s.leds; i++ {
			if i%2 == 0 {
				ws2811.SetLed(i, x)
			} else {
				ws2811.SetLed(i, y)
			}
		}
		s.render()
		time.Sleep(250 * time.Millisecond)
	}
}

func (s *Strand) Pulse() {
	maxBright := maxBrightness
	minBright := 5
	k := minBright
	inc := true
	for i := 0; i < s.leds; i++ {
		ws2811.SetLed(i, 0xff0000)
	}
	for j := 0; j < 3; j++ {
		for {
			ws2811.SetBrightness(k)
			s.render()
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
	ws2811.SetBrightness(maxBrightness)
}

// push a color one LED at a time to the top
func stairClimb(leds int) {
	off := uint32(0x000000)
	on := uint32(0xffff00)
	s.Clear()

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

			s.render()
			time.Sleep(120 * time.Millisecond)
		}
	}

	for i := 0; i < leds; i++ {
		ws2811.SetLed(i, on)
	}
	s.render()
}

var pos = []int{0, 1, 2, 3, 4, 5, 6, 7}

func (S *Strand) Rainbow() {
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
			if pos[i] > s.leds {
				pos[i] = 0
			}
			ws2811.SetLed(pos[i], colors[i])
		}
		s.render()
		time.Sleep(66 * time.Millisecond)
	}
}
