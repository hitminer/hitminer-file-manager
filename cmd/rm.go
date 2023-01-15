package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"hitminer-file-manager/login"
	"hitminer-file-manager/server"
	"path/filepath"
)

var rmCmd = &cobra.Command{
	Use:   "rm [remote_path]",
	Short: "Remove files to hitminer file systems",
	Long: `Remove files to hitminer file systems.
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
		recursive := viper.GetBool("recursive")
		if usr == "" || pw == "" {
			return fmt.Errorf("not login")
		}
		token, err := login.Login(usr, pw)
		if err != nil {
			return err
		}
		info, err := login.Verify(token)
		if err != nil {
			return err
		}
		svr := server.NewServer(cmd.Context(), info.Endpoint, info.AccessKey, info.SecretKey, filepath.ToSlash(filepath.Join("workplace", info.Uid))+"/")
		if recursive {
			svr.RemoveObjects(cmd.Context(), filepath.ToSlash(filepath.Join("workplace", info.Uid, args[0])))
		} else {
			svr.RemoveObject(cmd.Context(), filepath.ToSlash(filepath.Join("workplace", info.Uid, args[0])))
		}
		svr.Wait()
		svr.Finish()
		if svr.Err != nil {
			return svr.Err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolP("recursive", "r", false, "remove files recursively")
	_ = viper.BindPFlag("recursive", rmCmd.Flags().Lookup("recursive"))
}
