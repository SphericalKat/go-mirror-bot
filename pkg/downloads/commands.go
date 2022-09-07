package downloads

import (
	"context"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
	"github.com/SphericalKat/go-mirror-bot/pkg/aria2c"
	"github.com/rs/zerolog/log"
	"github.com/siku2/arigo"
)

var ticker *time.Ticker

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "I'm alive.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	return err
}

func upload(b *gotgbot.Bot, ctx *ext.Context) error {
	if IsAuthorized(ctx.EffectiveMessage, false) < 0 {
		ctx.EffectiveMessage.Reply(b, "You are not authorized to use this bot.", nil)
	} else {
		if len(ctx.Args()) == 2 {
			PrepDownload(ctx.EffectiveMessage, ctx.Args()[1], false)
		} else {
			return nil
		}
	}
	return nil
}

func onDownloadStart(event *arigo.DownloadEvent, retry uint) {
	dlm := GetDownloadManager()
	details := dlm.GetDownloadByGid(event.GID)
	if details != nil {
		dlm.MoveDownloadToActive(details)
		log.Info().Str("gid", event.GID).Str("dir", details.DownloadDir).Msg("Download started")
		updateAllStatusMessages()

		_, _, _, err := GetStatus(details)
		if err == nil {
			// TODO: handle disallowed filenames
		}

		if ticker == nil {
			ticker = time.NewTicker(time.Duration(config.Conf.StatusUpdateDuration) * time.Millisecond)
		}

	} else if retry <= 8 {
		go func() {
			time.Sleep(500 * time.Millisecond)
			onDownloadStart(event, retry+1)
		}()
	} else {
		log.Error().Str("gid", event.GID).Msg("Download details empty even after 8 retries, giving up")
	}
}

func onDownloadStop(event *arigo.DownloadEvent, retry uint) {
	dlm := GetDownloadManager()
	details := dlm.GetDownloadByGid(event.GID)
	if details != nil {
		log.Info().Str("gid", event.GID).Msg("Download stopped")
		message := "Download stopped."
		cleanupDownload(event.GID, message, "", nil)
	} else if retry <= 8 {
		go func() {
			time.Sleep(500 * time.Millisecond)
			onDownloadStop(event, retry+1)
		}()
	} else {
		log.Error().Str("gid", event.GID).Msg("Download details empty even after 8 retries, giving up")
	}
}

func RegisterCommands(ctx context.Context, wg *sync.WaitGroup) {
	lifecycle.Dispatcher.AddHandler(handlers.NewCommand("start", start))
	lifecycle.Dispatcher.AddHandler(handlers.NewCommand("upload", upload))

	aria2c.Aria.Subscribe(arigo.StartEvent, func(event *arigo.DownloadEvent) {
		onDownloadStart(event, 1)
	})

	aria2c.Aria.Subscribe(arigo.StopEvent, func(event *arigo.DownloadEvent) {
		onDownloadStop(event, 1)
	})

	aria2c.Aria.Subscribe(arigo.CompleteEvent, func(event *arigo.DownloadEvent) {

	})

	aria2c.Aria.Subscribe(arigo.ErrorEvent, func(event *arigo.DownloadEvent) {

	})

	// listen for shutdowns and tickers
	go func() {
		for {
			select {
			case <-ctx.Done():
				if ticker != nil {
					ticker.Stop()
				}
				wg.Done()
			case <-ticker.C:
				updateAllStatusMessages()
			}
		}
	}()
}
