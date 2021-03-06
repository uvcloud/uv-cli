package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yottab/cli/config"
)

// RootCmd represents the base command when called without any subcommands
var (
	flagConfirm bool
	rootCmd     = &cobra.Command{
		Use:     config.APP_NAME,
		Short:   "YOTTAb cli",
		Long:    `YOTTAb cli for client side usage`,
		Version: version,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Here you will define your flags and configuration settings.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&config.ConfigManualAddress, "config", "", "Config file (default is $HOME/.yb/config.json)")
	rootCmd.PersistentFlags().StringP(config.KEY_USER, "", "", "yottab account username")
	rootCmd.PersistentFlags().StringVarP(&password, config.KEY_PASSWORD, "", "", "yottab account password")
	rootCmd.PersistentFlags().StringP(config.KEY_HOST, "", config.DEFAULTE_CONTOROLLER, "Address of Controller. a fully-qualified controller URI")
	rootCmd.PersistentFlags().StringP(config.KEY_TOKEN, "", "", "Manual Send 'TOKEN' for Authentication")
	rootCmd.PersistentFlags().BoolVar(&flagConfirm, "confirm", false, "Confirm if some command need a confirmation, Useful for non-tty environments such as build pipelines")

	viper.BindPFlag(config.KEY_USER, rootCmd.PersistentFlags().Lookup(config.KEY_USER))
	viper.BindPFlag(config.KEY_HOST, rootCmd.PersistentFlags().Lookup(config.KEY_HOST))
	viper.BindPFlag(config.KEY_TOKEN, rootCmd.PersistentFlags().Lookup(config.KEY_TOKEN))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config.UpdateVarByConfigFile()
	command := os.Args[1]
	token := viper.GetString(config.KEY_TOKEN)
	if len(token) == 0 && command != "login" {
		log.Fatal("You must login first.")
	}

	// set token for S3.yottab.io
	setS3AccessKeyID(token)
}
