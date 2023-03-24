package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"hitminer-file-manager/login"
	"os"
	"path/filepath"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login hitminer",
	Long:  "Login hitminer with your username and password",
	RunE: func(cmd *cobra.Command, args []string) error {
		usr := viper.GetString("username")
		pw := viper.GetString("password")
		host := viper.GetString("host")
		_, err := login.Login(host, usr, pw)
		if err != nil {
			return err
		}

		home, _ := os.UserHomeDir()
		path := filepath.Join(home, ".config", "hitminer")
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}

		err = viper.WriteConfigAs(filepath.Join(path, "file_manager.toml"))
		if err != nil {
			return err
		}

		cmd.Printf("login successful\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "user name")
	loginCmd.Flags().StringP("password", "p", "", "password")
	loginCmd.Flags().StringP("host", "", "www.hitminer.cn", "colony host")
	_ = loginCmd.MarkFlagRequired("username")
	_ = loginCmd.MarkFlagRequired("password")
	_ = viper.BindPFlag("username", loginCmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("password", loginCmd.Flags().Lookup("password"))
	_ = viper.BindPFlag("host", loginCmd.Flags().Lookup("host"))
}
