package cmd

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cacheDir = filepath.Join(xdg.CacheHome, appName)

var ouiCmd = &cobra.Command{
	Use:   "oui",
	Short: "Interact with OUI database",
}

func init() {
	rootCmd.AddCommand(ouiCmd)
	ouiCmd.PersistentFlags().String("cachedir", cacheDir, "Directory of the utility cache files")

	_ = viper.BindPFlag("cachedir", ouiCmd.PersistentFlags().Lookup("cachedir"))
	viper.SetDefault("cachedir", cacheDir)
}
