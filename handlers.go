package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
}

func reset(s *discordgo.Session, m *discordgo.MessageCreate) {
	// don't respond to myself
	if m.Author.ID == s.State.User.ID {
		return
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
