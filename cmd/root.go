package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var rootCmd = &cobra.Command{
	Use:          "hitminer-file-manager",
	Short:        "hitminer-file-manager client",
	Long:         "hitminer-file-manager client, it can be used to manage hitminer file systems remotely",
	Version:      version,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_ = cmd.Usage()
		return err
	})

	viper.SetDefault("host", "www.hitminer.cn")
	viper.SetConfigName("file_manager")
	viper.SetConfigType("toml")
	home, _ := os.UserHomeDir()
	viper.AddConfigPath(filepath.Join(home, ".config", "hitminer"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_ = fmt.Errorf("Error:%s\n", err.Error())
		os.Exit(1)
	}
}
