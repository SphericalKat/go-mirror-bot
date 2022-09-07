package downloads

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
)

func isAuthorized(msg *gotgbot.Message, skip bool) int {
	for _, v := range config.Conf.SudoUsers {
		if v == msg.From.Id {
			return 0
		}
	}

	if !skip && msg.ReplyToMessage != nil {
		details := GetDownloadManager().GetDownloadByMsgId(msg.ReplyToMessage)
		if details != nil {
			if msg.From.Id == details.TgFromId {
				return 1
			}
		}
	}

	// this is intentional
	var isAuthorizedChat bool = false
	for _, v := range config.Conf.AuthorizedChats {
		if msg.Chat.Id == v {
			isAuthorizedChat = true
			break
		}
	}

	if isAuthorizedChat {
		return 3
	}

	return -1

}
