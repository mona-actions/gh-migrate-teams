package repository

import "github.com/mona-actions/gh-migrate-teams/internal/api"

type repositories []Repository

type Repository struct {
	Name          string
	Collaborators []Collaborator
}

type Collaborator struct {
	Login      string
	Email      string
	Permission string
}

func GetSourceOrganizationRepositories() repositories {
	data := api.GetSourceOrganizationRepositories()

	repositories := make([]Repository, 0)
	for _, repository := range data {
		repositories = append(repositories, Repository{
			Name:          repository["Name"],
			Collaborators: getRepositoryCollaborators(repository["Name"]),
		})
	}

	return repositories
}

func getRepositoryCollaborators(repository string) []Collaborator {
	data := api.GetRepositoryCollaborators(repository)

	collaborators := make([]Collaborator, 0)
	for _, collaborator := range data {
		collaborators = append(collaborators, Collaborator{
			Login:      collaborator["Login"],
			Email:      collaborator["Email"],
			Permission: collaborator["Permission"],
		})
	}

	return collaborators
}

func (r repositories) ExportRepositoryCollaborators() [][]string {
	collaborators := make([][]string, 0)

	for _, repository := range r {
		for _, collaborator := range repository.Collaborators {
			collaborators = append(collaborators, []string{
				repository.Name,
				collaborator.Login,
				collaborator.Email,
				collaborator.Permission,
			})
		}
	}

	return collaborators
}
