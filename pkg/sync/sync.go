package sync

import (
	"github.com/mona-actions/gh-migrate-teams/internal/team"
	"github.com/pterm/pterm"
)

func SyncTeams() {
	// Get all teams from source organization
	teamsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching teams from organization...")
	teams := team.GetSourceOrganizationTeams()
	teamsSpinnerSuccess.Success()

	// Create teams in target organization
	createTeamsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating teams in target organization...")
	for _, team := range teams {
		team.CreateTeam()
	}
	createTeamsSpinnerSuccess.Success()
}
