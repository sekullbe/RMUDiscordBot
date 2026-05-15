package main

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"

	jdice "github.com/justinian/dice"
)

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func rollHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := optionMap(i)

	modifier := 0
	flat := false
	if opt, ok := opts["modifier"]; ok {
		modifier = int(opt.IntValue())
	}
	if opt, ok := opts["flat"]; ok {
		flat = opt.BoolValue()
	}

	diceResult, details := doRoll(flat)
	rollsByUser[userID(i)] = append(rollsByUser[userID(i)], diceResult)
	if details != "" {
		details = fmt.Sprintf("[%s]", details)
	}
	diceResult += modifier

	who := whoIsThis(i)
	var msg string
	switch {
	case modifier > 0:
		msg = fmt.Sprintf("%s rolls: %s + %d = %d", who, details, modifier, diceResult)
	case modifier < 0:
		msg = fmt.Sprintf("%s rolls: %s %d = %d", who, details, modifier, diceResult)
	default:
		msg = fmt.Sprintf("%s rolls: %s %d", who, details, diceResult)
	}
	respond(s, i, msg)
}

func generalDiceHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	expr := optionMap(i)["expression"].StringValue()
	res, _, err := jdice.Roll(expr)
	if err != nil {
		respond(s, i, "Cannot parse requested dice")
		return
	}
	respond(s, i, res.String())
}

func sayHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	diceResult, _ := doRoll(false)
	rollsByUser[userID(i)] = append(rollsByUser[userID(i)], diceResult)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%d", diceResult),
			TTS:     true,
		},
	})
	if err != nil {
		log.Println(err)
	}
}

func averagesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	avgAll := averageSlice(allRolls)
	avgUser := averageSlice(rollsByUser[userID(i)])
	respondEphemeral(s, i, fmt.Sprintf("All: %.1f  You: %.1f", avgAll, avgUser))
}

func resetHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	allRolls = []int{}
	clear(rollsByUser)
	respond(s, i, "Reset averages")
}

func initiativeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	subcommand := i.ApplicationCommandData().Options[0]
	opts := optionMapFromOptions(subcommand.Options)
	who := whoIsThis(i)
	uid := userID(i)

	switch subcommand.Name {
	case "quick":
		modifier := 0
		if opt, ok := opts["modifier"]; ok {
			modifier = int(opt.IntValue())
		}
		result := d10.RollN(2)
		respond(s, i, fmt.Sprintf("%s: %v + %d = %d", who, result.Rolls, modifier, result.Total+modifier))

	case "reg":
		if initiatives[i.ChannelID] == nil {
			initiatives[i.ChannelID] = make(map[string]initStore)
		}
		mod := 0
		if opt, ok := opts["modifier"]; ok {
			mod = int(opt.IntValue())
		}
		initiatives[i.ChannelID][uid] = initStore{id: uid, name: who, mod: mod, isPC: true}
		respond(s, i, fmt.Sprintf("Registered initiative modifier for %s", who))

	case "addnpc":
		if initiatives[i.ChannelID] == nil {
			initiatives[i.ChannelID] = make(map[string]initStore)
		}
		name := opts["name"].StringValue()
		mod := 0
		if opt, ok := opts["modifier"]; ok {
			mod = int(opt.IntValue())
		}
		initiatives[i.ChannelID][name] = initStore{id: name, name: name, mod: mod, isPC: false}
		respond(s, i, fmt.Sprintf("Registered initiative modifier for %s", name))

	case "rem":
		if initiatives[i.ChannelID] != nil {
			delete(initiatives[i.ChannelID], uid)
		}
		respond(s, i, fmt.Sprintf("Removed initiative modifier for %s", who))

	case "remnpc":
		name := opts["name"].StringValue()
		if initiatives[i.ChannelID] != nil {
			delete(initiatives[i.ChannelID], name)
		}
		respond(s, i, fmt.Sprintf("Removed initiative modifier for %s", name))

	case "roll":
		respond(s, i, rollRound(i.ChannelID))

	case "list":
		msg := "Registered initiative modifiers:\n"
		for _, is := range initiatives[i.ChannelID] {
			pc := "NPC"
			if is.isPC {
				pc = "PC"
			}
			msg += fmt.Sprintf("%s (%s): %d\n", is.name, pc, is.mod)
		}
		respond(s, i, msg)

	case "clearnpc":
		for id, is := range initiatives[i.ChannelID] {
			if !is.isPC {
				delete(initiatives[i.ChannelID], id)
			}
		}
		respond(s, i, "Removed all NPCs from initiative list")

	case "help":
		respondEphemeral(s, i, `Initiative System:
/init quick [modifier] - Roll 2d10 initiative with an optional modifier
/init reg [modifier] - Register your PC's initiative modifier
/init addnpc name [modifier] - Register an NPC's initiative modifier
/init rem - Remove your PC from the initiative list
/init remnpc name - Remove an NPC from the initiative list
/init roll - Roll initiative for all registered characters
/init list - List all registered initiative modifiers
/init clearnpc - Remove all NPCs from the initiative list (e.g. after a combat)`)
	}
}

func rollRound(channelID string) string {
	initRolls := make(map[string]int)
	for _, is := range initiatives[channelID] {
		initRolls[is.name] = d10.RollN(2).Total + is.mod
	}

	type nameRoll struct {
		name string
		roll int
	}
	rolls := make([]nameRoll, 0, len(initRolls))
	for name, roll := range initRolls {
		rolls = append(rolls, nameRoll{name: name, roll: roll})
	}

	slices.SortFunc(rolls, func(a, b nameRoll) int {
		return b.roll - a.roll
	})

	var response strings.Builder
	response.WriteString("Initiative Order:\n")
	for _, nr := range rolls {
		response.WriteString(fmt.Sprintf("%s: %d\n", nr.name, nr.roll))
	}
	return response.String()
}

func helpHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondEphemeral(s, i, `RMU Bot Commands:
/roll - make an open ended d100 roll. Add modifier option for a bonus, flat:true for a plain roll.
/init quick [modifier] - roll initiative (2d10)
/init help - display initiative system help
/dice expression - roll general dice. /dhelp for format. Not tracked for averages.
/avg - display average RM d100 rolls
/reset - reset all averages`)
}

func diceHelpHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondEphemeral(s, i, `Dice Roll Formatting:
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

func whoIsThis(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		if i.Member.User != nil {
			if i.Member.User.GlobalName != "" {
				return i.Member.User.GlobalName
			}
			return i.Member.User.Username
		}
	}
	if i.User != nil {
		if i.User.GlobalName != "" {
			return i.User.GlobalName
		}
		return i.User.Username
	}
	return "Unknown"
}

func userID(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func optionMap(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	return optionMapFromOptions(i.ApplicationCommandData().Options)
}

func optionMapFromOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		m[opt.Name] = opt
	}
	return m
}
