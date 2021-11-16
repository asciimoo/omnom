package cmd

import (
	"log"
	"os"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/webapp"

	"github.com/spf13/cobra"
)

var cfgFile string
var cfg *config.Config

func initDB(cmd *cobra.Command, args []string) {
	err := model.Init(cfg)
	if err != nil {
		panic(err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "omnom",
	Short:   "webpage bookmarking and snapshotting service.",
	Long:    `A webpage bookmarking and snapshotting service.`,
	Version: "v0.1.0",
}

var listenCmd = &cobra.Command{
	Use:    "listen",
	Short:  "start server",
	Long:   ``,
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		webapp.Run(cfg)
	},
}

var createUserCmd = &cobra.Command{
	Use:    "create-user [username] [email]",
	Short:  "create new user",
	Long:   `create-user [username] [email]`,
	Args:   cobra.ExactArgs(2),
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		err := model.CreateUser(args[0], args[1])
		if err != nil {
			log.Println("Cannot create new user:", err)
			os.Exit(1)
		}
		log.Println("User", args[0], "successfully created")
	},
}

var createTokenCmd = &cobra.Command{
	Use:    "create-token [username] [token type (login/addon)]",
	Short:  "create new user token",
	Long:   `create-token [username] [token type (login/addon)]`,
	Args:   cobra.ExactArgs(2),
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		if args[1] != "login" && args[1] != "addon" {
			log.Println("Invalid token type. Allowed values are 'login' or 'addon'")
			os.Exit(1)
		}
		u := model.GetUser(args[0])
		if u == nil {
			log.Println("User not found")
			os.Exit(1)
		}
		tok := model.GenerateToken()
		if args[1] == "login" {
			u.LoginToken = tok
			err := model.DB.Save(u).Error
			if err != nil {
				log.Println("Failed to set token:", err)
				os.Exit(1)
			}
		} else {
			t := &model.Token{
				UserID: u.ID,
				Text:   tok,
			}
			err := model.DB.Save(t).Error
			if err != nil {
				log.Println("Failed to set token:", err)
				os.Exit(1)
			}
		}
		log.Printf("Token %s created\n", tok)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yml", "config file (default is config.yml)")

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "turn on debug mode")
	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(createTokenCmd)
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if b, _ := rootCmd.PersistentFlags().GetBool("debug"); b {
		cfg.App.Debug = true
	}
	if err != nil {
		panic(err)
	}
}
