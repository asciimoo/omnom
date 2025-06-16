// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package cmd

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/feed"
	"github.com/asciimoo/omnom/mail"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/storage"
	"github.com/asciimoo/omnom/webapp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var fs embed.FS

const (
	loginCmd = "login"
	addonCmd = "addon"
)

var cfgFile string
var cfg *config.Config

func exit(errno int, msg string) {
	if errno != 0 {
		fmt.Println("Error!")
	}
	fmt.Println(msg)
	os.Exit(errno)
}

func initDB(cmd *cobra.Command, args []string) {
	err := model.Init(cfg)
	if err != nil {
		exit(1, err.Error())
	}
	log.Debug().Msg("DB initialization complete")
}

func initStorage() {
	err := storage.Init(cfg.Storage)
	if err != nil {
		exit(1, err.Error())
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

func setStrArg(cmd *cobra.Command, arg string, dest *string) {
	if v, err := cmd.Flags().GetString(arg); err == nil && (cmd.Flags().Changed(arg) || *dest == "") {
		*dest = v
	}
}

func setBoolArg(cmd *cobra.Command, arg string, dest *bool) {
	if v, err := cmd.Flags().GetBool(arg); err == nil && cmd.Flags().Changed(arg) {
		*dest = v
	}
}

func setIntArg(cmd *cobra.Command, arg string, dest *int) {
	if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
		//nolint: gosec // conversion is safe. TODO use uint by default
		*dest = int(v)
	}
}

func setInt64Arg(cmd *cobra.Command, arg string, dest *int64) {
	if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
		//nolint: gosec // conversion is safe. TODO use uint by default
		*dest = int64(v)
	}
}

func setUintArg(cmd *cobra.Command, arg string, dest *uint) {
	if v, err := cmd.Flags().GetUint(arg); err == nil && cmd.Flags().Changed(arg) {
		*dest = v
	}
}

var listenCmd = &cobra.Command{
	Use:    "listen",
	Short:  "start server",
	Long:   ``,
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		setStrArg(cmd, "address", &cfg.Server.Address)
		setInt64Arg(cmd, "results-per-page", &cfg.App.ResultsPerPage)
		setIntArg(cmd, "webapp-snapshotter-timeout", &cfg.App.WebappSnapshotterTimeout)
		setBoolArg(cmd, "create-bookmark-from-webapp", &cfg.App.CreateBookmarkFromWebapp)
		setBoolArg(cmd, "secure-cookie", &cfg.Server.SecureCookie)
		setStrArg(cmd, "db-type", &cfg.DB.Type)
		setStrArg(cmd, "db-connection", &cfg.DB.Connection)
		setStrArg(cmd, "smtp-host", &cfg.SMTP.Host)
		setIntArg(cmd, "smtp-port", &cfg.SMTP.Port)
		setStrArg(cmd, "smtp-username", &cfg.SMTP.Username)
		setStrArg(cmd, "smtp-password", &cfg.SMTP.Password)
		setStrArg(cmd, "smtp-sender", &cfg.SMTP.Sender)
		setBoolArg(cmd, "smtp-tls", &cfg.SMTP.TLS)
		setBoolArg(cmd, "smtp-tls-allow-insecure", &cfg.SMTP.TLSAllowInsecure)
		setIntArg(cmd, "smtp-send-timeout", &cfg.SMTP.SendTimeout)
		setIntArg(cmd, "smtp-connection-timeout", &cfg.SMTP.ConnectionTimeout)
		setUintArg(cmd, "feed-items-per-page", &cfg.Feed.ItemsPerPage)
		if v, err := cmd.Flags().GetString("data-directory"); err == nil && cmd.Flags().Changed("data-directory") {
			if cfg.Storage.Filesystem == nil {
				cfg.Storage.Filesystem = &config.StorageFilesystem{}
			}
			cfg.Storage.Filesystem.RootDir = v
		}
		initStorage()
		err := initMail()
		if err != nil {
			exit(1, "Failed to initialize mailing: "+err.Error())
		}
		err = initActivityPub()
		if err != nil {
			exit(1, "Failed to initialize ActivityPub keys: "+err.Error())
		}
		webapp.Run(cfg)
	},
}

