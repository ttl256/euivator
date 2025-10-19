/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"io"
	"strings"

	"github.com/spf13/cobra"

	berrors "github.com/pkg/errors"

	"github.com/ttl256/euivator/pkg/hwaddr"
)

var verifyCmd = &cobra.Command{
	Use:   "verify [eui ...]",
	Short: "Verify an EUI",
	Long: `Verify an EUI. Supported formats:
XX:XX:XX:XX:XX:XX
XX-XX-XX-XX-XX-XX
XXXX.XXXX.XXXX
XXXXXXXXXXXX`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r io.Reader
		if len(args) > 0 {
			r = strings.NewReader(strings.Join(args, "\n"))
		} else {
			r = cmd.InOrStdin()
		}

		return verifyAction(r)
	},
}

func init() {
	euiCmd.AddCommand(verifyCmd)
}

func verifyAction(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	var lineN int

	for scanner.Scan() {
		lineN++
		line := scanner.Text()
		_, err := hwaddr.ParseAddr(line)
		if err != nil {
			return AtInputPositionError{Position: lineN, Err: err}
		}
	}

	if err := scanner.Err(); err != nil {
		return berrors.WithStack(err)
	}

	return nil
}
