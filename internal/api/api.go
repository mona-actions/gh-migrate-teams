package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v62/github"
	"github.com/jferrl/go-githubauth"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func newHTTPClient() *http.Client {

	token := viper.GetString("TARGET_TOKEN")
	appId := viper.GetString("TARGET_APP_ID")
	privateKey := []byte(viper.GetString("TARGET_PRIVATE_KEY"))
	installationId := viper.GetInt64("TARGET_INSTALLATION_ID")

	// check that Target token or GitHub App values are set
	if token == "" && (appId == "" || len(privateKey) == 0 || installationId == 0) {
		log.Fatalf(
			"Please provide a target token or a target GitHub App ID and private key.")
	}

	if appId != "" && len(privateKey) != 0 && installationId != 0 {
		// GitHub App authentication

		appIdInt, err := strconv.ParseInt(appId, 10, 64)
		if err != nil {
			log.Fatalf("Error converting app ID to int64: %v", err)
		}
		appToken, err := githubauth.NewApplicationTokenSource(appIdInt, privateKey)
		if err != nil {
			log.Fatalf("Error creating app token: %v", err)
		}

		installationToken := githubauth.NewInstallationTokenSource(installationId, appToken)

		// Create HTTP client with automatic token refresh
		httpClient := oauth2.NewClient(context.Background(), installationToken)

		return httpClient

	}
	// Personal access token authentication
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(context.Background(), src)

}

type RateLimitAwareGraphQLClient struct {
	client *githubv4.Client
}

func (c *RateLimitAwareGraphQLClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	var rateLimitQuery struct {
		RateLimit struct {
			Remaining int
			ResetAt   githubv4.DateTime
		}
	}

	for {
		// Check the current rate limit
		if err := c.client.Query(ctx, &rateLimitQuery, nil); err != nil {
			return err
		}

		log.Println("Rate limit remaining:", rateLimitQuery.RateLimit.Remaining)

		if rateLimitQuery.RateLimit.Remaining > 0 {
			// Proceed with the actual query
			err := c.client.Query(ctx, q, variables)
			if err != nil {
				return err
			}
			return nil
		} else {
			// Sleep until rate limit resets
			log.Println("Rate limit exceeded, sleeping until reset at:", rateLimitQuery.RateLimit.ResetAt.Time)
			time.Sleep(time.Until(rateLimitQuery.RateLimit.ResetAt.Time))

		}
	}
}

func newGHGraphqlClient(token string) *RateLimitAwareGraphQLClient {
	hostname := viper.GetString("SOURCE_HOSTNAME")
	var baseClient *githubv4.Client

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(httpClient.Transport)

	if err != nil {
		panic(err)
	}

	// Trim any trailing slashes from the hostname
	hostname = strings.TrimSuffix(hostname, "/")

	// If hostname is received, create a new client with the hostname
	if hostname != "" {
		hostname = strings.TrimSuffix(hostname, "/")
		if !strings.HasPrefix(hostname, "https://") {
			hostname = "https://" + hostname
		}
		baseClient = githubv4.NewEnterpriseClient(hostname+"/api/graphql", rateLimiter)
	} else {
		baseClient = githubv4.NewClient(rateLimiter)
	}

	return &RateLimitAwareGraphQLClient{
		client: baseClient,
	}
}

func newGHRestClient() *github.Client {
	httpClient := newHTTPClient()
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(httpClient.Transport)

	if err != nil {
		panic(err)
	}

	return github.NewClient(rateLimiter)
}

