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
var zero = int64(0)

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

	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		err = discord.UpdateStatus(0, "Angrily muttering about revenge")
		if err != nil {
			fmt.Println("Error attempting to set my status")
		}
		guilds := discord.State.Guilds
		log.Printf("Started on %d servers", len(guilds))
		for _, guild := range guilds {
			// TODO need to figure out how to get the guild names, probably needs more permissions (?)
			log.Printf("\t%s - %s\n", guild.ID, guild.Name)
		}
	})
	// create the router
	router := exrouter.New()

	// add the test message
	router.On("grudge", func(ctx *exrouter.Context) {
		log.Print("grudge command invoked")
		content := strings.Split(ctx.Msg.Content, "|")

		if len(content) == 1 {
			ctx.Reply("You'll need to tell me who you want to grudge")
			ctx.Reply("Example: " + commandPrefix + "grudge the player you hate|the reason for it")
			return
		}

		target := strings.Join(strings.Split(content[0], " ")[1:], " ")
		why := content[1]
		who := ctx.Msg.Author.Username

		// try for the nickname
		member, err := discord.GuildMember(ctx.Msg.GuildID, ctx.Msg.Author.ID)
		if err != nil {
			log.Printf("Couldn't lookup user's nickname %s %s %s", ctx.Msg.GuildID, ctx.Msg.Author.ID, who)
		} else if member.Nick != "" {
			who = member.Nick
		}

		Grudge(ctx.Msg.GuildID, who, target, why)
		ctx.Reply("added grudge against " + target)
	}).Desc("Report a grudge against someone, put a | (pipe symbol) between the who and the why.")

	router.On("ungrudge", func(ctx *exrouter.Context) {
		log.Print("ungrudge command invoked")
		content := strings.Split(ctx.Msg.Content, " ")
		if len(content) == 1 {
			ctx.Reply("I can't remove a grudge against nobody")
		} else {
			ungrudge := strings.Join(content[1:], " ")
			rows := Ungrudge(ctx.Msg.GuildID, ungrudge)
			if rows > zero {
				ctx.Reply("Removed grudges against " + ungrudge)
			} else {
				ctx.Reply("Didn't find a grudge against " + ungrudge)
			}
		}
	}).Desc("Remove someone from the list")

	router.On("grudges", func(ctx *exrouter.Context) {
		log.Print("grudges command invoked")
		grudges := Grudges(ctx.Msg.GuildID)
		log.Printf("grudges %s", grudges)
		if grudges != "" {
			chunkMessage(ctx, "target : reported by : why @ when\n", grudges)
		} else {
			ctx.Reply("Hooray, there's no one we have a grudge against")
		}
	}).Desc("Show the current list of grudges")

	router.On("ally", func(ctx *exrouter.Context) {
		content := strings.Split(ctx.Msg.Content, "|")
		if 2 == len(content) {
			tmp := strings.Split(content[0], " ")
			if len(tmp) > 1 {
				ally := strings.Join(tmp[1:], " ")
				Ally(ctx.Msg.GuildID, ally, content[1])
				ctx.Reply("Saved " + ally + " as your ally(" + content[1] + ")")
				return
			}
		}
		ctx.Reply("The format for the ally command is 'Ally|STATUS'")
	}).Desc("Add a new ally")

	router.On("unally", func(ctx *exrouter.Context) {
		content := strings.Split(ctx.Msg.Content, " ")
		if len(content) >= 2 {
			ally := strings.Join(content[1:], " ")

			rows := Unally(ctx.Msg.GuildID, ally)
			if rows > zero {
				ctx.Reply("Removed ally " + ally)
			} else {
				ctx.Reply("Didn't find a " + ally + " to remove")
			}
			return
		}
		ctx.Reply("The format for the unally command is \"unally 'the ally name'\"")
	}).Desc("Remove an ally")

	router.On("allies", func(ctx *exrouter.Context) {
		content := Allies(ctx.Msg.GuildID)
		if "" == content {
			ctx.Reply("We have no allies at the moment")
		} else {
			chunkMessage(ctx, "Ally : Status @ As of when\n", content)
		}
	}).Desc("List our allies")

	router.On("roe", func(ctx *exrouter.Context) {
		channels, err := discord.GuildChannels(ctx.Msg.GuildID)

		if nil != err {
			log.Print("error reading channels: " + err.Error())
			return
		}
		log.Printf("channels %s", channels)

		for _, channel := range channels {
			log.Printf("looking at channel %s", channel.Name)
			if ("roe" == channel.Name ) {
				if (channel.ID != ctx.Msg.ChannelID) {
					messages, err := discord.ChannelMessages(channel.ID, 100, "", "", "")

					if nil != err {
						log.Print("error reading messages: " + err.Error())
						return
					}

					for i := len(messages) - 1; i >= 0; i-- {
						if !strings.HasPrefix(messages[i].Content, commandPrefix+"roe") {
							log.Printf("%s\n%s\n%s\n\n", messages[i].ID, messages[i].Content, messages[i].Timestamp)
							ctx.Reply(messages[i].Content)
						}
					}
					return
				} else {
					ctx.Reply("This function is not available inside the #roe channel")
					return
				}
			}
		}
		ctx.Reply("No rules of engagement found, define a channel named \"roe\" to use this function")
	}).Desc("Show the rules of engagement, requires that a 'roe' channel be defined")

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

func limit(mesg string) bool {
	return len(mesg) < 1990
}

func chunkMessage(ctx *exrouter.Context, header string, payload string) {
	i := 0
	mesg := ""
	parts := strings.Split(payload, "\n")

	for i < len(parts) {
		if limit(header + mesg + parts[i] + "\n") {
			mesg += parts[i] + "\n"
		} else {
			_, err := ctx.Reply(header + mesg)
			if err != nil {
				log.Printf("Error sending grudges list '%s'", err.Error())
			}
			mesg = ""
		}
		i++
	}

	if len(mesg) > 0 {
		_, err := ctx.Reply(header + mesg)
		if err != nil {
			log.Printf("Error sending grudges list '%s'", err.Error())
		}
	}
}
