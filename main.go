package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/brianfoshee/welltree/github"
	"github.com/brianfoshee/welltree/led"
)

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

	l, err := led.NewStrand(*leds)
	if err != nil {
		fmt.Println("could not initialize LED strip", err)
		os.Exit(1)
	}
	defer l.Close()
	l.Run()

	s := github.NewSearch(*author, *token, *repo)

	ticker := time.NewTicker(10 * time.Second)
	done := make(chan struct{}, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		fmt.Println("Catching signal", sig)
		ticker.Stop()
		l.Stop()
		done <- struct{}{}
	}()

	// store a failing state to handle transitions between passing/failing
	failing := false
runloop:
	for {
		select {
		case <-done:
			fmt.Println("Done.")
			break runloop
		case <-ticker.C:
			f, err := s.Failing()
			if err != nil {
				// Error is always due to a network or request issue
				fmt.Println(err)
				l.Rainbow()
				break
			}
			if f {
				fmt.Println("Failing...")
				// if it was passing before, and now we're failing, pulse
				if !failing {
					l.Pulse()
				}
				failing = true
				// turn them all steady red
				l.SetAll(0xff0000)
			} else {
				fmt.Println("Passing!")
				// if it was failing before, and now we're passing, do another transition
				if failing {
					l.Swap()
				}
				failing = false
				// turn them all steady green
				l.SetAll(0x00ff00)
			}
		}
	}
}