var showUserCmd = &cobra.Command{
	Use:    "show-user USERNAME",
	Short:  "show user details",
	Long:   `show-user USERNAME`,
	Args:   cobra.ExactArgs(1),
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		u := model.GetUser(args[0])
		if u == nil {
			exit(1, "Cannot find user")
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
	Use:    "create-user USERNAME EMAIL",
	Short:  "create new user",
	Long:   `create-user USERNAME EMAIL`,
	Args:   cobra.ExactArgs(2),
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		err := model.CreateUser(args[0], args[1])
		if err != nil {
			exit(1, "Cannot create new user: "+err.Error())
		}
		u := model.GetUser(args[0])
		fmt.Println("User", args[0], "successfully created")
		fmt.Printf("Visit %s/login?token=%s to sign in\n", cfg.Server.Address, u.LoginToken)
	},
}

var createTokenCmd = &cobra.Command{
	Use:    "create-token USERNAME (addon|login)",
	Short:  "create new login/addon token for a user",
	Long:   `create-token USERNAME <token type: login | addon>`,
	Args:   cobra.ExactArgs(2),
	PreRun: initDB,
	Run:    createToken,
}

var createBookmarkCmd = &cobra.Command{
	Use:    "create-bookmark USERNAME TITLE URL",
	Short:  "create new bookmark",
	Long:   `create-bookmark USERNAME TITLE URL`,
	Args:   cobra.ExactArgs(3),
	PreRun: initDB,
	Run:    createBookmark,
}

var createConfigCmd = &cobra.Command{
	Use:   "create-config FILENAME",
	Short: "create default configuration file",
	Long:  `create-config FILENAME`,
	Args:  cobra.ExactArgs(1),
	Run:   createConfig,
}

var setTokenCmd = &cobra.Command{
	Use:    "set-token USERNAME (addon|login) TOKEN",
	Short:  "set new login/addon token for a user",
	Long:   `set-token USERNAME <token type: login | addon> TOKEN`,
	Args:   cobra.ExactArgs(3),
	PreRun: initDB,
	Run:    setToken,
}

var updateFeedsCmd = &cobra.Command{
	Use:    "update-feeds",
	Short:  "update RSS/Atom feeds",
	Long:   `update-feeds`,
	Args:   cobra.ExactArgs(0),
	PreRun: initDB,
	Run: func(cmd *cobra.Command, args []string) {
		err := feed.Update()
		if err != nil {
			exit(1, "Failed to update feeds: "+err.Error())
		}
		fmt.Println("Successful feed update")
	},
}

var generateAPIDocsMDCmd = &cobra.Command{
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
		exit(1, "Invalid token type. Allowed values are 'login' or 'addon'")
	}
	u := model.GetUser(args[0])
	if u == nil {
		exit(1, "User not found")
	}
	if args[1] == loginCmd {
		u.LoginToken = tok
		err := model.DB.Save(u).Error
		if err != nil {
			exit(1, "Failed to set token: "+err.Error())
		}
	} else {
		t := &model.Token{
			UserID: u.ID,
			Text:   tok,
		}
		err := model.DB.Save(t).Error
		if err != nil {
			exit(1, "Failed to set token: "+err.Error())
		}
	}
	fmt.Printf("Token %s created\n", tok)
	if args[1] == loginCmd {
		fmt.Printf("Visit %s/login?token=%s to sign in\n", cfg.Server.Address, tok)
	}
}

func createConfig(cmd *cobra.Command, args []string) {
	fname := args[0]
	if _, err := os.Stat(fname); err == nil {
		exit(1, fmt.Sprintf(`File "%s" already exists`, fname))
	}
	fc, err := fs.ReadFile("config.yml_sample")
	if err != nil {
		exit(1, `Cannot read sample config: `+err.Error())
	}
	if err := os.WriteFile(fname, fc, 0600); err != nil {
		exit(1, `Failed to create config file: `+err.Error())
	}
	fmt.Println("Config file created")
}

