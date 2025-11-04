package cmd

import (
	"fmt"

	"github.com/hookie/cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored authentication",
	Long:  `Clear the stored authentication token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}
		fmt.Println("✓ Logged out successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

