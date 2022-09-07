package main

import (
	"context"
	"sync"

	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
	"github.com/SphericalKat/go-mirror-bot/pkg/aria2c"
	"github.com/SphericalKat/go-mirror-bot/pkg/downloads"
	"github.com/rs/zerolog/log"
)

func main() {
	config.Load()

	// create a waitgroup for all tasks
	wg := sync.WaitGroup{}

	// create context for background tasks
	ctx, cancelFunc := context.WithCancel(context.Background())

	botInit := make(chan struct{}, 1)

	// connect to aria2 RPC
	wg.Add(1)
	aria2c.ConnectRPC(ctx, &wg)

	// start event loop
	wg.Add(1)
	go downloads.ConsumeStatusQueue(ctx, &wg)

	// start bot polling
	wg.Add(1)
	go lifecycle.StartBot(ctx, &wg, botInit)

	// wait for bot to be initialized
	<-botInit
	downloads.RegisterCommands()

	// add signal handler to gracefully shut down tasks
	wg.Add(1)
	go lifecycle.ShutdownListener(&wg, &cancelFunc)

	// wait for all long running tasks to complete
	wg.Wait()

	log.Info().Msg("Graceful shutdown complete")
}
