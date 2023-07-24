package export

import (
	"encoding/csv"
	"os"

	"github.com/mona-actions/gh-migrate-teams/internal/api"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

func CreateCSVs() {
	// Get all teams from source organization
	teamsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching teams from organization...")
	teams := api.GetSourceOrganizationTeams()
	teamsSpinnerSuccess.Success()

	// Get all team members from source organization
	membershipsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching team memberships...")
	memberships := getMemberships(teams)
	membershipsSpinnerSuccess.Success()

	// Get all team repositories from source organization
	repositoriesSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching team repositories...")
	repositories := getRepositories(teams)
	repositoriesSpinnerSuccess.Success()

	// Create team membership csv
	createCSVMembershipsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating team membership csv...")
	createCSV(memberships, viper.GetString("OUTPUT_FILE")+"-team-membership.csv")
	createCSVMembershipsSpinnerSuccess.Success()

	// Create team repository csv
	createCSVRepositoriesSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating team repository csv...")
	createCSV(repositories, viper.GetString("OUTPUT_FILE")+"-team-repositories.csv")
	createCSVRepositoriesSpinnerSuccess.Success()
}

func getMemberships(teams []string) [][]string {
	memberships := make([][]string, 0)
	for _, team := range teams {
		members := api.GetTeamMemberships(team)
		for _, member := range members {
			memberships = append(memberships, []string{team, member[0], member[1]})
		}
	}
	return memberships
}

func getRepositories(teams []string) [][]string {
	repositories := make([][]string, 0)
	for _, team := range teams {
		repos := api.GetTeamRepositories(team)
		for _, repo := range repos {
			repositories = append(repositories, []string{team, repo[0], repo[1]})
		}
	}
	return repositories
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
