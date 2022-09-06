package downloads

import (
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type DownloadDetails struct {
	IsUploading              bool
	UploadedBytes            int64
	UploadedBytesLast        int64
	LastUploadCheckTimeStamp *time.Time
	IsDownloadingAllowed     int64
	IsDownloading            bool
	Gid                      string
	TgFromId                 int64
	TgUsername               string
	TgRepliedUsername        string
	TgChatId                 int64
	TgMsgId                  int64
	StartTime                int64
	DownloadDir              string
}

func getUsername(msg *gotgbot.Message) string {
	if msg.From.Username != "" {
		return fmt.Sprintf("@%s", msg.From.Username)
	} else {
		return fmt.Sprintf("<a href=\"tg://user?id=%d\">%s</a>", msg.From.Id, msg.From.FirstName)
	}
}

func NewDownloadDetails(gid string, msg *gotgbot.Message, isTar bool, downloadDir string) *DownloadDetails {
	dd := &DownloadDetails{
		Gid:               gid,
		DownloadDir:       downloadDir,
		TgFromId:          msg.From.Id,
		TgChatId:          msg.Chat.Id,
		TgMsgId:           msg.MessageId,
		StartTime:         time.Now().Unix(),
		UploadedBytes:     0,
		UploadedBytesLast: 0,
	}

	dd.TgUsername = getUsername(msg)

	if msg.ReplyToMessage != nil {
		dd.TgRepliedUsername = getUsername(msg.ReplyToMessage)
	}

	return dd
}

