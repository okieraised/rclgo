package cmd

import (
	"fmt"
	"os"

	"github.com/okieraised/rclgo/internal/utilities"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "gen-rlc",
	Short: "ROS2 client library in Golang - ROS2 Message generator",
	Long:  `Call this program to generate Go types for the ROS2 messages found in your ROS2 environment.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	//cobra.OnInitialize(initConfig)
	//
	//rootCmd.PersistentFlags().StringP("config-file", "c", "", "config file (default is $HOME/.rclgo.yaml)")
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.LocalFlags())
}

// initConfig reads in the config file and ENV variables if set.
func initConfig() {
	if viper.GetString("config-file") != "" {
		viper.SetConfigFile(viper.GetString("config-file"))
	} else {
		home, err := utilities.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".rclgo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".rclgo")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
