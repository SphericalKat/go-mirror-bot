package downloads

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalKat/arigo"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
	"github.com/SphericalKat/go-mirror-bot/pkg/aria2c"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func PrepDownload(msg *gotgbot.Message, match string, isTar bool) {
	dlDir := uuid.NewString()
	dlm := GetDownloadManager()

	// renew the connection. just in case
	aria2c.ReconnectRPC()

	log.Debug().Str("uri", match).Msg("Adding URI to aria")

	gid, err := aria2c.Aria.AddURI([]string{match}, &arigo.Options{
		Dir: filepath.Join(config.Conf.DownloadDir, dlDir),
	})

	dlm.AddDownload(gid.GID, dlDir, msg, isTar)

	if err != nil {
		message := fmt.Sprintf("Failed to start the download. %s", err)
		log.Error().Err(err).Str("uri", match).Msg("failed to start download")
		cleanupDownload(gid.GID, message, "", nil)
		return
	}

	log.Info().Str("gid", gid.GID).Str("uri", match).Msg("Starting download")

	go func(msg *gotgbot.Message) {
		time.Sleep(1 * time.Second)
		dlm.QueueStatus(msg, sendStatusMessage)
	}(msg)
}

func cleanupDownload(gid string, message string, url string, details *DownloadDetails) {
	dlm := GetDownloadManager()
	if details == nil {
		details = dlm.GetDownloadByGid(gid)
	}

	if details != nil {
		wasCancelAlled := false
		dlm.ForEachCancelledDownload(func(detail *DownloadDetails) {
			if detail.Gid == gid {
				wasCancelAlled = true
			}
		})

		if !wasCancelAlled {
			// If the dl was stopped with a cancelAll command, a message has already been sent to the chat.
			// Do not send another one.
			var user string
			if details.TgRepliedUsername != "" {
				user = details.TgRepliedUsername
			} else {
				user = details.TgUsername
			}
			message += fmt.Sprintf("\ncc: %s", user)
			_, err := lifecycle.Bot.SendMessage(details.TgChatId, message, &gotgbot.SendMessageOpts{
				ReplyToMessageId: details.TgMsgId,
				ParseMode:        "html",
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send cleanup message")
			}

		}
		dlm.RemoveCancelledDownload(gid)
		dlm.DeleteDownload(gid)
		deleteDownloadedFile(details.DownloadDir)
	} else {
		log.Error().Str("gid", gid).Msg("Could not get DownloadDetails")
	}
}

func sendStatusMessage(msg *gotgbot.Message, keepForever bool) {
	dlm := GetDownloadManager()
	lastStatus := dlm.GetStatus(msg.Chat.Id)

	if lastStatus != nil {
		_, err := lifecycle.Bot.DeleteMessage(msg.Chat.Id, lastStatus.Msg.MessageId, nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to delete last status message")
		}
		dlm.DeleteStatus(msg.Chat.Id)
	}

	res := GetStatusMessage()
	if keepForever {
		msg, err := lifecycle.Bot.SendMessage(msg.Chat.Id, res.message, &gotgbot.SendMessageOpts{
			ReplyToMessageId: msg.MessageId,
			ParseMode:        "html",
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send status message")
			return
		}
		dlm.AddStatus(msg, res.message)
		return
	} else {
		msg, err := lifecycle.Bot.SendMessage(msg.Chat.Id, res.message, &gotgbot.SendMessageOpts{
			ReplyToMessageId: msg.MessageId,
			ParseMode:        "html",
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send status message")
			return
		}
		dlm.AddStatus(msg, res.message)
		go func(msg *gotgbot.Message) {
			time.Sleep(1 * time.Minute)
			_, err := lifecycle.Bot.DeleteMessage(msg.Chat.Id, msg.MessageId, nil)
			if err != nil {
				log.Error().Err(err).Msg("Failed to delete status message")
			}
			dlm.DeleteStatus(msg.Chat.Id)
		}(msg)
	}
}
