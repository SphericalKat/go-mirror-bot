package downloads

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/SphericalKat/go-mirror-bot/internal/lifecycle"
)

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

func RegisterCommands() {
	lifecycle.Dispatcher.AddHandler(handlers.NewCommand("start", start))
	lifecycle.Dispatcher.AddHandler(handlers.NewCommand("upload", upload))
}
