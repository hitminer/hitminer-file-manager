package cmd

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hitminer/hitminer-file-manager/login"
	"github.com/hitminer/hitminer-file-manager/server"
	"github.com/hitminer/hitminer-file-manager/server/s3gateway"
	"github.com/hitminer/hitminer-file-manager/util/multibar/cmdbar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

var lsCmd = &cobra.Command{
	Use:   "ls [remote_path]",
	Short: "List files in hitminer file systems",
	Long: `List files in hitminer file systems.
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
		loc := time.Now().Location()
		for object := range svr.ListObjects(cmd.Context(), args[0], "/") {
			var buffer bytes.Buffer
			if object.IsDirectory {
				buffer.WriteString("drwxr-xr-x\t")
				buffer.WriteString(fmt.Sprintf("%9s\t", humanize.IBytes(uint64(object.Size))))
				buffer.WriteString(object.LastModifiedTime.In(loc).Format("Jan _2 15:04"))
				buffer.WriteString("\t")
				buffer.WriteString(object.Name)
				buffer.WriteString("\n")
			} else {
				buffer.WriteString("-rwxr-xr--\t")
				buffer.WriteString(fmt.Sprintf("%9s\t", humanize.IBytes(uint64(object.Size))))
				buffer.WriteString(object.LastModifiedTime.In(loc).Format("Jan 02 15:04"))
				buffer.WriteString("\t")
				buffer.WriteString(object.Name)
				buffer.WriteString("\n")
			}
			_, _ = cmd.OutOrStdout().Write(buffer.Bytes())
		}
		err = svr.GetError()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
