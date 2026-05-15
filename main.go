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
var d10 dice.Die

var allRolls []int
var rollsByUser map[string][]int

type initStore struct {
	id   string
	name string
	mod  int
	isPC bool
}

var initiatives map[string]map[string]initStore

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "roll",
		Description: "Make an open ended d100 roll",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "modifier",
				Description: "Modifier to add to the roll",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "flat",
				Description: "Make a plain d100 roll without open-ending",
				Required:    false,
			},
		},
	},
	{
		Name:        "dice",
		Description: "Roll general dice (not tracked for averages)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "expression",
				Description: "Dice expression (e.g. 4d6kh3+4)",
				Required:    true,
			},
		},
	},
	{
		Name:        "say",
		Description: "Roll dice and announce result with text-to-speech",
	},
	{
		Name:        "avg",
		Description: "Display your average RM d100 rolls (only visible to you)",
	},
	{
		Name:        "reset",
		Description: "Reset all roll averages",
	},
	{
		Name:        "help",
		Description: "Display bot commands (only visible to you)",
	},
	{
		Name:        "dhelp",
		Description: "Display dice roll formatting help (only visible to you)",
	},
	{
		Name:        "init",
		Description: "Initiative system",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "quick",
				Description: "Roll 2d10 initiative with an optional modifier",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "modifier",
						Description: "Initiative modifier",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "roll",
				Description: "Roll initiative for all registered characters",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "reg",
				Description: "Register your PC's initiative modifier",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "modifier",
						Description: "Initiative modifier",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "addnpc",
				Description: "Register an NPC's initiative modifier",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "NPC name (no spaces)",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "modifier",
						Description: "Initiative modifier",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "rem",
				Description: "Remove your PC from initiative",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remnpc",
				Description: "Remove an NPC from initiative",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "NPC name",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all registered initiative modifiers",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "clearnpc",
				Description: "Remove all NPCs from initiative list",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "help",
				Description: "Display initiative system help (only visible to you)",
			},
		},
	},
}

func main() {

	var token string
	var guildID string
	tokenPtr := flag.String("t", "", "Discord API token")
	guildPtr := flag.String("g", "", "Guild ID for instant guild-scoped commands (omit for global)")
	flag.Parse()
	token = *tokenPtr
	guildID = *guildPtr

	if token == "" {
		envToken, foundInEnv := os.LookupEnv("DISCORD_RMUBOT_TOKEN")
		if !foundInEnv {
			log.Fatal("Token not set in env 'DISCORD_RMUBOT_TOKEN' or provided on command line '-t TOKEN'")
		}
		token = envToken
	}

	if guildID == "" {
		guildID = os.Getenv("DISCORD_RMUBOT_GUILD")
	}

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(interactionDispatch)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	setupDice()
	rollsByUser = make(map[string][]int)
	initiatives = make(map[string]map[string]initStore)

	for _, cmd := range commands {
		if _, err := sess.ApplicationCommandCreate(sess.State.User.ID, guildID, cmd); err != nil {
			log.Fatalf("Cannot create command %q: %v", cmd.Name, err)
		}
	}

	fmt.Println("the bot is online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func interactionDispatch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "roll":
		rollHandler(s, i)
	case "dice":
		generalDiceHandler(s, i)
	case "say":
		sayHandler(s, i)
	case "avg":
		averagesHandler(s, i)
	case "reset":
		resetHandler(s, i)
	case "help":
		helpHandler(s, i)
	case "dhelp":
		diceHelpHandler(s, i)
	case "init":
		initiativeHandler(s, i)
	}
}
