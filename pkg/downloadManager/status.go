package downloadmanager

import "github.com/PaulSonOfLars/gotgbot/v2"

type Status struct {
	Msg        *gotgbot.Message
	LastStatus string
}
