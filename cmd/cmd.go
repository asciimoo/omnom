package cmd

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/mail"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"
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

func initStorage() {
	err := storage.Init(cfg.Storage.Type, cfg.Storage.Root)
	if err != nil {
		panic(err)
	}
}

func initMail() {
	err := mail.Init(cfg)
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
		initStorage()
		initMail()
		webapp.Run(cfg)
	},
}

var showUserCmd = &cobra.Command{
	Use:    "show-user [username]",
	Short:  "show user details",
	Long:   `show-user [username]`,
	Args:   cobra.ExactArgs(1),
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		u := model.GetUser(args[0])
		if u == nil {
			log.Println("Cannot find user:")
			os.Exit(1)
		}
		s := reflect.ValueOf(u).Elem()
		typeOfT := s.Type()

		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			fname := typeOfT.Field(i).Name
			if fname == "Model" || fname == "Bookmarks" || fname == "SubmissionTokens" {
				continue
			}
			fmt.Printf("%20s: %v\n", fname, f.Interface())
		}
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
		u := model.GetUser(args[0])
		log.Println("User", args[0], "successfully created")
		log.Printf("Visit %s/login?token=%s to sign in\n", cfg.Server.Address, u.LoginToken)
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
		if args[1] == "login" {
			log.Printf("Visit %s/login?token=%s to sign in\n", cfg.Server.Address, tok)
		}
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
	rootCmd.AddCommand(showUserCmd)
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