func createBookmark(cmd *cobra.Command, args []string) {
	// TODO add snapshot if server side snapshotting is configured
	u := model.GetUser(args[0])
	if u == nil {
		exit(1, "User not found")
	}
	tags := ""
	if v, err := cmd.Flags().GetString("tags"); err == nil {
		tags = v
	}
	notes := ""
	if v, err := cmd.Flags().GetString("notes"); err == nil {
		notes = v
	}
	public := ""
	if v, err := cmd.Flags().GetBool("public"); err == nil {
		if v {
			public = "1"
		}
	}
	unread := ""
	if v, err := cmd.Flags().GetBool("unread"); err == nil {
		if v {
			unread = "1"
		}
	}
	collection := ""
	if v, err := cmd.Flags().GetString("collection"); err == nil {
		collection = v
	}
	_, new, err := model.GetOrCreateBookmark(
		u,
		args[2],
		args[1],
		tags,
		notes,
		public,
		"",
		collection,
		unread,
	)
	if err != nil {
		exit(1, "Failed to add bookmark: "+err.Error())
	}
	if !new {
		exit(0, "Bookmark already exists")
	}
	fmt.Println("Bookmark successfully added")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(f embed.FS) {
	fs = f
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
	rootCmd.AddCommand(generateAPIDocsMDCmd)
	rootCmd.AddCommand(createConfigCmd)
	rootCmd.AddCommand(createBookmarkCmd)
	rootCmd.AddCommand(updateFeedsCmd)

	dcfg := config.CreateDefaultConfig()
	listenCmd.Flags().StringP("address", "a", dcfg.Server.Address, "Listen address")
	listenCmd.Flags().String("data-directory", "./static/data", "Data directory location to store snapshots and resources using file system storage")
	//nolint: gosec // conversion is safe. TODO use uint by default
	listenCmd.Flags().Uint("results-per-page", uint(dcfg.App.ResultsPerPage), "Number of bookmarks/snapshots per page")
	//nolint: gosec // conversion is safe. TODO use uint by default
	listenCmd.Flags().Uint("webapp-snapshotter-timeout", uint(dcfg.App.WebappSnapshotterTimeout), "Timeout duration for webapp snapshotter (seconds)")
	listenCmd.Flags().Bool("create-bookmark-from-webapp", dcfg.App.CreateBookmarkFromWebapp, "Allow creating bookmarks from webapp (requires chromium)")
	listenCmd.Flags().Bool("secure-cookie", dcfg.Server.SecureCookie, "Use secure cookies")
	listenCmd.Flags().String("db-type", dcfg.DB.Type, "Database type")
	listenCmd.Flags().String("db-connection", dcfg.DB.Connection, "Database connection string (path for sqlite)")
	listenCmd.Flags().String("smtp-host", dcfg.SMTP.Host, "Host of the SMTP server (leave it blank to disable SMTP)")
	//nolint: gosec // conversion is safe. TODO use uint by default
	listenCmd.Flags().Uint("smtp-port", uint(dcfg.SMTP.Port), "Port of the SMTP server")
	listenCmd.Flags().String("smtp-username", dcfg.SMTP.Username, "SMTP username")
	listenCmd.Flags().String("smtp-password", dcfg.SMTP.Password, "SMTP password")
	listenCmd.Flags().String("smtp-sender", dcfg.SMTP.Sender, "SMTP sender")
	listenCmd.Flags().Bool("smtp-tls", dcfg.SMTP.TLS, "Use TLS for SMTP")
	listenCmd.Flags().Bool("smtp-tls-allow-insecure", dcfg.SMTP.TLSAllowInsecure, "Allow insecure TLS connections for SMTP")
	//nolint: gosec // conversion is safe. TODO use uint by default
	listenCmd.Flags().Uint("smtp-send-timeout", uint(dcfg.SMTP.SendTimeout), "SMTP send timeout (seconds)")
	//nolint: gosec // conversion is safe. TODO use uint by default
	listenCmd.Flags().Uint("smtp-connection-timeout", uint(dcfg.SMTP.ConnectionTimeout), "SMTP connection timeout (seconds)")
	listenCmd.Flags().Uint("feed-items-per-page", dcfg.Feed.ItemsPerPage, "Number of feed items per page")

	createBookmarkCmd.Flags().Bool("public", true, "Set bookmark to public or private")
	createBookmarkCmd.Flags().Bool("unread", false, "Mark bookmark as unread")
	createBookmarkCmd.Flags().String("tags", "", "Comma separated list of tags")
	createBookmarkCmd.Flags().String("notes", "", "Bookmark notes")
	createBookmarkCmd.Flags().String("collection", "", "Collection name")

	cobra.OnInitialize(initialize)

	out := zerolog.ConsoleWriter{
		Out: os.Stderr,
		FormatTimestamp: func(i interface{}) string {
			return i.(string)
		},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		},
	}
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		dir, fn := filepath.Split(file)
		if dir == "" {
			return fn + ":" + strconv.Itoa(line)
		}
		_, subdir := filepath.Split(strings.TrimSuffix(dir, "/"))
		return subdir + "/" + fn + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(out)
}

func initialize() {
	initConfig()
	initLog()
	log.Debug().Str("filename", cfg.Filename()).Msg("Config initialization complete")
	log.Debug().Msg("Logging initialization complete")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		exit(1, "Failed to initialize config:"+err.Error())
	}
	if l, _ := rootCmd.PersistentFlags().GetString("log-level"); l != "" && (rootCmd.Flags().Changed("log-level") || cfg.App.LogLevel == "") {
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
