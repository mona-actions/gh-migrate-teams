package sync

import (
	"encoding/csv"
	"log"
	"os"

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
		// Map members
		if os.Getenv("GHMT_MAPPING_FILE") != "" {
			team = mapMembers(team)
		}
		team.CreateTeam()
		createTeamsSpinnerSuccess.Success()
	}
}

func mapMembers(team team.Team) team.Team {
	for i, member := range team.Members {
		// Check if member handle is in mapping file
		target_handle, err := getTargetHandle(os.Getenv("GHMT_MAPPING_FILE"), member.Login)
		if err != nil {
			log.Println("Unable to read or open mapping file")
		}
		team.Members[i] = updateMemberHandle(member, member.Login, target_handle)
	}
	return team
}

func updateMemberHandle(member team.Member, source_handle string, target_handle string) team.Member {
	// Update member handles
	if member.Login == source_handle {
		member.Login = target_handle
	}
	return member
}

func getTargetHandle(filename string, source_handle string) (string, error) {
	// Open mapping file
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Parse mapping file
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	// Find target value for source value
	for _, record := range records[1:] {
		if record[0] == source_handle {
			return record[1], nil
		}
	}

	return source_handle, nil
}
