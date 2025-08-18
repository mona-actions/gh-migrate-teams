package sync

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mona-actions/gh-migrate-teams/internal/team"
)

func TestFilterTeamRepositories(t *testing.T) {
	tests := []struct {
		name          string
		team          team.Team
		repoList      []string
		expectedRepos []team.Repository
	}{
		{
			name: "Filter repositories - match found",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "repo1", Permission: "admin"},
					{Name: "repo2", Permission: "write"},
					{Name: "repo3", Permission: "read"},
				},
			},
			repoList: []string{"owner/repo1", "owner/repo3"},
			expectedRepos: []team.Repository{
				{Name: "repo1", Permission: "admin"},
				{Name: "repo3", Permission: "read"},
			},
		},
		{
			name: "Filter repositories - no matches",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "repo1", Permission: "admin"},
					{Name: "repo2", Permission: "write"},
				},
			},
			repoList:      []string{"owner/repo3", "owner/repo4"},
			expectedRepos: []team.Repository{},
		},
		{
			name: "Filter repositories - all match",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "repo1", Permission: "admin"},
					{Name: "repo2", Permission: "write"},
				},
			},
			repoList: []string{"owner/repo1", "owner/repo2"},
			expectedRepos: []team.Repository{
				{Name: "repo1", Permission: "admin"},
				{Name: "repo2", Permission: "write"},
			},
		},
		{
			name: "Empty team repositories",
			team: team.Team{
				Name:         "test-team",
				Repositories: []team.Repository{},
			},
			repoList:      []string{"owner/repo1", "owner/repo2"},
			expectedRepos: []team.Repository{},
		},
		{
			name: "Empty repo list",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "repo1", Permission: "admin"},
					{Name: "repo2", Permission: "write"},
				},
			},
			repoList:      []string{},
			expectedRepos: []team.Repository{},
		},
		{
			name: "Invalid repo format in list",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "repo1", Permission: "admin"},
					{Name: "repo2", Permission: "write"},
				},
			},
			repoList: []string{"repo1", "owner/repo2", "invalid/format/extra"},
			expectedRepos: []team.Repository{
				{Name: "repo2", Permission: "write"},
			},
		},
		{
			name: "Case sensitivity test",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "Repo1", Permission: "admin"},
					{Name: "repo2", Permission: "write"},
				},
			},
			repoList: []string{"owner/repo1", "owner/Repo1"},
			expectedRepos: []team.Repository{
				{Name: "Repo1", Permission: "admin"},
			},
		},
		{
			name: "Multiple owners same repo name",
			team: team.Team{
				Name: "test-team",
				Repositories: []team.Repository{
					{Name: "common-repo", Permission: "admin"},
					{Name: "unique-repo", Permission: "write"},
				},
			},
			repoList: []string{"owner1/common-repo", "owner2/different-repo"},
			expectedRepos: []team.Repository{
				{Name: "common-repo", Permission: "admin"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the team to avoid modifying the original test data
			teamCopy := team.Team{
				Id:             tt.team.Id,
				DatabaseId:     tt.team.DatabaseId,
				Name:           tt.team.Name,
				Slug:           tt.team.Slug,
				Description:    tt.team.Description,
				Privacy:        tt.team.Privacy,
				ParentTeamId:   tt.team.ParentTeamId,
				Members:        tt.team.Members,
				Repositories:   make([]team.Repository, len(tt.team.Repositories)),
				ParentTeamName: tt.team.ParentTeamName,
			}
			copy(teamCopy.Repositories, tt.team.Repositories)

			result := filterTeamRepositories(teamCopy, tt.repoList)

			// Check if the filtered repositories match expected
			if !reflect.DeepEqual(result.Repositories, tt.expectedRepos) {
				t.Errorf("filterTeamRepositories() = %v, expected %v", result.Repositories, tt.expectedRepos)
			}

			// Ensure other team fields are preserved
			if result.Name != tt.team.Name {
				t.Errorf("Team name was modified: got %v, expected %v", result.Name, tt.team.Name)
			}
		})
	}
}

func TestFilterTeamRepositories_PreservesTeamFields(t *testing.T) {
	originalTeam := team.Team{
		Id:           "123",
		DatabaseId:   456,
		Name:         "test-team",
		Slug:         "test-slug",
		Description:  "Test description",
		Privacy:      "closed",
		ParentTeamId: "parent-123",
		Members: []team.Member{
			{Login: "user1", Email: "user1@example.com", Role: "member"},
		},
		Repositories: []team.Repository{
			{Name: "repo1", Permission: "admin"},
		},
		ParentTeamName: "parent-team",
	}

	repoList := []string{"owner/repo1"}
	result := filterTeamRepositories(originalTeam, repoList)

	// Check that all non-repository fields are preserved
	if result.Id != originalTeam.Id {
		t.Errorf("Id was modified: got %v, expected %v", result.Id, originalTeam.Id)
	}
	if result.DatabaseId != originalTeam.DatabaseId {
		t.Errorf("DatabaseId was modified: got %v, expected %v", result.DatabaseId, originalTeam.DatabaseId)
	}
	if result.Name != originalTeam.Name {
		t.Errorf("Name was modified: got %v, expected %v", result.Name, originalTeam.Name)
	}
	if result.Slug != originalTeam.Slug {
		t.Errorf("Slug was modified: got %v, expected %v", result.Slug, originalTeam.Slug)
	}
	if result.Description != originalTeam.Description {
		t.Errorf("Description was modified: got %v, expected %v", result.Description, originalTeam.Description)
	}
	if result.Privacy != originalTeam.Privacy {
		t.Errorf("Privacy was modified: got %v, expected %v", result.Privacy, originalTeam.Privacy)
	}
	if result.ParentTeamId != originalTeam.ParentTeamId {
		t.Errorf("ParentTeamId was modified: got %v, expected %v", result.ParentTeamId, originalTeam.ParentTeamId)
	}
	if result.ParentTeamName != originalTeam.ParentTeamName {
		t.Errorf("ParentTeamName was modified: got %v, expected %v", result.ParentTeamName, originalTeam.ParentTeamName)
	}
	if !reflect.DeepEqual(result.Members, originalTeam.Members) {
		t.Errorf("Members were modified: got %v, expected %v", result.Members, originalTeam.Members)
	}
}

func BenchmarkFilterTeamRepositories(b *testing.B) {
	// Create a team with many repositories
	repositories := make([]team.Repository, 100)
	for i := 0; i < 100; i++ {
		repositories[i] = team.Repository{
			Name:       fmt.Sprintf("repo%d", i),
			Permission: "read",
		}
	}

	testTeam := team.Team{
		Name:         "benchmark-team",
		Repositories: repositories,
	}

	// Create a repo list with half the repositories
	repoList := make([]string, 50)
	for i := 0; i < 50; i++ {
		repoList[i] = fmt.Sprintf("owner/repo%d", i*2) // Every other repo
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterTeamRepositories(testTeam, repoList)
	}
}
