package cmd

import (
	"fmt"
	"github.com/hitminer/hitminer-file-manager/login"
	"github.com/hitminer/hitminer-file-manager/server"
	"github.com/hitminer/hitminer-file-manager/server/s3gateway"
	"github.com/hitminer/hitminer-file-manager/util/multibar/cmdbar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir [remote_path]",
	Short: "Make directory to hitminer file systems",
	Long: `Make directory to hitminer file systems.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("not login")
		}
		usr := viper.GetString("username")
		pw := viper.GetString("password")
		host := viper.GetString("host")
		if usr == "" || pw == "" {
			return fmt.Errorf("not login")
		}

		token, err := login.Login(host, usr, pw)
		if err != nil {
			return err
		}

		var svr server.S3Server = s3gateway.NewS3Server(cmd.Context(), host, token, cmdbar.NewBar(cmd.OutOrStdout()))
		err = svr.MakeDirectory(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkdirCmd)
}
