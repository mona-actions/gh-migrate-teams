package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mona-actions/gh-migrate-teams/internal/api"
	"github.com/mona-actions/gh-migrate-teams/internal/team"
	"github.com/spf13/viper"
)

func CreateCSVs() {
	// Get team membership
	var teams []team.Team
	teams = api.GetSourceOrganizationTeams()

	fmt.Println("Found", len(teams), "teams")

	// Create Membership CSV
	createMembershipCSV(teams)

	// Create Repository CSV
	createRepositoryCSV(teams)
}

func createMembershipCSV(teams []team.Team) {
	// Create team membership csv
	file, err := os.Create(viper.GetString("OUTPUT_FILE") + "-team-membership.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Initialize csv writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write team membership to csv
	for _, team := range teams {
		for _, member := range team.Members {
			writer.Write([]string{team.Name, member.Login, member.Email})
		}
	}
}

func createRepositoryCSV(teams []team.Team) {
	// Create team membership csv
	file, err := os.Create(viper.GetString("OUTPUT_FILE") + "-team-repository-permissions.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Initialize csv writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write team membership to csv
	for _, team := range teams {
		for _, repository := range team.Repositories {
			writer.Write([]string{team.Name, repository.Name, repository.Permission})
		}
	}
}
