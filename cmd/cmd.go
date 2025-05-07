// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package cmd

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/mail"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"
	"github.com/asciimoo/omnom/webapp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	loginCmd = "login"
	addonCmd = "addon"
)

var cfgFile string
var cfg *config.Config

func initDB(cmd *cobra.Command, args []string) {
	initStorage()
	err := model.Init(cfg)
	if err != nil {
		panic(err)
	}
	log.Debug().Msg("DB initialization complete")
}

func initStorage() {
	sCfg := map[string]string{
		"staticDir": cfg.App.StaticDir,
	}
	err := storage.Init(cfg.Storage.Type, sCfg)
	if err != nil {
		panic(err)
	}
	log.Debug().Msg("Storage initialization complete")
}

func initLog() {
	switch cfg.App.LogLevel {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Warn().Str("Invalid config log level", cfg.App.LogLevel)
	}
	out := zerolog.ConsoleWriter{
		Out: os.Stderr,
		FormatTimestamp: func(i interface{}) string {
			return i.(string)
		},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		},
	}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(out)
}

func initMail() error {
	return mail.Init(cfg)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "omnom",
	Short:   "webpage bookmarking and snapshotting service.",
	Long:    `A webpage bookmarking and snapshotting service.`,
	Version: "v0.2.0",
}

var listenCmd = &cobra.Command{
	Use:    "listen",
	Short:  "start server",
	Long:   ``,
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		arg := "address"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.Server.Address = v
		}
		arg = "results-per-page"
		if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.App.ResultsPerPage = int64(v)
		}
		arg = "webapp-snapshotter-timeout"
		if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.App.WebappSnapshotterTimeout = int(v)
		}
		arg = "create-bookmark-from-webapp"
		if v, err := cmd.Flags().GetBool(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.App.CreateBookmarkFromWebapp = v
		}
		arg = "secure-cookie"
		if v, err := cmd.Flags().GetBool(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.Server.SecureCookie = v
		}
		arg = "db-type"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.DB.Type = v
		}
		arg = "db-connection"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.DB.Connection = v
		}
		arg = "smtp-host"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.Host = v
		}
		arg = "smtp-port"
		if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.Port = int(v)
		}
		arg = "smtp-username"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.Username = v
		}
		arg = "smtp-password"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.Password = v
		}
		arg = "smtp-sender"
		if v, err := cmd.Flags().GetString(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.Sender = v
		}
		arg = "smtp-tls"
		if v, err := cmd.Flags().GetBool(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.TLS = v
		}
		arg = "smtp-tls-allow-insecure"
		if v, err := cmd.Flags().GetBool(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.TLSAllowInsecure = v
		}
		arg = "smtp-send-timeout"
		if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.SendTimeout = int(v)
		}
		arg = "smtp-connection-timeout"
		if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
			cfg.SMTP.ConnectionTimeout = int(v)
		}
		err := initMail()
		if err != nil {
			fmt.Println("Failed to initialize mailing:", err)
			os.Exit(1)
		}
		err = initActivityPub()
		if err != nil {
			fmt.Println("Failed to initialize ActivityPub keys:", err)
			os.Exit(1)
		}
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
			fmt.Println("Cannot find user:")
			os.Exit(3)
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
			fmt.Println("Cannot create new user:", err)
			os.Exit(4)
		}
		u := model.GetUser(args[0])
		fmt.Println("User", args[0], "successfully created")
		fmt.Printf("Visit %s/login?token=%s to sign in\n", cfg.Server.Address, u.LoginToken)
	},
}

var createTokenCmd = &cobra.Command{
	Use:    "create-token [username] [token type (login/addon)]",
	Short:  "create new login/addon token for a user",
	Long:   `create-token [username] [token type (login/addon)]`,
	Args:   cobra.ExactArgs(2),
	PreRun: initDB,
	Run:    createToken,
}

var setTokenCmd = &cobra.Command{
	Use:    "set-token [username] [token type (login/addon)] [token]",
	Short:  "set new login/addon token for a user",
	Long:   `set-token [username] [token type (login/addon)] [token]`,
	Args:   cobra.ExactArgs(3),
	PreRun: initDB,
	Run:    setToken,
}

var generateAPIDocsMD = &cobra.Command{
	Use:   "generate-api-docs-md",
	Short: "Generate Markdown API documentation",
	Long:  `generate-api-docs-md`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("# API documentation\n\n")
		fmt.Printf("## Endpoints\n\n")
		for _, e := range webapp.Endpoints {
			fmt.Printf(
				"- [%s `%s`](#%s-%s)\n",
				e.Name,
				e.Method,
				strings.ReplaceAll(strings.ToLower(e.Name), " ", "-"),
				strings.ToLower(e.Method),
			)
		}
		fmt.Println()
		for _, e := range webapp.Endpoints {
			fmt.Printf("### %s `%s`\n\n", e.Name, e.Method)
			fmt.Printf("`%s %s`\n\n", e.Method, e.Path)
			fmt.Println(e.Description)
			fmt.Println()
			if e.AuthRequired {
				fmt.Printf("#### Authentication required\n\n")
			}
			if len(e.Args) > 0 {
				fmt.Println(`#### Arguments

|Name|Type|Required|Description|
|----|----|--------|-----------|`)
				for _, a := range e.Args {
					fmt.Printf("|**%s**|`%s`|`%t`|%s|\n", a.Name, a.Type, a.Required, a.Description)
				}
				fmt.Println()
			}
			fmt.Printf("---\n\n")
		}
	},
}

