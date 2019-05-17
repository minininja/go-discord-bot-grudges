package main

import (
	"flag"
	"os"
	"log"
	"github.com/bwmarrin/discordgo"
	"github.com/Necroforger/dgrouter/exrouter"
	"strings"
	"fmt"
)

var (
	commandPrefix string
	botID         string
)

var Session, _ = discordgo.New()

func init() {
	Session.Token = os.Getenv("DG_TOKEN")
	if (Session.Token == "") {
		flag.StringVar(&Session.Token, "t", "", "Discord Auth Token")
	}

	commandPrefix = os.Getenv( "DG_COMMAND_PREFIX")
	if (commandPrefix == "") {
		flag.StringVar(&commandPrefix, "cp", "!", "Discord command prefix")
	}
}

func main() {
	flag.Parse()

	if (Session.Token == "") {
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
		target := strings.Split(ctx.Msg.Content, " ")[1]
		why := strings.Join(strings.Split(ctx.Msg.Content, " ")[2:], " ")
		log.Printf("adding grudge against %s because of %s\n", target, why)
		InsertGrudge(ctx.Msg.Author.Username, target, why)
		ctx.Reply("added grudge against " + target)
	}).Desc("Report a grudge against someone, format is <target> <why>")

	router.On("ungrudge", func(ctx *exrouter.Context) {
		target := strings.Join(strings.Split(ctx.Msg.Content, " ")[1:], " ")
		DeleteGrudge(target)
		ctx.Reply("removed grudges against " + target)
	}).Desc("Remove someone from the list")

	router.On("grudges", func(ctx *exrouter.Context) {
		grudges := ListGrudges()
		if (grudges != "") {
			ctx.Reply(grudges)
		} else {
			ctx.Reply("hooray, there's no one we have a grudge against")
		}
	}).Desc("Show the current list of grudges")

	// add the default/help message
	router.Default = router.On("help",func(ctx *exrouter.Context) {
		var text = ""
		for _, v := range router.Routes {
			text += fmt.Sprintf("%8s: %s\n", v.Name, v.Description)
		}
		ctx.Reply("```" + text + "```")
	}).Desc("Displays the the help menu")

	// add the router as a handler
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



