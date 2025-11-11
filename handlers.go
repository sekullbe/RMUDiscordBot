package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func rollHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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
!roll, !r - make an open ended d100 roll
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
