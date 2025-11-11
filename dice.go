package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jcheng31/diceroller/dice"
	"github.com/jcheng31/diceroller/roller"
)

func doRoll(command string) (int, string) {

	var diceResult int
	var details string

	var diceDetails string
	_, err := fmt.Sscanf(command, "!rollHandler %s", &diceDetails)
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
