package export

import (
	"encoding/csv"
	"os"

	"github.com/mona-actions/gh-migrate-teams/internal/team"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

func CreateCSVs() {
	// Get all teams from source organization
	teamsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching teams from organization...")
	teams := team.GetSourceOrganizationTeams()
	teamsSpinnerSuccess.Success()

	// Create team membership csv
	createCSVMembershipsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating team membership csv...")
	createCSV(teams.ExportTeamMemberships(), viper.GetString("OUTPUT_FILE")+"-team-membership.csv")
	createCSVMembershipsSpinnerSuccess.Success()

	// Create team repository csv
	createCSVRepositoriesSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating team repository csv...")
	createCSV(teams.ExportTeamRepositories(), viper.GetString("OUTPUT_FILE")+"-team-repositories.csv")
	createCSVRepositoriesSpinnerSuccess.Success()
}

func createCSV(data [][]string, filename string) {
	// Create team membership csv
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Initialize csv writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write team memberships to csv

	for _, line := range data {
		writer.Write(line)
	}
}
