package graceful

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func ListenExitSignal(ec chan<- error) {
	if ec == nil {
		ec = make(chan error, 1)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
		ec <- fmt.Errorf("%s", <-sigChan)
	}()
}
