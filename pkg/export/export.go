package export

import (
	"encoding/csv"
	"os"

	"github.com/mona-actions/gh-migrate-teams/internal/repository"
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

	// Get all repositories from source organization
	repositoriesSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching repositories from organization...")
	repositories := repository.GetSourceOrganizationRepositories()
	repositoriesSpinnerSuccess.Success()

	// Create repository collaborator csv
	createCSVCollaboratorsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating repository collaborator csv...")
	createCSV(repositories.ExportRepositoryCollaborators(), viper.GetString("OUTPUT_FILE")+"-repository-collaborators.csv")
	createCSVCollaboratorsSpinnerSuccess.Success()
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
