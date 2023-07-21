package api

import (
	"context"
	"fmt"

	"github.com/mona-actions/gh-migrate-teams/internal/team"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var query struct {
	Organization struct {
		Teams struct {
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
			Edges []struct {
				Node struct {
					Name    string
					Members struct {
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
						Edges []struct {
							Node struct {
								Login string
								Email string
							}
						}
					} `graphql:"members(first: $first, after: $members_after)"`
					Repositories struct {
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
						Edges []struct {
							Permission string
							Node       struct {
								Name string
							}
						}
					} `graphql:"repositories(first: $first, after: $repositories_after)"`
				}
			}
		} `graphql:"teams(first: $first, after: $teams_after)"`
	} `graphql:"organization(login: $login)"`
}

func newGHClient(token string) *githubv4.Client {
	fmt.Println("Using token: ", token)
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)

	return githubv4.NewClient(httpClient)
}

func GetSourceOrganizationTeams() []team.Team {
	client := newGHClient(viper.GetString("source_token"))

	variables := map[string]interface{}{
		"login":              githubv4.String(viper.Get("SOURCE_ORGANIZATION").(string)),
		"first":              githubv4.Int(100),
		"teams_after":        (*githubv4.String)(nil),
		"members_after":      (*githubv4.String)(nil),
		"repositories_after": (*githubv4.String)(nil),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		panic(err)
	}

	// Get all teams, team membership, and team repositories
	var teams []team.Team
	for {
		for _, orgTeam := range query.Organization.Teams.Edges {
			// Get all team members
			var members []team.Member
			for {
				for _, member := range orgTeam.Node.Members.Edges {
					members = append(members, team.Member{Login: member.Node.Login, Email: member.Node.Email})
				}
				if !orgTeam.Node.Members.PageInfo.HasNextPage {
					break
				}
				variables["members_after"] = githubv4.NewString(orgTeam.Node.Members.PageInfo.EndCursor)
			}
			// Get all team repositories
			var repositories []team.Repository
			for {
				for _, repository := range orgTeam.Node.Repositories.Edges {
					repositories = append(repositories, team.Repository{Name: repository.Node.Name, Permission: repository.Permission})
				}
				if !orgTeam.Node.Repositories.PageInfo.HasNextPage {
					break
				}
				variables["repositories_after"] = githubv4.NewString(orgTeam.Node.Repositories.PageInfo.EndCursor)
			}
			// Append to teams slice
			teams = append(teams, team.Team{Name: orgTeam.Node.Name, Members: members, Repositories: repositories})
		}

		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["teams_after"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)
	}

	return teams
}
