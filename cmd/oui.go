package cmd

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

const appName = "euivator"

var cacheDir = filepath.Join(xdg.CacheHome, appName)

var ouiCmd = &cobra.Command{
	Use:   "oui",
	Short: "Interact with OUI database",
}

func init() {
	rootCmd.AddCommand(ouiCmd)
	ouiCmd.PersistentFlags().String("dir", cacheDir, "Directory for the utility cache files")
}
