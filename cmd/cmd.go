package cmd

import (
	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/model"
	"github.com/asciimoo/omnom/webapp"

	"github.com/spf13/cobra"
)

var cfgFile string
var cfg *config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "omnom",
	Short: "A webpage bookmarking and snapshotting service.",
	Long:  `A webpage bookmarking and snapshotting service.`,
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Start server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		webapp.Run(cfg)
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

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Turn on debug mode")
	rootCmd.AddCommand(listenCmd)
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
	err = model.Init(cfg)
	if err != nil {
		panic(err)
	}
}
