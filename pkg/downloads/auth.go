package downloads

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/SphericalKat/go-mirror-bot/internal/config"
)

type SingleStatus struct {
	Message  string
	Filename *string
	Details  *DownloadDetails
}


func getSingleStatus(details *DownloadDetails, msg *gotgbot.Message) *SingleStatus {
	var authCode int
	if msg != nil {
		authCode = isAuthorized(msg, false)
	} else {
		authCode = 1
	}

	if authCode > -1 {
		
	} else {
		return &SingleStatus{Message: "You aren't authorized to use this bot here."}
	}
}

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
