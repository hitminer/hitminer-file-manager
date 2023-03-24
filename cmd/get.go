package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"hitminer-file-manager/login"
	"hitminer-file-manager/server"
	"hitminer-file-manager/server/s3gateway"
)

var getCmd = &cobra.Command{
	Use:   "get [remote_path] [local_path] ",
	Short: "Get files to hitminer file systems recursively",
	Long: `Get files to hitminer file systems recursively.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var remote string
		loacl := "."
		if len(args) == 1 {
			remote = args[0]
		} else if len(args) == 2 {
			remote, loacl = args[0], args[1]
		} else {
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
		err = svr.GetObjects(cmd.Context(), loacl, remote)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
