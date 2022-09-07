package downloads

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/SphericalKat/arigo"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
	"github.com/SphericalKat/go-mirror-bot/pkg/aria2c"
	"github.com/rs/zerolog/log"
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
		log.Info().Uint("retry", retry).Msg("Retrying download start")
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

func onDownloadError(event *arigo.DownloadEvent, retry uint) {
	dlm := GetDownloadManager()
	details := dlm.GetDownloadByGid(event.GID)
	if details != nil {
		var message string
		status, err := aria2c.Aria.TellStatus(event.GID, "errorMessage")
		if err != nil {
			message = "Download stopped."
			log.Error().Str("gid", event.GID).Err(err).Msg("Download failed. Unable to get failure reason.")
		} else {
			message = fmt.Sprintf("Download stopped. %s", status.ErrorMessage)
			log.Error().Str("gid", event.GID).Msg("Download failed.")
		}
		cleanupDownload(event.GID, message, "", details)
	} else if retry <= 8 {
		go func() {
			time.Sleep(500 * time.Millisecond)
			onDownloadError(event, retry+1)
		}()
	} else {
		log.Error().Str("gid", event.GID).Msg("Download details empty even after 8 retries, giving up")
	}
}

func onDownloadComplete(event *arigo.DownloadEvent, retry uint) {
	dlm := GetDownloadManager()
	details := dlm.GetDownloadByGid(event.GID)
	if details != nil {
		files, err := aria2c.Aria.GetFiles(event.GID)
		if err != nil {
			log.Error().Err(err).Str("gid", event.GID).Msg("Error getting file path for completed download")
			msg := "Upload failed. Could not find downloaded files."
			cleanupDownload(event.GID, msg, "", nil)
			return
		}

		file := findAriaFilePath(files).Path
		if file != "" {
			_, err := aria2c.Aria.TellStatus(event.GID, "totalLength")
			if err != nil {
				log.Error().Err(err).Str("gid", event.GID).Msg("Error getting file size for completed download")
				msg := "Upload failed. Could not get file size."
				cleanupDownload(event.GID, msg, "", nil)
				return
			}

			filename := getFileNameFromPath(file, "", "")
			details.IsUploading = true
			log.Info().Str("gid", event.GID).Str("filename", filename).Msg("Download complete. Starting upload")

		} else {
			status, err := aria2c.Aria.TellStatus(event.GID, "followedBy")
			if err != nil {
				log.Error().Err(err).Str("gid", event.GID).Msg("Failed to check if it was a metadata download")
				msg := "Upload failed. Could not check if the file is metadata."
				cleanupDownload(event.GID, msg, "", nil)
			} else if status.FollowedBy != nil && len(status.FollowedBy) != 0 {
				log.Info().Str("oldGid", event.GID).Str("newGid", status.FollowedBy[0]).Msg("Download GID changed.")
				dlm.ChangeDownloadGid(event.GID, status.FollowedBy[0])
			} else {
				log.Error().Err(err).Str("gid", event.GID).Msg("No files - not metadata")
				msg := "Upload failed. Could not get files"
				cleanupDownload(event.GID, msg, "", nil)
			}
		}
	} else if retry <= 8 {
		go func() {
			time.Sleep(500 * time.Millisecond)
			onDownloadComplete(event, retry+1)
		}()
	} else {
		log.Error().Str("gid", event.GID).Msg("Download details empty even after 8 retries, giving up")
	}
}

func RegisterCommands(ctx context.Context, wg *sync.WaitGroup) {
	lifecycle.Dispatcher.AddHandler(handlers.NewCommand("start", start))
	lifecycle.Dispatcher.AddHandler(handlers.NewCommand("upload", upload))

	aria2c.Aria.OnDownloadStart(func(event *arigo.DownloadEvent) {
		onDownloadStart(event, 1)
	})

	aria2c.Aria.OnDownloadStop(func(event *arigo.DownloadEvent) {
		onDownloadStop(event, 1)
	})

	aria2c.Aria.OnDownloadComplete(func(event *arigo.DownloadEvent) {
		onDownloadComplete(event, 1)
	})

	aria2c.Aria.OnDownloadError(func(event *arigo.DownloadEvent) {
		onDownloadError(event, 1)
	})

	// listen for shutdowns and tickers
	go func() {
		go func() {
			<-ctx.Done()
			if ticker != nil {
				ticker.Stop()
			}
			wg.Done()
		}()
		for {
			if ticker != nil {
				<-ticker.C
				updateAllStatusMessages()

			}
		}
	}()
}
