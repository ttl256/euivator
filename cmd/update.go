package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	berrors "github.com/pkg/errors"

	"github.com/ttl256/euivator/internal/fetcher"
	"github.com/ttl256/euivator/internal/registry"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update OUI database",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, _ []string) error {
		return updateAction(cmd, viper.GetString("cachedir"), logger)
	},
}

func init() {
	ouiCmd.AddCommand(updateCmd)
	updateCmd.Flags().Bool("force", false, "bypass client-caching (ETags)")
}

func updateAction(cmd *cobra.Command, dir string, logger *slog.Logger) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return berrors.WithStack(err)
	}

	fetch := fetcher.New(fetcher.GetSources(), dir, logger)
	err = fetch.DownloadFiles(cmd.Context(), !force)
	if err != nil {
		return fmt.Errorf("updating OUI database: %w", err)
	}
	logger.LogAttrs(cmd.Context(), slog.LevelDebug, "prepared all CSV files")

	trie := registry.NewTrie()

	for _, file := range registry.NameNames() {
		err = appendFromCVStoTrie(trie, filepath.Join(dir, strings.Join([]string{file, "csv"}, ".")))
		if err != nil {
			return err
		}
	}
	logger.LogAttrs(cmd.Context(), slog.LevelDebug, "prepared lookup data structure")

	f, err := os.OpenFile(
		filepath.Join(dir, LookupFile),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return berrors.WithStack(err)
	}
	defer f.Close()

	err = trie.EncodeGOB(f)
	if err != nil {
		return berrors.WithStack(err)
	}
	logger.LogAttrs(cmd.Context(), slog.LevelDebug, "dumped lookup data structure on disk")
	logger.LogAttrs(cmd.Context(), slog.LevelInfo, "all done")

	return nil
}

func appendFromCVStoTrie(t *registry.Trie, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return berrors.WithStack(err)
	}
	defer f.Close()

	records, err := registry.ParseRecordsFromCSV(f)
	if err != nil {
		return berrors.WithStack(err)
	}

	t.InsertMany(records)
	return nil
}
