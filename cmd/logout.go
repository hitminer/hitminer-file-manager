package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout hitminer",
	Long:  "Logout hitminer",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		err := os.RemoveAll(filepath.Join(home, ".config", "hitminer", "file_manager.toml"))
		if err != nil {
			return err
		}
		cmd.Printf("logout successful\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
