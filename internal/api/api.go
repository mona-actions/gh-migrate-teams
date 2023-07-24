package api

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/github"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func newGHGraphqlClient(token string) *githubv4.Client {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)

	return githubv4.NewClient(httpClient)
}

func newGHRestClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

	if err != nil {
		panic(err)
	}

	return github.NewClient(rateLimiter)
}

func GetSourceOrganizationTeams() []map[string]string {
	client := newGHGraphqlClient(viper.GetString("SOURCE_TOKEN"))

	var query struct {
		Organization struct {
			Teams struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []struct {
					Node struct {
						Id          string
						Name        string
						Description string
						Slug        string
						Privacy     string
						ParentTeam  struct {
							Id string
						}
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

	var teams = []map[string]string{}
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, team := range query.Organization.Teams.Edges {
			teams = append(teams, map[string]string{"Id": team.Node.Id, "Name": team.Node.Name, "Slug": team.Node.Slug, "Description": team.Node.Description, "Privacy": team.Node.Privacy, "ParentTeamId": team.Node.ParentTeam.Id})
		}

		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)
	}

	return teams
}

func GetTeamMemberships(team string) []map[string]string {
	client := newGHGraphqlClient(viper.GetString("source_token"))

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

	var members = []map[string]string{}
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, member := range query.Organization.Team.Members.Edges {
			members = append(members, map[string]string{"Login": member.Node.Login, "Email": member.Node.Email})
		}

		if !query.Organization.Team.Members.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Team.Members.PageInfo.EndCursor)
	}

	return members
}

func GetTeamRepositories(team string) []map[string]string {
	client := newGHGraphqlClient(viper.GetString("source_token"))

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

	var repositories = []map[string]string{}
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, repo := range query.Organization.Team.Repositories.Edges {
			repositories = append(repositories, map[string]string{"Name": repo.Node.Name, "Permission": repo.Permission})
		}

		if !query.Organization.Team.Repositories.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Team.Repositories.PageInfo.EndCursor)
	}

	return repositories
}

func CreateTeam(name string, description string, privacy string, parentTeamId string) {
	client := newGHRestClient(viper.GetString("TARGET_TOKEN"))

	r := github.NewTeam{Name: name, Description: &description, Privacy: &privacy}
	_, _, err := client.Teams.CreateTeam(context.Background(), viper.Get("TARGET_ORGANIZATION").(string), r)

	if err != nil {
		if strings.Contains(err.Error(), "Name must be unique for this org") {
			fmt.Println("Error creating team, team already exists: ", name)
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

// func AddTeamRepository(teamId int64, repo string, permission string) {
// 	client := newGHRestClient(viper.GetString("TARGET_TOKEN"))

// 	_, err := client.Teams.AddTeamRepo(context.Background(), teamId, viper.Get("TARGET_ORGANIZATION").(string), repo, &github.TeamAddTeamRepoOptions{Permission: permission})

// 	if err != nil {
// 		panic(err)
// 	}
// }
