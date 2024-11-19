/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/mona-actions/gh-migrate-teams/pkg/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// byReposCmd represents the byRepos command
var byReposCmd = &cobra.Command{
	Use:   "byRepos",
	Short: "Migrates teams by repository",
	Long: `Migrates team based on a repository list. 
	
	It will migrate all the teams that have access to the repositories in the list.`,
	Run: func(cmd *cobra.Command, args []string) {

		targetOrganization := cmd.Flag("target-organization").Value.String()
		sourceToken := cmd.Flag("source-token").Value.String()
		targetToken := cmd.Flag("target-token").Value.String()
		mappingFile := cmd.Flag("mapping-file").Value.String()
		ghHostname := cmd.Flag("source-hostname").Value.String()
		repoFile := cmd.Flag("from-file").Value.String()
		skipTeams := cmd.Flag("skip-teams").Value.String()
		tAppId := cmd.Flag("target-app-id").Value.String()
		tInstallationId := cmd.Flag("target-installation-id").Value.String()

		// Set ENV variables
		os.Setenv("GHMT_TARGET_ORGANIZATION", targetOrganization)
		os.Setenv("GHMT_SOURCE_TOKEN", sourceToken)
		os.Setenv("GHMT_TARGET_TOKEN", targetToken)
		os.Setenv("GHMT_MAPPING_FILE", mappingFile)
		os.Setenv("GHMT_SOURCE_HOSTNAME", ghHostname)
		os.Setenv("GHMT_REPO_FILE", repoFile)
		os.Setenv("GHMT_SKIP_TEAMS", skipTeams)
		os.Setenv("GHMT_TARGET_APP_ID", tAppId)
		os.Setenv("GHMT_TARGET_INSTALLATION_ID", tInstallationId)

		// Bind ENV variables in Viper
		viper.BindEnv("TARGET_ORGANIZATION")
		viper.BindEnv("SOURCE_TOKEN")
		viper.BindEnv("TARGET_TOKEN")
		viper.BindEnv("MAPPING_FILE")
		viper.BindEnv("SOURCE_HOSTNAME")
		viper.BindEnv("USER_SYNC")
		viper.BindEnv("SKIP_TEAMS")
		viper.BindEnv("REPO_FILE")
		viper.BindEnv("TARGET_PRIVATE_KEY")
		viper.BindEnv("TARGET_APP_ID")
		viper.BindEnv("TARGET_INSTALLATION_ID")

		sync.SyncTeamsByRepo()
	},
}

func init() {

	syncCmd.AddCommand(byReposCmd)

	// Here you will define your flags and configuration settings.

	byReposCmd.Flags().StringP("target-organization", "t", "", "Target Organization to sync teams from")
	byReposCmd.MarkFlagRequired("target-organization")

	byReposCmd.Flags().StringP("source-token", "a", "", "Source Organization GitHub token. Scopes: read:org, read:user, user:email")
	byReposCmd.MarkFlagRequired("source-token")

	byReposCmd.Flags().StringP("target-token", "b", "", "Target Organization GitHub token. Scopes: admin:org")

	byReposCmd.Flags().StringP("from-file", "f", "repositories.txt", "File path to use for repository list")
	byReposCmd.MarkFlagRequired("from-file")

	byReposCmd.Flags().StringP("mapping-file", "m", "", "Mapping file path to use for mapping teams members handles")

	byReposCmd.Flags().BoolP("skip-teams", "k", false, "Skips adding members and repos to teams that already exist to save on API requests (default \"false\")")

	byReposCmd.Flags().StringP("source-hostname", "u", os.Getenv("SOURCE_HOST"), "GitHub Enterprise source hostname url (optional) Ex. https://github.example.com")

	byReposCmd.Flags().StringP("target-private-key", "p", "", "Private key for GitHub App authentication. Ideally set as an env variable `GHMT_TARGET_PRIVATE_KEY`")
	viper.BindPFlag("TARGET_PRIVATE_KEY", byReposCmd.Flags().Lookup("target-private-key"))

	byReposCmd.Flags().StringP("target-app-id", "i", "", "GitHub App ID")

	byReposCmd.Flags().Int64P("target-installation-id", "l", 0, "GitHub App Installation ID")
}
