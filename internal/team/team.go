package team

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mona-actions/gh-migrate-teams/internal/api"
	"github.com/spf13/viper"
)

type Teams []Team

type Team struct {
	Id             string
	DatabaseId     int64
	Name           string
	Slug           string
	Description    string
	Privacy        string
	ParentTeamId   string
	Members        []Member
	Repositories   []Repository
	ParentTeamName string
}

type Member struct {
	Login string
	Email string
	Role  string
}

type Repository struct {
	Name       string
	Permission string
}

func GetSourceOrganizationTeams() Teams {
	data := api.GetSourceOrganizationTeams()

	teams := make([]Team, 0)
	for _, team := range data {
		// Fixing privacy values
		privacy := "SECRET"
		if team["Privacy"] != "SECRET" {
			privacy = "closed"
		}

		teams = append(teams, Team{
			Id:             team["Id"],
			Name:           team["Name"],
			Slug:           team["Slug"],
			Description:    team["Description"],
			Privacy:        privacy,
			ParentTeamId:   team["ParentTeamId"],
			ParentTeamName: team["ParentTeamName"],
			Members:        getTeamMemberships(team["Slug"]),
			Repositories:   getTeamRepositories(team["Slug"]),
		})
	}

	return teams
}

func getTeamMemberships(team string) []Member {
	data := api.GetTeamMemberships(team)

	members := make([]Member, 0)
	for _, member := range data {
		members = append(members, Member{
			Login: member["Login"],
			Email: member["Email"],
			Role:  member["Role"],
		})
	}

	return members
}

func getTeamRepositories(team string) []Repository {

	//check if the repo-mappings.csv file exists
	filePath := viper.GetString("REPO_MAPPING_FILE")
	var repoMappings map[string]string
	if _, err := os.Stat(filePath); err == nil {
		// Read repo mappings
		repoMappings, err = readMappings(filePath)
		if err != nil {
			log.Println("Unable to read repo mappings - ", err)
		}
	}

	data := api.GetTeamRepositories(team)

	repositories := make([]Repository, 0)
	for _, repository := range data {
		if repository["Name"] != "" {
			// Fixing permission values
			permission := "pull"
			if repository["Permission"] == "WRITE" {
				permission = "push"
			} else if repository["Permission"] == "ADMIN" {
				permission = "admin"
			}

			repoName := repository["Name"]
			sourceOrg := viper.GetString("SOURCE_ORGANIZATION")
			repoWithOwner := sourceOrg + "/" + repoName
			// Check if the repository name exists in the mappings
			if newName, exists := repoMappings[repoWithOwner]; exists {
				repoName = newName
			}

			repositories = append(repositories, Repository{
				Name:       repoName,
				Permission: permission,
			})
		}
	}

	return repositories
}

func (t Team) CreateTeam() {
	// We Send ParentTeamName as that is easiest to get the ParentTeamId
	err := api.CreateTeam(t.Name, t.Description, t.Privacy, t.ParentTeamName)

	skipTeams := viper.GetBool("SKIP_TEAMS")

	// Adding a wait to account for race condition
	time.Sleep(3 * time.Second)

	//skip adding repositories and members if team already exists to save on API calls
	if err != nil && skipTeams {
		if strings.Contains(err.Error(), "Name must be unique for this org") {
			return
		}
	} else {
		for _, repository := range t.Repositories {
			api.AddTeamRepository(t.Slug, repository.Name, repository.Permission)
		}

		// Check to see if user sync has been disabled
		userSync := viper.GetString("USER_SYNC")

		if userSync != "disable" {
			authenticatedUser, err := api.GetAuthenticatedUser()
			if err != nil {
				log.Println("Unable to get authenticated user - ", err)
			}

			authUserLogin := authenticatedUser.GetLogin()

			memberMap := make(map[string]bool)
			for _, member := range t.Members {
				memberMap[member.Login] = true
				api.AddTeamMember(t.Slug, member.Login, member.Role)
			}

			//If authenticated user is not part of the members, remove them from the team
			if authUserLogin != "" && !memberMap[authUserLogin] {
				err := api.RemoveTeamMember(t.Slug, authenticatedUser.GetLogin())
				if err != nil {
					log.Println("Unable to remove authenticated user from team - ", err)
				} else {
					log.Println(authenticatedUser.GetLogin(), "removed from team as they are not part of the members list")
				}
			}
		}
	}
}

func (t Teams) ExportTeamMemberships() [][]string {
	memberships := make([][]string, 0)
	for _, team := range t {
		for _, member := range team.Members {
			memberships = append(memberships, []string{team.Name, member.Login, member.Email})
		}
	}

	return memberships
}

func (t Teams) ExportTeamRepositories() [][]string {
	repositories := make([][]string, 0)
	for _, team := range t {
		for _, repository := range team.Repositories {
			repositories = append(repositories, []string{team.Name, repository.Name, repository.Permission})
		}
	}

	return repositories
}

func GetRepositoryTeams(repository string) Teams {
	//split the repository string to get the owner and repo name
	repo := strings.Split(repository, "/")
	owner := repo[0]
	repoName := repo[1]
	viper.Set("SOURCE_ORGANIZATION", owner)
	data, err := api.GetRepositoryTeams(owner, repoName)

	if err != nil {
		log.Println("Unable to get repository teams - ", err)
	}
	parentTeamID := ""
	parentTeamName := ""

	// Check if the team-mappings.csv file exists
	filePath := viper.GetString("TEAM_MAPPING_FILE")
	var teamMappings map[string]string
	if _, err := os.Stat(filePath); err == nil {
		// Read team mappings
		teamMappings, err = readMappings(filePath)
		if err != nil {
			log.Println("Unable to read team mappings - ", err)
		}
	}

	teams := make(Teams, 0, len(data))
	for _, team := range data {
		if team.Parent != nil {
			parentTeamID = strconv.FormatInt(team.Parent.GetID(), 10)
			parentTeamName = *team.Parent.Name
		}

		teamName := team.GetName()
		teamSlug := team.GetSlug()
		// Check if the team name exists in the mappings
		if newName, exists := teamMappings[owner+"/"+teamName]; exists {
			teamName = newName
			teamSlug = newName
		}

		team := Team{
			Id:             strconv.FormatInt(team.GetID(), 10),
			Name:           teamName,
			Slug:           teamSlug,
			Description:    team.GetDescription(),
			Privacy:        team.GetPrivacy(),
			ParentTeamId:   parentTeamID,
			ParentTeamName: parentTeamName,
			Members:        getTeamMemberships(*team.Slug),
			Repositories:   getTeamRepositories(*team.Slug),
		}
		teams = append(teams, team)
	}

	return teams
}

func readMappings(filePath string) (map[string]string, error) {
	mappings := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records[1:] { // Skip header
		mappings[record[0]] = record[1]
	}

	return mappings, nil
}
