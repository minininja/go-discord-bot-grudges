package main

import (
	"flag"
	"fmt"
	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"strings"
)

var (
	commandPrefix string
	botID         string
	debug         bool
)

var Session, _ = discordgo.New()

func init() {
	Session.Token = os.Getenv("DG_TOKEN")
	if Session.Token == "" {
		flag.StringVar(&Session.Token, "t", "", "Discord Auth Token")
	}

	commandPrefix = os.Getenv("DG_COMMAND_PREFIX")
	if commandPrefix == "" {
		flag.StringVar(&commandPrefix, "cp", "!", "Discord command prefix")
	}

	flag.BoolVar(&debug, "debug", false, "Enable debug message logger mode")
}

func main() {
	flag.Parse()

	if Session.Token == "" {
		log.Fatal("A discord token must be provided")
		return
	}

	discord, err := discordgo.New("Bot " + Session.Token)
	errCheck("error creating discord session", err)

	// make sure we have a user account
	user, err := discord.User("@me")
	errCheck("error retrieving account", err)
	log.Printf("Running as %s\n", user.Username)
	log.Printf("Command prefix is %s\n", commandPrefix)

	// create the router
	router := exrouter.New()

	// add the test message
	router.On("grudge", func(ctx *exrouter.Context) {
		content := strings.Split(ctx.Msg.Content, " ")

		if len(content) == 1 {
			ctx.Reply("You'll need to tell me who you want to grudge")
			return
		}

		if len(content) == 2 {
			ctx.Reply("I need a reason to grudge " + content[1])
			return
		}

		target := content[1]
		why := strings.Join(content[2:], " ")
		who := ctx.Msg.Author.Username

		// try for the nickname
		member, err := discord.GuildMember(ctx.Msg.GuildID, ctx.Msg.Author.ID)
		if err != nil {
			log.Printf("Couldn't lookup user's nickname %s %s %s", ctx.Msg.GuildID, ctx.Msg.Author.ID, who)
		} else if member.Nick != "" {
			who = member.Nick
		}

		InsertGrudge(ctx.Msg.GuildID, who, target, why)
		ctx.Reply("added grudge against " + target)
	}).Desc("Report a grudge against someone, format is <target> <why>")

	router.On("ungrudge", func(ctx *exrouter.Context) {
		target := strings.Join(strings.Split(ctx.Msg.Content, " ")[1:], " ")
		DeleteGrudge(ctx.Msg.GuildID, target)
		ctx.Reply("removed grudges against " + target)
	}).Desc("Remove someone from the list")

	router.On("grudges", func(ctx *exrouter.Context) {
		grudges := ListGrudges(ctx.Msg.GuildID)
		if grudges != "" {
			ctx.Reply("target : reported by : why @ when\n" + grudges)
		} else {
			ctx.Reply("hooray, there's no one we have a grudge against")
		}
	}).Desc("Show the current list of grudges")

	// add the default/help message
	router.Default = router.On("help", func(ctx *exrouter.Context) {
		var text = ""
		for _, v := range router.Routes {
			text += fmt.Sprintf("%8s: %s\n", v.Name, v.Description)
		}
		ctx.Reply("```" + text + "```")
	}).Desc("Displays the the help menu")

	// add the router as a handler
	discord.AddHandler(messageLogger)
	discord.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(discord, commandPrefix, discord.State.User.ID, m.Message)
	})

	// connect to discord
	err = discord.Open()
	errCheck("Error opening connection to Discord", err)

	log.Println("Bot is now running")
	<-make(chan struct{})
}

func errCheck(msg string, err error) {
	if err != nil {
		log.Fatalf("%s %s\n", msg, err)
		panic(err)
	}
}

func messageLogger(session *discordgo.Session, message *discordgo.MessageCreate) {
	if debug {
		// no need to log our own messages
		if session.State.User.ID == message.Author.ID {
			return
		}

		log.Printf("%s %s %s %s\n", message.GuildID, message.ChannelID, message.Author.Username, message.Content)
	}
}
