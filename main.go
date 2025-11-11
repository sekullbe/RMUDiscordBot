package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	sess.AddHandler(roll)
	sess.AddHandler(say)
	sess.AddHandler(averages)
	sess.AddHandler(help)

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
