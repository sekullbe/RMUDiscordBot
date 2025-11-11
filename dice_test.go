package main

import (
	"testing"

	"github.com/jcheng31/diceroller/dice"
	"github.com/jcheng31/diceroller/roller"
	"github.com/stretchr/testify/assert"
)

func Test_rollOEHelper(t *testing.T) {

	predetermined := roller.WithSequence([]int{2, 100, 50, 6, 50, 98, 50, 2, 100, 100, 50, 100, 100, 50})
	d100 := dice.Regular(predetermined, 100)

	roll, details := rollOEHelper(true, false, "", d100)
	assert.Equal(t, -148, roll)
	assert.Equal(t, "2 -100 -50", details)

	roll, details = rollOEHelper(true, false, "", d100)
	assert.Equal(t, 6, roll)
	assert.Equal(t, "6", details)

	roll, details = rollOEHelper(true, false, "", d100)
	assert.Equal(t, 50, roll)
	assert.Equal(t, "50", details)

	roll, details = rollOEHelper(true, false, "", d100)
	assert.Equal(t, 148, roll)
	assert.Equal(t, "98 50", details)

	roll, details = rollOEHelper(true, false, "", d100)
	assert.Equal(t, -248, roll)
	assert.Equal(t, "2 -100 -100 -50", details)

	roll, details = rollOEHelper(true, false, "", d100)
	assert.Equal(t, 250, roll)
	assert.Equal(t, "100 100 50", details)
}

func Test_doRoll(t *testing.T) {

	predetermined := roller.WithSequence([]int{100, 50, 99, 100, 50, 100, 50, 1, 50})
	d100 = dice.Regular(predetermined, 100)

	roll, details := doRoll("!roll")
	assert.Equal(t, 150, roll)
	assert.Equal(t, "100 50", details)

	roll, details = doRoll("!roll flat")
	assert.Equal(t, 99, roll)
	assert.Equal(t, "", details)

	roll, details = doRoll("!roll dfjhjkdfhg")
	assert.Equal(t, 150, roll)
	assert.Equal(t, "100 50", details)

	roll, details = doRoll("!roll  ")
	assert.Equal(t, 150, roll)
	assert.Equal(t, "100 50", details)

	roll, details = doRoll("!roll  ")
	assert.Equal(t, -49, roll)
	assert.Equal(t, "1 -50", details)

}
