package lifecycle

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/SphericalKat/go-mirror-bot/pkg/commands"
	"github.com/rs/zerolog/log"
)

var Dispatcher *ext.Dispatcher

func StartBot(ctx context.Context, wg *sync.WaitGroup) {
	token := config.Conf.BotToken
	if token == "" {
		log.Fatal().Msg("BOT_TOKEN is empty!")
	}

	log.Debug().Msg("Attempting to create a bot")
	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client: http.Client{},
		DefaultRequestOpts: &gotgbot.RequestOpts{
			Timeout: gotgbot.DefaultTimeout,
			APIURL:  gotgbot.DefaultAPIURL,
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new bot")
	}

	log.Debug().Msg("Bot created")

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: nil,
		DispatcherOpts: ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Error().Err(err).Msg("an error occurred while handling update:")
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		},
	})

	Dispatcher = updater.Dispatcher

	commands.RegisterCommands(updater.Dispatcher)

	log.Debug().Msg("Updater and dispatcher created")

	log.Debug().Msg("Starting long polling")
	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start polling")
	}

	log.Info().Msg("Bot has started...")

	// listen for context cancellation
	<-ctx.Done()

	// shut down bot
	log.Info().Msg("Gracefully shutting down bot")

	err = updater.Stop()
	if err != nil {
		log.Error().Err(err).Msg("Error while shutting down bot")
	}

	wg.Done()
}
