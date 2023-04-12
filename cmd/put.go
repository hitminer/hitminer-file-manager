package cmd

import (
	"fmt"
	"github.com/hitminer/hitminer-file-manager/login"
	"github.com/hitminer/hitminer-file-manager/server"
	"github.com/hitminer/hitminer-file-manager/server/s3gateway"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var putCmd = &cobra.Command{
	Use:   "put [local_path] [remote_path]",
	Short: "Put files to hitminer file systems recursively",
	Long: `Put files to hitminer file systems recursively.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
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

		var svr server.S3Server = s3gateway.NewS3Server(cmd.Context(), host, token)
		err = svr.PutObjects(cmd.Context(), args[0], args[1])
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
