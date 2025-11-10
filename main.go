package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jcheng31/diceroller/dice"
	"github.com/jcheng31/diceroller/roller"
)

var d100 dice.Die

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

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	setupDice()

	fmt.Println("the bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func setupDice() {
	src := rand.NewSource(time.Now().UnixNano())
	random := roller.WithRandomSource(src)
	d100 = dice.Regular(random, 100)
}

func rollOE() (int, string) {
	return rollOEHelper(true, true, "", d100)
}

func rollOEHelper(up, down bool, details string, d100 dice.Die) (int, string) {
	roll := d100.RollN(1)
	rollVal := roll.Total
	var newDetails string
	if details == "" {
		newDetails = fmt.Sprintf("%d", rollVal)
	} else {
		newDetails = fmt.Sprintf("%s %d", details, rollVal)
	}
	if up && rollVal >= 96 {
		newTotal, newDetails := rollOEHelper(true, false, details, d100)
		return newTotal + rollVal, fmt.Sprintf("%d %s", rollVal, newDetails)
	}

	if down && rollVal <= 5 {
		newTotal, newDetails := rollOEHelper(true, false, details, d100)
		return rollVal - newTotal, fmt.Sprintf("%d %s", rollVal, newDetails)

	}
	return rollVal, newDetails
}

func roll(s *discordgo.Session, m *discordgo.MessageCreate) {
	// don't respond to myself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.HasPrefix(m.Content, "!roll") {

		diceResult, details := rollOE()

		log.Println(fmt.Sprintf("Result: [%s] %d ", details, diceResult))

		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: [%s] %d", details, diceResult))
		if err != nil {
			log.Println(err)
		}
	}
}
