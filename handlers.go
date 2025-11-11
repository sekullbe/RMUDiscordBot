package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	jdice "github.com/justinian/dice"
)

func sendMessage(session *discordgo.Session, channelID string, message string) {
	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Println(err)
	}
}

func rwaHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, rollWithArguments(m))
}

func rollWithArguments(m *discordgo.MessageCreate) string {

	parts := strings.Split(m.Content, " ")
	modifier := 0
	for _, p := range parts[1:] {
		mod, err := strconv.Atoi(p)
		if err == nil {
			modifier += mod
		}
	}

	// roll the dice
	diceResult, details := doRoll(m.Content)
	rollsByUser[m.Author.ID] = append(rollsByUser[m.Author.ID], diceResult)
	if details != "" {
		details = fmt.Sprintf("[%s]", details)
	}

	diceResult += modifier

	switch {
	case modifier > 0:
		return fmt.Sprintf("Result: %s +%d = +%d", details, modifier, diceResult)
	case modifier < 0:
		return fmt.Sprintf("Result: %s %d = %d", details, modifier, diceResult)
	default:
		return fmt.Sprintf("Result: %s %d", details, diceResult)
	}
}

func generalDiceHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, generalDice(m))
}

func generalDice(m *discordgo.MessageCreate) string {

	res, _, err := jdice.Roll(m.Content)
	if err != nil {
		return "Cannot parse requested dice"
	}
	return res.String()
}

func sayHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	diceResult, _ := doRoll(m.Content)
	rollsByUser[m.Author.ID] = append(rollsByUser[m.Author.ID], diceResult)
	_, err := s.ChannelMessageSendTTS(m.ChannelID, fmt.Sprintf("%d", diceResult))
	if err != nil {
		log.Println(err)
	}
}

func averagesHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	avgAll := averageSlice(allRolls)
	avgUser := averageSlice(rollsByUser[m.Author.ID])
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("All: %.1f  You: %.1f", avgAll, avgUser))
	if err != nil {
		log.Println(err)
	}
}

func resetHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	allRolls = []int{}
	clear(rollsByUser)
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Reset averages"))
	if err != nil {
		log.Println(err)
	}
}

func helpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, `RMU Bot Commands:
!roll, !r - make an open ended d100 roll. Any numbers after this will be treated as modifiers.
!roll flat - make a plain d100 roll
!dice, !d - roll general dice. !dhelp for dice format. These dice are not tracked for averages.
!avg - display average RM d100  rolls
!reset - reset all averages`)
}

func diceHelpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, `Dice Roll Formatting:
Standard: xdy[[k|d][h|l]z][+/-c] - rolls and sums x y-sided dice, keeping or dropping the lowest or highest z dice and optionally adding or subtracting c. Example: 4d6kh3+4
Fudge: xdf[+/-c] - rolls and sums x fudge dice (Dice that returns numbers between -1 and 1), and optionally adding or subtracting c. Example: 4df+4
Versus: xdy[e|r]vt - rolls x y-sided dice, counting the number that roll t or greater.
EotE: xc [xc ...] - rolls x dice of color c (b, blk, g, p, r, w, y) and returns the aggregate result.
Adding an e to the Versus rolls above makes dice 'explode' - Dice are rerolled and have the rolled value added to their total when they roll a y. Adding an r makes dice rolling a y add another die to the pool instead.`)
}

// helper fns -------------------------

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
