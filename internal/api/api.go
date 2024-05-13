package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v53/github"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func newGHGraphqlClient(token string) *githubv4.Client {
	hostname := viper.GetString("SOURCE_HOSTNAME")
	var client *githubv4.Client
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(httpClient.Transport)

	if err != nil {
		panic(err)
	}
	client = githubv4.NewClient(rateLimiter)

	// Trim any trailing slashes from the hostname
	hostname = strings.TrimSuffix(hostname, "/")

	// If hostname is received, create a new client with the hostname
	if hostname != "" {
		client = githubv4.NewEnterpriseClient(hostname+"/api/graphql", rateLimiter)
	}
	return client
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
							Id   string
							Slug string
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
			teams = append(teams, map[string]string{
				"Id":             team.Node.Id,
				"Name":           team.Node.Name,
				"Slug":           team.Node.Slug,
				"Description":    team.Node.Description,
				"Privacy":        team.Node.Privacy,
				"ParentTeamId":   team.Node.ParentTeam.Id,
				"ParentTeamName": team.Node.ParentTeam.Slug,
			})
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

func GetSourceOrganizationRepositories() []map[string]string {
	client := newGHGraphqlClient(viper.GetString("source_token"))

	var query struct {
		Organization struct {
			Repositories struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []struct {
					Node struct {
						Name string
					}
				}
			} `graphql:"repositories(first: $first, after: $after)"`
		} `graphql:"organization(login: $login)"`
	}

	variables := map[string]interface{}{
		"login": githubv4.String(viper.Get("SOURCE_ORGANIZATION").(string)),
		"first": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	var repositories = []map[string]string{}
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, repo := range query.Organization.Repositories.Edges {
			repositories = append(repositories, map[string]string{"Name": repo.Node.Name})
		}

		if !query.Organization.Repositories.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Organization.Repositories.PageInfo.EndCursor)
	}

	return repositories
}

func GetRepositoryCollaborators(repository string) []map[string]string {
	client := newGHGraphqlClient(viper.GetString("source_token"))

	var query struct {
		Repository struct {
			Collaborators struct {
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
				Edges []struct {
					Permission string
					Node       struct {
						Login string
						Email string
					}
				}
			} `graphql:"collaborators(first: $first, after: $after)"`
		} `graphql:"repository(name: $name, owner: $owner)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(viper.Get("SOURCE_ORGANIZATION").(string)),
		"name":  githubv4.String(repository),
		"first": githubv4.Int(100),
		"after": (*githubv4.String)(nil),
	}

	var collaborators = []map[string]string{}
	for {
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			panic(err)
		}

		for _, collaborator := range query.Repository.Collaborators.Edges {
			collaborators = append(collaborators, map[string]string{"Login": collaborator.Node.Login, "Email": collaborator.Node.Email, "Permission": collaborator.Permission})
		}

		if !query.Repository.Collaborators.PageInfo.HasNextPage {
			break
		}

		variables["after"] = githubv4.NewString(query.Repository.Collaborators.PageInfo.EndCursor)
	}

	return collaborators
}

func CreateTeam(name string, description string, privacy string, parentTeamName string) {
	client := newGHRestClient(viper.GetString("TARGET_TOKEN"))

	t := github.NewTeam{Name: name, Description: &description, Privacy: &privacy}
	if parentTeamName != "" {
		parentTeamID, err := GetTeamId(parentTeamName)
		if err != nil {
			fmt.Println(err)
		}
		t.ParentTeamID = &parentTeamID
	}

	_, _, err := client.Teams.CreateTeam(context.Background(), viper.Get("TARGET_ORGANIZATION").(string), t)

	if err != nil {
		if strings.Contains(err.Error(), "Name must be unique for this org") {
			fmt.Println("Error creating team, team already exists: ", name)
		} else {
			fmt.Println("Unable to create team:", name, err.Error())
		}
	}
}

func AddTeamRepository(slug string, repo string, permission string) {
	client := newGHRestClient(viper.GetString("TARGET_TOKEN"))

	fmt.Println("Adding repository to team: ", slug, repo, permission)

	_, err := client.Teams.AddTeamRepoBySlug(context.Background(), viper.Get("TARGET_ORGANIZATION").(string), slug, viper.Get("TARGET_ORGANIZATION").(string), repo, &github.TeamAddTeamRepoOptions{Permission: permission})

	if err != nil {
		if strings.Contains(err.Error(), "422 Validation Failed") {
			fmt.Println("Error adding repository to team: ", slug, repo, permission)
		} else if strings.Contains(err.Error(), "404 Not Found") {
			fmt.Println("Error adding repository to team, repository not found: ", slug, repo, permission)
		} else {
			fmt.Println("adding repository to team: ", slug, repo, permission, "Unknown error", err, err.Error())
		}
	}
}

func AddTeamMember(slug string, member string) {
	client := newGHRestClient(viper.GetString("TARGET_TOKEN"))

	fmt.Println("Adding member to team: ", slug, member)
	_, _, err := client.Teams.AddTeamMembershipBySlug(context.Background(), viper.Get("TARGET_ORGANIZATION").(string), slug, member, &github.TeamAddTeamMembershipOptions{Role: "member"})
	if err != nil {
		fmt.Println("Error adding member to team: ", slug, member)
	}
}

func GetTeamId(TeamName string) (int64, error) {
	client := newGHRestClient(viper.GetString("TARGET_TOKEN"))
	team, _, err := client.Teams.GetTeamBySlug(context.Background(), viper.Get("TARGET_ORGANIZATION").(string), TeamName)
	if err != nil {
		fmt.Println("Error getting team ID: ", TeamName)
		return 0, err
	}
	return *team.ID, nil
}