func createToken(cmd *cobra.Command, args []string) {
	tok := model.GenerateToken()
	changeToken(args, tok)
}

func setToken(cmd *cobra.Command, args []string) {
	changeToken(args, args[2])
}

func changeToken(args []string, tok string) {
	if args[1] != loginCmd && args[1] != addonCmd {
		fmt.Println("Invalid token type. Allowed values are 'login' or 'addon'")
		os.Exit(5)
	}
	u := model.GetUser(args[0])
	if u == nil {
		fmt.Println("User not found")
		os.Exit(5)
	}
	if args[1] == loginCmd {
		u.LoginToken = tok
		err := model.DB.Save(u).Error
		if err != nil {
			fmt.Println("Failed to set token:", err)
			os.Exit(6)
		}
	} else {
		t := &model.Token{
			UserID: u.ID,
			Text:   tok,
		}
		err := model.DB.Save(t).Error
		if err != nil {
			fmt.Println("Failed to set token:", err)
			os.Exit(7)
		}
	}
	fmt.Printf("Token %s created\n", tok)
	if args[1] == loginCmd {
		fmt.Printf("Visit %s/login?token=%s to sign in\n", cfg.Server.Address, tok)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yml", "config file (default paths: ./config.yml or $HOME/.omnomrc or $HOME/.config/omnom/config.yml)")

	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "set log level (possible options: error, warning, info, debug, trace)")
	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(createTokenCmd)
	rootCmd.AddCommand(setTokenCmd)
	rootCmd.AddCommand(showUserCmd)
	rootCmd.AddCommand(generateAPIDocsMD)

	listenCmd.Flags().StringP("address", "a", "127.0.0.1:7331", "Listen address")
	listenCmd.Flags().Uint("results-per-page", 20, "Number of bookmarks/snapshots per page")
	listenCmd.Flags().Uint("webapp-snapshotter-timeout", 15, "Timeout duration for webapp snapshotter (seconds)")
	listenCmd.Flags().Bool("create-bookmark-from-webapp", false, "Allow creating bookmarks from webapp (requires chromium)")
	listenCmd.Flags().Bool("secure-cookie", false, "Use secure cookies")
	listenCmd.Flags().String("db-type", "sqlite", "Database type")
	listenCmd.Flags().String("db-connection", "db.sqlite", "Database connection string (path for sqlite)")
	listenCmd.Flags().String("smtp-host", "", "Host of the SMTP server (leave it blank to disable SMTP)")
	listenCmd.Flags().Uint("smtp-port", 25, "Port of the SMTP server")
	listenCmd.Flags().String("smtp-username", "", "SMTP username")
	listenCmd.Flags().String("smtp-password", "", "SMTP password")
	listenCmd.Flags().String("smtp-sender", "Omnom <omnom@127.0.0.1>", "SMTP sender")
	listenCmd.Flags().Bool("smtp-tls", false, "Use TLS for SMTP")
	listenCmd.Flags().Bool("smtp-tls-allow-insecure", false, "Allow insecure TLS connections for SMTP")
	listenCmd.Flags().Uint("smtp-send-timeout", 10, "SMTP send timeout (seconds)")
	listenCmd.Flags().Uint("smtp-connection-timeout", 5, "SMTP connection timeout (seconds)")

	cobra.OnInitialize(initialize)
}

func initialize() {
	initConfig()
	initLog()
	log.Debug().Msg("Config initialization complete")
	log.Debug().Msg("Logging initialization complete")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Println("Failed to initialize config:", err)
		os.Exit(2)
	}
	if l, _ := rootCmd.PersistentFlags().GetString("log-level"); l != "" && (rootCmd.Flags().Changed("log-level") || cfg.App.LogLevel == "") {
		fmt.Println("YO")
		cfg.App.LogLevel = l
	}
}

func initActivityPub() error {
	if cfg.ActivityPub == nil || cfg.ActivityPub.PubKeyPath == "" || cfg.ActivityPub.PrivKeyPath == "" {
		return errors.New("cannot find ActivityPub config - check config.yml_sample for sample configuration")
	}
	privBytes, err := os.ReadFile(cfg.ActivityPub.PrivKeyPath)
	if err != nil {
		prvb, err := cfg.ActivityPub.ExportPrivKey()
		if err != nil {
			return err
		}
		err = os.WriteFile(cfg.ActivityPub.PrivKeyPath, prvb, 0400)
		if err != nil {
			return errors.New("failed to write privkey")
		}
		pubb, err := cfg.ActivityPub.ExportPubKey()
		if err != nil {
			return err
		}
		err = os.WriteFile(cfg.ActivityPub.PubKeyPath, pubb, 0400)
		if err != nil {
			return errors.New("failed to write pubkey")
		}
		return nil
	}
	err = cfg.ActivityPub.ParsePrivKey(privBytes)
	if err != nil {
		return err
	}
	pubBytes, err := os.ReadFile(cfg.ActivityPub.PubKeyPath)
	if err != nil {
		return errors.New("failed to read pubkey")
	}
	err = cfg.ActivityPub.ParsePubKey(pubBytes)
	if err != nil {
		return err
	}
	return nil
}
