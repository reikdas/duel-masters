package tests

import (
	"duel-masters/game/cards"
	"duel-masters/game/match"
	"duel-masters/tests/scenario"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// tenDistinctCards gives a deck whose order is visible, unlike the default
// scenario deck of 40 copies of one card.
func tenDistinctCards() []scenario.DeckEntry {
	uids := make([]string, 0, len(cards.DM01))
	for uid := range cards.DM01 {
		uids = append(uids, uid)
	}
	slices.Sort(uids)

	entries := make([]scenario.DeckEntry, 0, 10)
	for _, uid := range uids[:10] {
		entries = append(entries, scenario.DeckEntry{UID: uid, Count: 4})
	}
	return entries
}

// openingState lists the cards in each player's hand, shields and deck, in
// order, by card type.
func openingState(t *testing.T, seed int64) []string {
	t.Helper()

	scn := scenario.New(scenario.WithSeed(seed), scenario.WithDeck(tenDistinctCards()...))
	defer scn.Match.Dispose()

	state := []string{}
	for _, p := range []*match.Player{scn.Match.Player1.Player, scn.Match.Player2.Player} {
		for _, zone := range []string{match.HAND, match.SHIELDZONE, match.DECK} {
			zoneCards, err := p.Container(zone)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range zoneCards {
				state = append(state, zone+":"+c.ImageID)
			}
		}
	}
	return state
}

func TestSeedReproducesShuffle(t *testing.T) {
	t.Run("same seed deals the same game", func(t *testing.T) {
		assert.Equal(t, openingState(t, 42), openingState(t, 42))
	})

	t.Run("different seeds deal different games", func(t *testing.T) {
		assert.NotEqual(t, openingState(t, 42), openingState(t, 43))
	})
}
