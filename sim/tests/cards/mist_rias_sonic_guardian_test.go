package cards

import (
	"duel-masters/server"
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mistRiasUID    = "6a909066-2011-4de7-9a97-cab333dca4cf"
	mistRiasVorg   = "af3bc221-1cc2-4f58-83ea-2673ac2c66c5" // Immortal Baron, Vorg (2000)
	mistRiasSource = "mist_rias_test_setup"
)

// "Whenever another creature is put into the battle zone, you may draw a
// card."
//
// Mist Rias asks while the opponent's creature is still being put into the
// battle zone, partway through the opponent's summon. The board the player
// is sent before the question must already show that creature, or they decide
// without seeing what arrived.
func TestMistRiasAsksWithTheArrivalOnTheBoard(t *testing.T) {
	scn, player, opponent := setupDuel(t)
	putCardInBattlezone(t, scn, player.Player, mistRiasUID, mistRiasSource)
	require.NoError(t, scn.ActionEndTurn(player))

	start, err := scn.MessageCount(player)
	require.NoError(t, err)
	arrival := summonWithOwnMana(t, scn, opponent, mistRiasVorg)

	question, err := scn.WaitForAction(player, start)
	require.NoError(t, err, "Mist Rias should offer a draw")
	assert.Equal(t, "question", question.ActionType)

	messages, err := scn.Messages(player, start)
	require.NoError(t, err)
	var board *server.MatchState
	for _, raw := range messages {
		var msg struct {
			Header     string             `json:"header"`
			ActionType string             `json:"actionType"`
			State      *server.MatchState `json:"state"`
		}
		require.NoError(t, json.Unmarshal([]byte(raw), &msg))
		if msg.Header == "action" && msg.ActionType == "question" {
			break
		}
		if msg.Header == "state_update" {
			board = msg.State
		}
	}
	require.NotNil(t, board, "no board was sent between the summon and the question")
	assert.True(t, slices.ContainsFunc(board.Opponent.Battlezone, func(c server.CardState) bool { return c.CardID == arrival.ID }),
		"the last board sent before the question doesn't show the creature that arrived")

	cancelInTurn(t, scn, player)
	settleTurn(t, scn)
}
