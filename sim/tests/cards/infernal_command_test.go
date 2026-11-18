package cards

import (
	"duel-masters/game/civ"
	"duel-masters/game/match"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	infernalCommandUID      = "08f2ebba-8996-405d-be3b-deca75936afc"
	infernalCommandPlainUID = "af3bc221-1cc2-4f58-83ea-2673ac2c66c5" // Immortal Baron, Vorg (2000, no ability)
	infernalCommandSetupSrc = "infernal_command_test_setup"
)

func TestInfernalCommand(t *testing.T) {
	t.Run("printed characteristics", func(t *testing.T) {
		_, player, _ := setupDuel(t)

		spell, err := player.Player.SpawnCard(infernalCommandUID, match.HAND)
		require.NoError(t, err)

		assert.Equal(t, "Infernal Command", spell.Name)
		assert.Equal(t, 1, spell.ManaCost)
		assert.Equal(t, []string{civ.Darkness}, spell.Civs)
		assert.Equal(t, []string{civ.Darkness}, spell.ManaRequirement)
	})

	t.Run("a creature that can only attack creatures must attack a tapped one", func(t *testing.T) {
		scn, player, opponent := setupDuel(t)
		putCardInBattlezone(t, scn, opponent.Player, poltalesterUID, infernalCommandSetupSrc)
		target := putCardInBattlezone(t, scn, player.Player, infernalCommandPlainUID, infernalCommandSetupSrc)
		passTurnToSelf(t, scn, player, opponent)

		castSpell(t, scn, player, infernalCommandUID)
		require.NoError(t, scn.ActionEndTurn(player))
		require.False(t, scn.Match.IsPlayerTurn(player.Player))

		target.Tapped = true
		require.NoError(t, scn.ActionEndTurn(opponent))
		assert.True(t, scn.Match.IsPlayerTurn(opponent.Player), "Poltalester can attack the tapped creature, so the turn cannot end")
	})

	t.Run("a creature with no legal attack does not hold up the turn", func(t *testing.T) {
		scn, player, opponent := setupDuel(t)
		putCardInBattlezone(t, scn, opponent.Player, poltalesterUID, infernalCommandSetupSrc)
		untouchable := putCardInBattlezone(t, scn, player.Player, koocPollonUID, infernalCommandSetupSrc)
		passTurnToSelf(t, scn, player, opponent)

		castSpell(t, scn, player, infernalCommandUID)
		require.NoError(t, scn.ActionEndTurn(player))
		require.False(t, scn.Match.IsPlayerTurn(player.Player))

		// Poltalester can't attack players, and the only tapped creature it
		// could attack can't be attacked, so it is unable to attack.
		untouchable.Tapped = true
		require.NoError(t, scn.ActionEndTurn(opponent))
		assert.True(t, scn.Match.IsPlayerTurn(player.Player), "the turn ends because Poltalester is unable to attack")
	})
}
