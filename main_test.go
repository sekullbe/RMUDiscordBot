package main

import (
	"testing"

	"github.com/jcheng31/diceroller/dice"
	"github.com/jcheng31/diceroller/roller"
	"github.com/stretchr/testify/assert"
)

func Test_rollOEHelper(t *testing.T) {

	predetermined := roller.WithSequence([]int{2, 100, 50, 6, 50, 98, 50})
	d100 := dice.Regular(predetermined, 100)

	roll, details := rollOEHelper(true, true, "", d100)
	assert.Equal(t, -148, roll)
	assert.Equal(t, "2 100 50", details)

	roll, details = rollOEHelper(true, true, "", d100)
	assert.Equal(t, 6, roll)
	assert.Equal(t, "6", details)

	roll, details = rollOEHelper(true, true, "", d100)
	assert.Equal(t, 50, roll)
	assert.Equal(t, "50", details)

	roll, details = rollOEHelper(true, true, "", d100)
	assert.Equal(t, 148, roll)
	assert.Equal(t, "98 50", details)
}
