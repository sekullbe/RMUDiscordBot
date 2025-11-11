package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func sendMessage(session *discordgo.Session, channelID string, message string) {
	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Println(err)
	}
}

func rollHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	diceResult, details := doRoll(m.Content)
	rollsByUser[m.Author.ID] = append(rollsByUser[m.Author.ID], diceResult)
	if details != "" {
		details = fmt.Sprintf("[%s]", details)
	}
	sendMessage(s, m.ChannelID, fmt.Sprintf("Result: %s %d", details, diceResult))
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
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(`RMU Bot Commands:
!roll, !r - make an open ended d100 roll. Any numbers after this will be treated as modifiers.
!roll flat - make a plain d100 roll
!avg - display average dice rolls
!reset - reset all averages`))
	if err != nil {
		log.Println(err)
	}
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