func newSourceGHRestClient() *github.Client {
	token := viper.GetString("SOURCE_TOKEN")
	hostname := viper.GetString("SOURCE_HOSTNAME")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)

	if err != nil {
		panic(err)
	}

	if hostname != "" {
		hostname = strings.TrimSuffix(hostname, "/")
		if !strings.HasPrefix(hostname, "https://") {
			hostname = "https://" + hostname
		}
		baseURL := fmt.Sprintf("%s/api/v3/", hostname)
		client, err := github.NewClient(rateLimiter).WithEnterpriseURLs(baseURL, baseURL)
		if err != nil {
			panic(err)
		}
		return client
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
						Role string
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
			members = append(members, map[string]string{"Login": member.Node.Login, "Email": member.Node.Email, "Role": member.Role})
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

func CreateTeam(name string, description string, privacy string, parentTeamName string) error {
	client := newGHRestClient()

	t := github.NewTeam{Name: name, Description: &description, Privacy: &privacy}
	if parentTeamName != "" {
		parentTeamID, err := GetTeamId(parentTeamName)
		if err != nil {
			fmt.Println("Team ID Not found", err)
		} else {
			t.ParentTeamID = &parentTeamID
		}
	}

	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
	_, _, err := client.Teams.CreateTeam(ctx, viper.Get("TARGET_ORGANIZATION").(string), t)

	if err != nil {
		if strings.Contains(err.Error(), "Name must be unique for this org") {
			fmt.Println("Team: ", name, "already exists in destination skipping...")
			return err
		} else {
			fmt.Println("Unable to create team:", name, err.Error())
			return err
		}
	}
	return nil
}

func AddTeamRepository(slug string, repo string, permission string) {
	client := newGHRestClient()

	fmt.Println("Adding repository to team: ", slug, repo, permission)

	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
	_, err := client.Teams.AddTeamRepoBySlug(ctx, viper.Get("TARGET_ORGANIZATION").(string), slug, viper.Get("TARGET_ORGANIZATION").(string), repo, &github.TeamAddTeamRepoOptions{Permission: permission})

	if err != nil {
		if strings.Contains(err.Error(), "422 Validation Failed") {
			fmt.Println("Error adding repository to team: ", slug, repo, permission)
		} else if strings.Contains(err.Error(), "404 Not Found") {
			fmt.Println("Error adding repository to team, repository not found: ", slug, repo, permission)
		} else {
			fmt.Println("error adding repository", repo, " to team: ", slug, "with permissions:", permission, "Unknown error", err, err.Error())
		}
	}
}

func AddTeamMember(slug string, member string, role string) {
	client := newGHRestClient()

	role = strings.ToLower(role) // lowercase to match github api
	fmt.Println("Adding member to team: ", slug, member, role)

	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
	_, _, err := client.Teams.AddTeamMembershipBySlug(ctx, viper.Get("TARGET_ORGANIZATION").(string), slug, member, &github.TeamAddTeamMembershipOptions{Role: role})
	if err != nil {
		fmt.Println("Error adding member ", member, " to team: ", slug, err)
	}
}

func GetTeamId(TeamName string) (int64, error) {
	client := newGHRestClient()

	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
	team, _, err := client.Teams.GetTeamBySlug(ctx, viper.Get("TARGET_ORGANIZATION").(string), TeamName)
	if err != nil {
		fmt.Println("Error getting parent team ID: ", TeamName)
		return 0, err
	}
	return *team.ID, nil
}

func GetRepositoryTeams(owner string, repo string) ([]*github.Team, error) {
	client := newSourceGHRestClient()
	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)

	// Get teams for the repository
	teams, _, err := client.Repositories.ListTeams(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func GetAuthenticatedUser() (*github.User, error) {
	client := newGHRestClient()
	ctx := context.Background()

	// Get authenticated user
	user, _, err := client.Users.Get(ctx, "")

	if err != nil {
		if strings.Contains(err.Error(), "403 Resource not accessible by integration") {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func RemoveTeamMember(slug string, member string) error {
	client := newGHRestClient()
	ctx := context.WithValue(context.Background(), github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)

	_, err := client.Teams.RemoveTeamMembershipBySlug(ctx, viper.Get("TARGET_ORGANIZATION").(string), slug, member)
	if err != nil {
		return err
	}
	return nil
}
