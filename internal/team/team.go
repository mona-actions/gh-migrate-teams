package team

import (
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

			repositories = append(repositories, Repository{
				Name:       repository["Name"],
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
			for _, member := range t.Members {
				api.AddTeamMember(t.Slug, member.Login, member.Role)
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
