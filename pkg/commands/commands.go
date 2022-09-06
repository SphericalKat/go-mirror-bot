package commands

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "I'm alive.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	return err
}

func RegisterCommands(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("start", start))
}
