//go:build (windows && amd64) || (linux && amd64) || (darwin && amd64) || (darwin && arm64)
// +build windows,amd64 linux,amd64 darwin,amd64 darwin,arm64

package cmd

import (
	"github.com/spf13/cobra"
	"hitminer-file-manager/upgrade"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade hitminer file manager",
	Long:  "Upgrade hitminer file manager to the latest versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := upgrade.Upgrade(cmd.Context())
		if err != nil {
			return err
		}
		cmd.Printf("upgrade successful\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
