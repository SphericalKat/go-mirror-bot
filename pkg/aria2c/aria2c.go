package aria2c

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/siku2/arigo"
)

var Aria *arigo.Client

func ConnectRPC(ctx context.Context, wg *sync.WaitGroup) {
	c, err := arigo.Dial("ws://localhost:8210/jsonrpc", "some")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to aria2 RPC")
	}
	Aria = &c
	log.Info().Msg("Connected to aria2 RPC")

	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := Aria.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error while closing aria2 RPC connection. This will not affect the shutdown.")
		}
		log.Info().Msg("Disconnected from aria2 RPC")
	}()
}
