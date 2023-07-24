package api

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func newGHClient(token string) *githubv4.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)

	return githubv4.NewClient(httpClient)
}

func GetSourceOrganizationTeams() []string {
	client := newGHClient(viper.GetString("source_token"))

	var query struct {
		Organization struct {
			Teams struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []struct {
					Node struct {
						Slug string
					}
				}
			} `graphql:"teams(first: $first, after: $after)"`
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login": githubv4.String(viper.Get("SOURCE_ORGANIZATION").(string)),
		"first": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	teams := make([]string, 0)
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, team := range query.Organization.Teams.Edges {
			teams = append(teams, team.Node.Slug)
		}

		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)
	}

	return teams
}

func GetTeamMemberships(team string) [][]string {
	client := newGHClient(viper.GetString("source_token"))

	var query struct {
		Organization struct {
			Team struct {
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
				} `graphql:"members(first: $first, after: $after)"`
			} `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login": githubv4.String(viper.Get("SOURCE_ORGANIZATION").(string)),
		"slug":  githubv4.String(team),
		"first": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	var members [][]string
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, member := range query.Organization.Team.Members.Edges {
			members = append(members, []string{member.Node.Login, member.Node.Email})
		}

		if !query.Organization.Team.Members.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Team.Members.PageInfo.EndCursor)
	}

	return members
}

func GetTeamRepositories(team string) [][]string {
	client := newGHClient(viper.GetString("source_token"))

	var query struct {
		Organization struct {
			Team struct {
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
				} `graphql:"repositories(first: $first, after: $after)"`
			} `graphql:"team(slug: $slug)"`
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login": githubv4.String(viper.Get("SOURCE_ORGANIZATION").(string)),
		"slug":  githubv4.String(team),
		"first": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	var repositories [][]string
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, repo := range query.Organization.Team.Repositories.Edges {
			repositories = append(repositories, []string{repo.Node.Name, repo.Permission})
		}

		if !query.Organization.Team.Repositories.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Team.Repositories.PageInfo.EndCursor)
	}

	return repositories
}
