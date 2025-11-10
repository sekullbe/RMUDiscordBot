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

		diceResult, details := doRoll(m.Content)
		rollsByUser[m.Author.ID] = append(rollsByUser[m.Author.ID], diceResult)
		if details != "" {
			details = fmt.Sprintf("[%s]", details)
		}
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Result: %s %d", details, diceResult))
		if err != nil {
			log.Println(err)
		}
	}
}

// TODO could probably combine this with roll()
func say(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.HasPrefix(m.Content, "!say") {
		diceResult, details := doRoll(m.Content)
		rollsByUser[m.Author.ID] = append(rollsByUser[m.Author.ID], diceResult)
		if details != "" {
			details = fmt.Sprintf("[%s]", details)
		}
		if strings.HasPrefix(m.Content, "!say") {
			_, err := s.ChannelMessageSendTTS(m.ChannelID, fmt.Sprintf("%d", diceResult))
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func doRoll(command string) (int, string) {

	var diceResult int
	var details string

	var diceDetails string
	_, err := fmt.Sscanf(command, "!roll %s", &diceDetails)
	if err != nil {
		diceResult, details = rollOE()
	} else if diceDetails == "flat" {
		diceResult = d100.RollN(1).Total
		details = ""
	} else {
		diceResult, details = rollOE()
	}

	allRolls = append(allRolls, diceResult)

	return diceResult, details
}

func averages(s *discordgo.Session, m *discordgo.MessageCreate) {
	// don't respond to myself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// a little unclear how this works if used by the same user in more than one server
	if strings.HasPrefix(m.Content, "!avg") {
		avgAll := averageSlice(allRolls)
		avgUser := averageSlice(rollsByUser[m.Author.ID])
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("All: %.1f  You: %.1f", avgAll, avgUser))
		if err != nil {
			log.Println(err)
		}
	}
	if strings.HasPrefix(m.Content, "!reset") {
		allRolls = []int{}
		clear(rollsByUser)
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Reset averages"))
		if err != nil {
			log.Println(err)
		}
	}
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !strings.HasPrefix(m.Content, "!help") {
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(`RMU Bot Commands:
!roll - make an open ended d100 roll
!roll flat - make a plain d100 roll
!avg - display average dice rolls
!reset - reset all averages`))
	if err != nil {
		log.Println(err)
	}
}

func averageSlice(numbers []int) float64 {
	if len(numbers) == 0 {
		return 0
	}
	var sum float64
	for _, number := range numbers {
		sum += float64(number)
	}
	return sum / float64(len(numbers))
}
