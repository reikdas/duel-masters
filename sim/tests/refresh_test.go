package tests

import (
	"duel-masters/game/fx"
	"duel-masters/game/match"
	"duel-masters/tests/scenario"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const rumblesaurQ = "eaa15452-23bb-4fad-a557-9df3c9752292"

// Rumblesaur Q gives every Survivor, itself included, speed attacker through a
// persistent effect. The effect starts when Rumblesaur Q enters the battle zone
// but only grants the ability the next time an event runs. The refresh after
// the summon must run it, so the creature can attack as soon as the summon is
// done, as it can in the real game.
func TestStateIsCurrentWhenTheMatchIsIdle(t *testing.T) {
	pauses := make(chan match.Pause, 8)
	scn := scenario.New(scenario.WithPauses(pauses))
	defer scn.Match.Dispose()

	nextPause(t, pauses)
	active := scn.Match.CurrentPlayer()

	creature, err := active.Player.SpawnCard(rumblesaurQ, match.HAND)
	require.NoError(t, err)

	mana := []string{}
	for range creature.ManaCost {
		c, err := active.Player.SpawnCard(rumblesaurQ, match.MANAZONE)
		require.NoError(t, err)
		mana = append(mana, c.ID)
	}

	send(t, scn.Match, active, map[string]any{"header": "add_to_playzone", "virtualId": creature.ID})
	require.True(t, nextPause(t, pauses).Prompt, "summoning asks which mana to tap")

	send(t, scn.Match, active, map[string]any{"header": "action", "cards": mana})
	p := nextPause(t, pauses)
	require.False(t, p.Prompt)
	require.Equal(t, match.BATTLEZONE, creature.Zone)

	assert.False(t, fx.HasSummoningSickness(creature), "Rumblesaur Q has speed attacker as soon as it is summoned")
}
