package utils

import (
	"os"
	"os/signal"
)

func Wait() {
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt)
	signal.Notify(exitChan, os.Kill)
	select {
	case <-exitChan:
		return
	}
}
