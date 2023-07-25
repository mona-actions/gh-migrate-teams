/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/mona-actions/gh-migrate-teams/pkg/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the export command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Recreates teams, membership, and team repo roles from a source organization to a target organization",
	Long:  "Recreates teams, membership, and team repo roles from a source organization to a target organization",
	Run: func(cmd *cobra.Command, args []string) {
		// Get parameters
		sourceOrganization := cmd.Flag("source-organization").Value.String()
		targetOrganization := cmd.Flag("target-organization").Value.String()
		sourceToken := cmd.Flag("source-token").Value.String()
		targetToken := cmd.Flag("target-token").Value.String()

		// Set ENV variables
		os.Setenv("GHMT_SOURCE_ORGANIZATION", sourceOrganization)
		os.Setenv("GHMT_TARGET_ORGANIZATION", targetOrganization)
		os.Setenv("GHMT_SOURCE_TOKEN", sourceToken)
		os.Setenv("GHMT_TARGET_TOKEN", targetToken)

		// Bind ENV variables in Viper
		viper.BindEnv("SOURCE_ORGANIZATION")
		viper.BindEnv("TARGET_ORGANIZATION")
		viper.BindEnv("SOURCE_TOKEN")
		viper.BindEnv("TARGET_TOKEN")

		// Call syncTeams
		sync.SyncTeams()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Flags
	syncCmd.Flags().StringP("source-organization", "s", "", "Source Organization to sync teams from")
	syncCmd.MarkFlagRequired("source-organization")

	syncCmd.Flags().StringP("target-organization", "t", "", "Target Organization to sync teams from")
	syncCmd.MarkFlagRequired("target-organization")

	syncCmd.Flags().StringP("source-token", "a", "", "Source Organization GitHub token")
	syncCmd.MarkFlagRequired("source-token")

	syncCmd.Flags().StringP("target-token", "b", "", "Target Organization GitHub token")
	syncCmd.MarkFlagRequired("target-token")

}
