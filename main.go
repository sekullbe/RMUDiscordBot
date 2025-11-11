package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jcheng31/diceroller/dice"
)

var d100 dice.Die

var allRolls []int
var rollsByUser map[string][]int

func main() {

	var token string
	tokenPtr := flag.String("t", "", "Discord API token")
	flag.Parse()
	token = *tokenPtr
	if token == "" {
		envToken, foundInEnv := os.LookupEnv("DISCORD_RMUBOT_TOKEN")
		if !foundInEnv {
			log.Fatal("Token not set in env 'DISCORD_RMUBOT_TOKEN' or provided on command line '-t TOKEN'")
		}
		token = envToken
	}

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(dispatch)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	setupDice()
	rollsByUser = make(map[string][]int)

	fmt.Println("the bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func dispatch(s *discordgo.Session, m *discordgo.MessageCreate) {
	// don't respond to myself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	tokens := strings.Split(m.Content, " ")
	command := strings.ToLower(tokens[0])
	//args := tokens[1:] // let them handle their own args
	// It's still necessary to pass the session to the handler so they can send messages, unless
	// they are refactored to return the response as a string for the dispatcher to send.
	// And the message so they can extract sender ID for things like the averagesHandler.
	switch command {
	case "!roll", "!r":
		//rollHandler(s, m)
		rwaHandler(s, m)
	case "!dice", "!d":
		generalDiceHandler(s, m)
	case "!dhelp":
		diceHelpHandler(s, m)
	case "!say":
		sayHandler(s, m)
	case "!avg":
		averagesHandler(s, m)
	case "!reset":
		resetHandler(s, m)
	case "!help":
		helpHandler(s, m)
	default:
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command: %s", command))
		if err != nil {
			log.Println(err)
		}
	}

}
