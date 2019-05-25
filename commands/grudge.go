package commands

import (
	"github.com/Necroforger/dgrouter/exrouter"
	"strings"
	"log"
)

type Grudge struct {
	Command
	CommandPrefix string
}

func (grudge *Grudge) Handle(ctx *exrouter.Context) string {
	content := strings.Split(ctx.Msg.Content, "|")

	if len(content) == 1 {
		ctx.Reply("You'll need to tell me who you want to grudge")
		ctx.Reply("Example: "+ CommandPrefix + "grudge the player you hate|the reason for it")
		return
	}

	target := strings.Join(strings.Split(content[0], " ")[1:], " ")
	why := content[1]
	who := ctx.Msg.Author.Username

	// try for the nickname
	tmp := Nikename(ctx.Msg.GuildID, ctx.Msg.Author.ID)
	if tmp != "" {
		who = tmp
	}

	Grudge(ctx.Msg.GuildID, who, target, why)
	ctx.Reply("added grudge against " + target)
}