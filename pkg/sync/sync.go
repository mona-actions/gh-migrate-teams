package sync

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mona-actions/gh-migrate-teams/internal/repository"
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
	}
	createTeamsSpinnerSuccess.Success()
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
			//if filename contains the string gei, return the third column
			if strings.Contains(filename, "gei") {
				return record[2], nil
			}
			return record[1], nil
		}
	}

	return source_handle, nil
}

func SyncTeamsByRepo() {

	teamsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Fetching teams from repository list...")
	repos, err := repository.ParseRepositoryFile(os.Getenv("GHMT_REPO_FILE"))
	teams := []team.Team{}
	teamMap := make(map[string]bool) // Map to track added teams
	totalMembers := 0

	if err != nil {
		log.Println("error while reading repository file - ", err)
		teamsSpinnerSuccess.Fail()
		return
	}
	log.Println("Fetched a total of " + strconv.Itoa(len(repos)) + " repositories from the repository list")

	for _, repo := range repos {
		log.Println("Fetching teams for repository: " + repo)
		// get all teams that have access to the repository
		repoTeams := team.GetRepositoryTeams(repo)
		for _, t := range repoTeams {
			// Check if the team is already in the map
			if _, exists := teamMap[t.Id]; !exists {
				// If the team is not in the map, add it to the map and the teams slice
				teamMap[t.Id] = true
				teams = append(teams, t)
				totalMembers += len(t.Members)
			}
		}
	}
	// Print out how many teams were found:
	teamsSpinnerSuccess.UpdateText("Fetched a total of " + strconv.Itoa(len(teams)) + " teams with total of " + strconv.Itoa(totalMembers) + " members from the repository list")

	if len(teams) == 0 {
		teamsSpinnerSuccess.Fail()
		log.Fatalf("No teams fetched from source. Check the values of org, repos, tokens to ensure they are correct. Check logs for more details.")
		return
	}

	teamsSpinnerSuccess.Success()

	// Create teams in target organization
	createTeamsSpinnerSuccess, _ := pterm.DefaultSpinner.Start("Creating teams in target organization...")
	for _, team := range teams {
		// Map members
		if os.Getenv("GHMT_MAPPING_FILE") != "" {
			team = mapMembers(team)
		}

		//update Spinner text with the team name
		log.Println("Creating team in target organization: " + team.Name)

		team.CreateTeam()

	}
	createTeamsSpinnerSuccess.UpdateText("Team creation process completed")
	createTeamsSpinnerSuccess.Success()
}
