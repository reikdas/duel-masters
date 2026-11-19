package fx

import (
	"duel-masters/game/match"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A confirmed attack is reported in the chat, naming the attacker and what it
// attacks, before the blocks and shield triggers it leads to. Without it, a
// defender whose creatures are all tapped can't tell which creature attacked.
func TestAttacksAreReportedInChat(t *testing.T) {

	t.Run("an attack on the opponent names the attacker and the opponent", func(t *testing.T) {
		scn, player, opponent := attackStepScenario(t)
		attacker := putInBattleZone(t, scn, player, attackStepAttackerUID)
		passTurn(t, scn, player, opponent)

		shields, err := opponent.Player.Container(match.SHIELDZONE)
		require.NoError(t, err)
		start, err := scn.MessageCount(opponent)
		require.NoError(t, err)

		_, err = scn.ActionAttackPlayer(player, attacker.ID)
		require.NoError(t, err)
		require.NoError(t, scn.ResolveAttack(player, shields[0].ID))
		require.NoError(t, scn.WaitForEventLoop())

		chat, err := scn.ChatMessages(opponent, start)
		require.NoError(t, err)
		report := slices.Index(chat, fmt.Sprintf("%s (2000) is attacking %s", attacker.Name, opponent.Player.Username()))
		broken := slices.IndexFunc(chat, func(m string) bool { return strings.Contains(m, "shields were broken") })
		require.NotEqual(t, -1, report, "the attack should be reported, got %q", chat)
		require.NotEqual(t, -1, broken, "the broken shield should be reported, got %q", chat)
		assert.Less(t, report, broken, "the attack should be reported before its shield breaks")
	})

	t.Run("an attack on a creature names both creatures", func(t *testing.T) {
		scn, player, opponent := attackStepScenario(t)
		attacker := putInBattleZone(t, scn, player, attackStepAttackerUID)
		defender := putInBattleZone(t, scn, opponent, attackStepAttackerUID)
		passTurn(t, scn, player, opponent)

		// Only a tapped creature can be attacked.
		defender.Tapped = true
		start, err := scn.MessageCount(opponent)
		require.NoError(t, err)

		require.NoError(t, scn.ActionAttackCreature(player, attacker.ID, defender.ID))
		require.NoError(t, scn.WaitForEventLoop())

		chat, err := scn.ChatMessages(opponent, start)
		require.NoError(t, err)
		assert.Contains(t, chat, fmt.Sprintf("%s (2000) is attacking %s (2000)", attacker.Name, defender.Name))
	})

	t.Run("a cancelled attack is not reported", func(t *testing.T) {
		scn, player, opponent := attackStepScenario(t)
		attacker := putInBattleZone(t, scn, player, attackStepAttackerUID)
		passTurn(t, scn, player, opponent)

		start, err := scn.MessageCount(opponent)
		require.NoError(t, err)

		_, err = scn.ActionAttackPlayer(player, attacker.ID)
		require.NoError(t, err)
		require.NoError(t, scn.CancelAction(player))
		require.NoError(t, scn.WaitForEventLoop())

		chat, err := scn.ChatMessages(opponent, start)
		require.NoError(t, err)
		for _, m := range chat {
			assert.NotContains(t, m, "is attacking")
		}
	})
}
