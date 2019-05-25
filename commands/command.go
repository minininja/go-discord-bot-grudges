package commands

import (
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"log"
)

type Command struct {
	KeywordStr string
	HelpStr string
	Session *discordgo.Session
}

func (command *Command) Help() string {
	return command.HelpStr
}

func (command *Command) Keyword() string {
	return command.KeywordStr
}

type CommandHandler interface {
	Handler(ctx *exrouter.Context)
}

func (command *Command) Nickname(guildId string, userId string) string {
	member, err := Session.GuildMember(guildId, userId)
	if err != nil {
		log.Printf("Couldn't lookup user's nickname %s %s", guildId, userId)
	} else if member.Nick != "" {
		return member.Nick
	}
}