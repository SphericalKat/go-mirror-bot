package aria2c

import (
	"github.com/rs/zerolog/log"
	"github.com/siku2/arigo"
)

func ConnectRPC() {
	c, err := arigo.Dial("ws://localhost:6800/jsonrpc", "")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to aria2 RPC")
	}

	c.AddURI()
}