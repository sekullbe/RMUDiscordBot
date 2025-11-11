package main

import (
	"fmt"
	"log"
	"slices"
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
		return fmt.Sprintf("Result: %s + %d = %d", details, modifier, diceResult)
	case modifier < 0:
		return fmt.Sprintf("Result: %s %d = %d", details, modifier, diceResult)
	default:
		return fmt.Sprintf("Result: %s %d", details, diceResult)
	}
}

func generalDiceHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, generalDice(m.Content))
}

func generalDice(req string) string {

	res, _, err := jdice.Roll(req)
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

func initiativeHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Split(m.Content, " ")
	parts = append(parts, "0")
	parts = append(parts, "0")
	parts = append(parts, "0")

	who := whoIsThis(m)

	if parts[1] == "reg" {
		if initiatives[m.ChannelID] == nil {
			initiatives[m.ChannelID] = make(map[string]initStore)
		}

		is := initStore{
			id:   m.Author.ID,
			name: who,
			mod:  0,
			isPC: false,
		}
		if isNumber(parts[2]) {
			is.mod = parseIntOrZero(parts[2])
			is.isPC = true
		} else {
			is.id = parts[2]
			is.name = parts[2]
			is.mod = parseIntOrZero(parts[3])
			is.isPC = false
		}

		initiatives[m.ChannelID][is.id] = is

		sendMessage(s, m.ChannelID, fmt.Sprintf("Registered initiative score for %s", is.name))
	} else if parts[1] == "rem" {
		if initiatives[m.ChannelID] == nil {
			initiatives[m.ChannelID] = make(map[string]initStore)
		}
		delId := ""
		delName := ""
		if isNumber(parts[2]) {
			delId = m.Author.ID
			delName = who
		} else {
			delId = parts[2]
			delName = parts[2]
		}
		delete(initiatives[m.ChannelID], delId)
		sendMessage(s, m.ChannelID, fmt.Sprintf("Removed initiative score for %s", delName))
	} else if parts[1] == "roll" || parts[1] == "round" {
		response := rollRound(s, m)
		sendMessage(s, m.ChannelID, response)
	} else if parts[1] == "clearnpc" {
		for id, is := range initiatives[m.ChannelID] {
			if !is.isPC {
				delete(initiatives[m.ChannelID], id)
			}
		}
		sendMessage(s, m.ChannelID, "Removed all NPCs from initiative list")
	} else if parts[1] == "help" {
		initHelpHandler(s, m)
	} else {
		parts := strings.Split(m.Content, " ")
		modifier := 0
		for _, p := range parts[1:] {
			mod, err := strconv.Atoi(p)
			if err == nil {
				modifier += mod
			}
		}
		result := d10.RollN(2)
		sendMessage(s, m.ChannelID, fmt.Sprintf("%v + %d = %d", result.Rolls, modifier, result.Total+modifier))
	}
}

func rollRound(s *discordgo.Session, m *discordgo.MessageCreate) string {
	initRolls := make(map[string]int)
	for _, is := range initiatives[m.ChannelID] {
		initRolls[is.name] = d10.RollN(2).Total + is.mod
	}

	// Convert map to slice for sorting
	type nameRoll struct {
		name string
		roll int
	}
	rolls := make([]nameRoll, 0, len(initRolls))
	for name, roll := range initRolls {
		rolls = append(rolls, nameRoll{name: name, roll: roll})
	}

	// Sort by roll value descending
	slices.SortFunc(rolls, func(a, b nameRoll) int {
		return b.roll - a.roll
	})

	// Build response message
	var response strings.Builder
	response.WriteString("Initiative Order:\n")
	for _, nr := range rolls {
		response.WriteString(fmt.Sprintf("%s: %d\n", nr.name, nr.roll))
	}
	return response.String()
}

func parseIntOrZero(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func helpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, `RMU Bot Commands:
!roll, !r - make an open ended d100 roll. Any numbers after this will be treated as modifiers.
!roll flat - make a plain d100 roll
!init - roll initiative (2d10). Any numbers after this will be treated as modifiers.
!init help - display initiative system help
!dice, !d - roll general dice. !dhelp for dice format. These dice are not tracked for averages.
!avg - display average RM d100 rolls
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

func initHelpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sendMessage(s, m.ChannelID, `Initiative System:
!init - roll 2d10 initiative. Any numbers after this are treated as modifiers
!init reg [modifier] - Register your PC's initiative with a modifier, or 0 if not specified
!init reg [name] [modifier] - Register a NPC's initiative. Names cannot contain spaces.
!init rem - remove your PC from the initiative list
!init rem [name] - remove a NPC from the initiative list
!init roll, !init round - Roll initiative for all registered characters
!init clearnpc - Remove all NPCs from the initiative list (e.g. after a combat)`)
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

func isNumber(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}

func whoIsThis(m *discordgo.MessageCreate) string {
	// There's a bug in the underlying library so that m.Member.DisplayName() doesn't work.
	name := m.Member.Nick
	if name == "" {
		name = m.Author.DisplayName()
	}
	return name
}
