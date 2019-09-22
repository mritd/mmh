package mmh

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mritd/mmh/utils"
)

func Ping(tagOrName string) {
	// use context to manage goroutine
	ctx, cancel := context.WithCancel(context.Background())

	// monitor os signal
	cancelChannel := make(chan os.Signal)
	signal.Notify(cancelChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		switch <-cancelChannel {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			// exit all goroutine
			cancel()
		}
	}()

	server := findServerByName(tagOrName)
	if server == nil {
		utils.Exit("server not found", 1)
	} else {

	}
}
