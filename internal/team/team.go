package team

import (
	"github.com/mona-actions/gh-migrate-teams/internal/api"
)

type Teams []Team

type Team struct {
	Id           string
	Name         string
	Slug         string
	Description  string
	Privacy      string
	ParentTeamId string
	Members      []Member
	Repositories []Repository
}

type Member struct {
	Login string
	Email string
}

type Repository struct {
	Name       string
	Permission string
}

func GetSourceOrganizationTeams() Teams {
	data := api.GetSourceOrganizationTeams()

	teams := make([]Team, 0)
	for _, team := range data {
		teams = append(teams, Team{
			Id:           team["Id"],
			Name:         team["Name"],
			Slug:         team["Slug"],
			Description:  team["Description"],
			Privacy:      team["Privacy"],
			ParentTeamId: team["ParentTeamId"],
			Members:      getTeamMemberships(team["Slug"]),
			Repositories: getTeamRepositories(team["Slug"]),
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
		})
	}

	return members
}

func getTeamRepositories(team string) []Repository {
	data := api.GetTeamRepositories(team)

	repositories := make([]Repository, 0)
	for _, repository := range data {
		repositories = append(repositories, Repository{
			Name:       repository["Name"],
			Permission: repository["Permission"],
		})
	}

	return repositories
}

func (t Team) CreateTeam() {
	api.CreateTeam(t.Name, t.Description, t.Privacy, t.ParentTeamId)
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
