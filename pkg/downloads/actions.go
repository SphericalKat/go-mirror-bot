package downloads

import (
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
	"github.com/SphericalKat/go-mirror-bot/pkg/aria2c"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/siku2/arigo"
)

func PrepDownload(msg *gotgbot.Message, match string, isTar bool) {
	dlDir := uuid.NewString()
	gid, err := aria2c.Aria.AddURI([]string{match}, &arigo.Options{
		Dir: dlDir,
	})
	GetDownloadManager().AddDownload(gid.GID, dlDir, msg, isTar)

	if err != nil {
		message := fmt.Sprintf("Failed to start the download.", err)
		log.Error().Err(err).Str("uri", match).Msg("failed to start download")
		return
	}

	log.Info().Str("gid", gid.GID).Str("uri", match).Msg("Starting download")

	go func() {
		time.Sleep(1 * time.Second)
		GetDownloadManager().QueueStatus(msg, sendStatusMessage)
	}()
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
			if (details.TgRepliedUsername != "")  {
				message += fmt.Sprintf("\ncc:%s", details.TgRepliedUsername)
				lifecycle.Bot.SendMessage(details.TgChatId, message, &gotgbot.SendMessageOpts{
					ReplyToMessageId: details.TgMsgId,
					ParseMode: "html",
				})
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

}
