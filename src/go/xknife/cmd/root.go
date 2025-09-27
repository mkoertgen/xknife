package cmd

import (
	"errors"
	"fmt"
	"github.com/michimani/gotwi"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "x-knife",
	Short: "X-Knife is a tool for the Twitter/X-API v2",
	Long:  `A command line tool to detect bots on Twitter using the Twitter API v2.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to X-Knife!")
	},
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing X-Knife '%s'\n", err)
		os.Exit(1)
	}
}

var cfgFile string
var userName string
var userId string
var pageSize int
var xClient *gotwi.Client

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(func() {
		if err := initTwitterClient(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize Twitter client: %v\n", err)
			// Don't exit here, let the commands handle the error
		}
	})
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.xknife.yaml)")
	rootCmd.PersistentFlags().StringVarP(&userName, "user", "u", "mkoertg", "X user account name")
	rootCmd.PersistentFlags().StringVar(&userId, "id", "", "X user id")
	rootCmd.PersistentFlags().IntVarP(&pageSize, "size", "s", 20, "Number of items to list at once")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("id", rootCmd.PersistentFlags().Lookup("id"))
	viper.BindPFlag("size", rootCmd.PersistentFlags().Lookup("size"))
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".xknife" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".xknife")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found; ignore error if desired
			//fmt.Println("Config file not found.")
			if cfgFile != "" {
				fmt.Println("Specified config file", cfgFile, "not found.")
				os.Exit(1)
			}
		}
	}
}

func initTwitterClient() error {
	client, err := gotwi.NewClient(&gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth2BearerToken,
	})
	
	if err != nil {
		return fmt.Errorf("failed to create Twitter client: %v", err)
	}
	
	xClient = client
	return nil
}
