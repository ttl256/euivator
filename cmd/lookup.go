/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	berrors "github.com/pkg/errors"

	"github.com/ttl256/euivator/internal/hwaddr"
	"github.com/ttl256/euivator/internal/registry"
)

type RecordResponse struct {
	Input    string            `json:"input"`
	InputRaw string            `json:"input_raw"`
	Records  []registry.Record `json:"records"`
}

var lookupCmd = &cobra.Command{
	Use:          "lookup [hex_prefix ...]",
	Short:        "Lookup an EUI/hex prefix in the OUI database",
	Long:         recordsResponseExample(),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r io.Reader
		if len(args) > 0 {
			r = strings.NewReader(strings.Join(args, "\n"))
		} else {
			r = cmd.InOrStdin()
		}

		return lookupAction(cmd.OutOrStdout(), r)
	},
}

func init() {
	ouiCmd.AddCommand(lookupCmd)
}

func lookupAction(w io.Writer, r io.Reader) error {
	lookupFile := filepath.Join(viper.GetString("cachedir"), LookupFile)
	f, err := os.Open(lookupFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf(
				"unable to open %q. try running '%s oui update' to prepare required cache",
				lookupFile,
				appName,
			)
		}
		return berrors.WithStack(err)
	}
	defer f.Close()

	trie := registry.NewTrie()
	err = trie.DecodeGOB(f)
	if err != nil {
		return fmt.Errorf("loading lookup database: %w", err)
	}

	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)

	var prefix string
	var data []byte
	for scanner.Scan() {
		line := scanner.Text()
		prefix, err = stringToHexPrefix(line)
		if err != nil {
			return err
		}

		records := trie.LongestPrefixMatch(prefix)
		result := RecordResponse{
			Input:    prefix,
			InputRaw: line,
			Records:  records,
		}
		data, err = json.Marshal(result)
		if err != nil {
			return berrors.WithStack(err)
		}
		data = append(data, '\n')

		_, err = writer.Write(data)
		if err != nil {
			return berrors.WithStack(err)
		}
	}

	if scanner.Err() != nil {
		return berrors.WithStack(scanner.Err())
	}

	err = writer.Flush()
	if err != nil {
		return berrors.WithStack(err)
	}

	return nil
}

func stringToHexPrefix(s string) (string, error) {
	buf := new(strings.Builder)
	for _, r := range s {
		if _, ok := validDelimeters[r]; ok {
			continue
		}
		char := unicode.ToUpper(r)
		if _, ok := validChars[char]; ok {
			buf.WriteRune(char)
		} else {
			return "", fmt.Errorf("invalid input character %q in %q", r, s)
		}
	}
	if buf.Len() > hwaddr.EUI64HexLen {
		return "", fmt.Errorf("expected sanitized input to be <= %d, got %d in %q", hwaddr.EUI64HexLen, buf.Len(), s)
	}

	return buf.String(), nil
}

var validDelimeters = map[rune]struct{}{
	':': {},
	'-': {},
	'.': {},
}

var (
	validChars = func() map[rune]struct{} {
		chars := map[rune]struct{}{}
		for _, r := range "0123456789ABCDEF" {
			chars[r] = struct{}{}
		}
		return chars
	}()
)

func recordsResponseExample() string {
	data, err := json.MarshalIndent(RecordResponse{
		Input:    "286FB9",
		InputRaw: "28:6f:b9:11:22:33",
		Records: []registry.Record{
			{
				Assignment: "286FB9",
				Registry:   registry.NameMAL,
				OrgName:    "Nokia Shanghai Bell Co., Ltd.",
				OrgAddress: "No.388 Ning Qiao Road,Jin Qiao Pudong Shanghai Shanghai   CN 201206",
			},
		},
	}, "", "  ")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`Lookup an EUI/hex prefix in the OUI database. Valid input is any hex string
with separators from [-:.]. Output is a JSON. Example of the output:
%s
The records key contains a list of zero-length when allocation is not found`, data)
}
