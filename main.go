package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jgarff/rpi_ws281x/golang/ws2811"
)

type searchResult struct {
	Items []item `json:"items"`
}

type item struct {
	User user `json:"user"`
}

type user struct {
	Login string `json:"login"`
}

func main() {
	author := flag.String("author", "", "the author to get failing PRs for")
	token := flag.String("token", "", "Github OAUTH token")
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

	ticker := time.NewTicker(10 * time.Second)
	done := make(chan struct{}, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Println("Catching signal", sig)
			ticker.Stop()
			done <- struct{}{}
		}
	}()

	ws2811.Init(18, *leds, 32)
	defer ws2811.Fini()

	// set all LEDs off initially
	ws2811.Clear()
	ws2811.Render()
	ws2811.Wait()

	client := &http.Client{}

	failing := false
	for {
		select {
		case <-done:
			fmt.Println("Done")
			ws2811.Clear()
			ws2811.Render()
			ws2811.Wait()
			return
		case <-ticker.C:
			// TODO pull this GH functionality out into a standalone package
			// so that I can test it when not on a raspberry pi
			qs := "?q=type:pr+repo:nytm/np-well+state:open+status:failure+author:" + *author
			req, err := http.NewRequest("GET", "https://api.github.com/search/issues"+qs, nil)
			if err != nil {
				fmt.Println("error making new request", err)
				return
			}
			req.Header.Add("Authorization", "token "+*token)
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("error doing request", err)
				return
			}

			if resp.StatusCode != 200 {
				fmt.Printf("bad response code %d\n", resp.StatusCode)
				resp.Body.Close()
				return
			}

			var sr searchResult
			if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
				fmt.Println("error decoding body", err)
				return
			}
			resp.Body.Close()

			// TODO add logic to handle the edge case of a PR failing and
			// showing up under the status:failing criteria, and then a
			// new commit is pushed up which forces the tests to run again
			// and sets the status to pending, which makes this check look
			// like it's passing but really it might not be.
			var color uint32
			if len(sr.Items) > 0 {
				// TODO push all items into a slice with PR #

				// if it was passing before, and now we're failing, pulse
				if !failing {
					pulse(*leds)
					ws2811.SetBrightness(32)
				}
				failing = true
				fmt.Println("Failing...")
				// turn them all steady red
				color = 0xff0000
			} else {
				// TODO before marking all as passed, check any PRs in the
				// stored slice to make sure they're not pending. If they are
				// pending, don't mark as passed yet. If they're closed, remove
				// them from the slice.

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
}

func swap(leds int) {
	var x uint32 = 0x00ff00
	var y uint32 = 0x0000ff
	for j := 0; j < 10; j++ {
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
	k := 5
	maxBright := 32
	minBright := 5
	inc := true
	for j := 0; j < 3; j++ {
		for {
			ws2811.SetBrightness(k)
			for i := 0; i < leds; i++ {
				ws2811.SetLed(i, 0xff0000)
				ws2811.Render()
				ws2811.Wait()
			}
			time.Sleep(15 * time.Millisecond)
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
