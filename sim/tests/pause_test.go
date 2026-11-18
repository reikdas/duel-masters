package tests

import (
	"duel-masters/game/match"
	"duel-masters/tests/scenario"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const creepingPlague = "452aead9-2a65-46b1-84b2-383fe99ddc5f"

func nextPause(t *testing.T, pauses <-chan match.Pause) match.Pause {
	t.Helper()

	select {
	case p := <-pauses:
		return p
	case <-time.After(2 * time.Second):
		t.Fatal("the match did not pause")
		return match.Pause{}
	}
}

func send(t *testing.T, m *match.Match, p *match.PlayerReference, msg map[string]any) {
	t.Helper()

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	m.Parse(p.Socket, data)
}

func TestPauseHook(t *testing.T) {
	pauses := make(chan match.Pause, 8)
	scn := scenario.New(
		scenario.WithDeck(scenario.DeckEntry{UID: creepingPlague, Count: 40}),
		scenario.WithPauses(pauses),
	)
	defer scn.Match.Dispose()

	first := nextPause(t, pauses)
	active := scn.Match.CurrentPlayer()
	assert.False(t, first.Prompt, "the game starts idle")
	assert.Same(t, active.Player, first.Player, "the first pause is for the player who goes first")

	hand, err := active.Player.Container(match.HAND)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hand), 2)
	manaCard, spellCard := hand[0], hand[1]

	send(t, scn.Match, active, map[string]any{"header": "add_to_playzone", "virtualId": spellCard.ID})
	p := nextPause(t, pauses)
	assert.False(t, p.Prompt, "a card the player cannot afford is rejected without a prompt")
	assert.Same(t, active.Player, p.Player)

	send(t, scn.Match, active, map[string]any{"header": "add_to_manazone", "virtualId": manaCard.ID})
	p = nextPause(t, pauses)
	assert.False(t, p.Prompt)
	assert.Same(t, active.Player, p.Player)

	send(t, scn.Match, active, map[string]any{"header": "add_to_playzone", "virtualId": spellCard.ID})
	p = nextPause(t, pauses)
	assert.True(t, p.Prompt, "casting the spell opens the mana prompt")
	assert.Same(t, active.Player, p.Player)

	send(t, scn.Match, active, map[string]any{"header": "action", "cards": []string{manaCard.ID}})
	p = nextPause(t, pauses)
	assert.False(t, p.Prompt, "paying the mana resolves the spell and the match is idle again")
	assert.Same(t, active.Player, p.Player)
	assert.True(t, active.Player.HasCard(match.GRAVEYARD, spellCard.ID))

	send(t, scn.Match, active, map[string]any{"header": "end_turn"})
	p = nextPause(t, pauses)
	assert.False(t, p.Prompt)
	assert.Same(t, scn.Match.Opponent(active.Player), p.Player, "ending the turn passes the decision to the opponent")

	select {
	case extra := <-pauses:
		t.Fatalf("one message produced more than one pause: %+v", extra)
	default:
	}
}
