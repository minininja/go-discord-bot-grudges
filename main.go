package main

import (
	"flag"
	"os"
	"fmt"
	"log"
	"github.com/bwmarrin/discordgo"
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
}

func main() {
	flag.Parse()

	if (Session.Token == "") {
		log.Fatal("A discord token must be provided")
		return
	}

	discord, err := discordgo.New("Bot " + Session.Token)
	errCheck("error creating discord session", err)
	user, err := discord.User("@me")
	errCheck("error retrieving account", err)

	botID = user.ID
	discord.AddHandler(commandHandler)
	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		err = discord.UpdateStatus(0, "A friendly helpful bot!")
		if err != nil {
			fmt.Println("Error attempting to set my status")
		}
		servers := discord.State.Guilds
		fmt.Printf("GrudgesBot has started on %d servers", len(servers))
	})

	err = discord.Open()
	errCheck("Error opening connection to Discord", err)
	defer discord.Close()

	commandPrefix = "/"

	<-make(chan struct{})
}

func errCheck(msg string, err error) {
	if err != nil {
		log.Fatal("%s %s", msg, err)
		panic(err)
	}
}

func commandHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	user := message.Author
	if user.ID == botID || user.Bot {
		//Do nothing because the bot is talking
		return
	}

	content := message.Content
	if (content == commandPrefix + "test") {
		discord.ChannelMessageSend(message.ChannelID, "Testing...")
	}

	fmt.Printf("Message: %+v || From: %s\n", message.Message, message.Author)
}