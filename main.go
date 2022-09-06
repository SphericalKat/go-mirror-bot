package main

import (
	"context"
	"sync"

	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
	"github.com/rs/zerolog/log"
)

func main() {
	config.Load()

	// create a waitgroup for all tasks
	wg := sync.WaitGroup{}

	// create context for background tasks
	ctx, cancelFunc := context.WithCancel(context.Background())

	// start bot polling
	wg.Add(1)
	go lifecycle.StartBot(ctx, &wg)

	// add signal handler to gracefully shut down tasks
	wg.Add(1)
	go lifecycle.ShutdownListener(&wg, &cancelFunc)

	// wait for all long running tasks to complete
	wg.Wait()

	log.Info().Msg("Graceful shutdown complete")
}
