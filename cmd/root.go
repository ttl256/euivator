package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultConfigPath = filepath.Join(xdg.ConfigHome, appName)
	logger            = new(slog.Logger)
	logLevel          = new(slog.LevelVar)
	version           = "dev" // Set from ldflags
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "A utility to work with EUIs",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		debug := viper.GetBool("debug")
		if debug {
			logLevel.Set(slog.LevelDebug)
			logger.LogAttrs(cmd.Context(), slog.LevelDebug, "verbose logging enabled") //nolint: sloglint // let me be
		}

		return nil
	},
	Version: version,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var loggerOptions = new(slog.HandlerOptions)
	loggerOptions.Level = logLevel
	logLevel.Set(slog.LevelInfo)
	logger = slog.New(slog.NewTextHandler(os.Stdout, loggerOptions))

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		fmt.Sprintf("config file (default is %s)", filepath.Join(
			defaultConfigPath, strings.Join([]string{appName, "yaml"}, ".")),
		),
	)
	rootCmd.PersistentFlags().Bool("debug", false, "enable verbose logging")
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("EUIVATOR")

	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName(appName)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, new(viper.ConfigFileNotFoundError)) {
			cobra.CheckErr(err)
		}
	}
}
