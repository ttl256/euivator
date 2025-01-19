package cmd

import (
	"bufio"
	"io"
	"strings"

	"github.com/spf13/cobra"

	berrors "github.com/pkg/errors"

	"github.com/ttl256/euivator/internal/hwaddr"
)

var convertCmd = &cobra.Command{
	Use:          "convert [eui ...]",
	Short:        "Convert an EUI to a chosen representation",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r io.Reader
		if len(args) > 0 {
			r = strings.NewReader(strings.Join(args, "\n"))
		} else {
			r = cmd.InOrStdin()
		}

		return convertAction(cmd.OutOrStdout(), r, flagEUIFormat)
	},
}

func init() {
	euiCmd.AddCommand(convertCmd)
}

func convertAction(w io.Writer, r io.Reader, format EUIFormat) error {
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	convertFunc := convertFuncMap[format]

	var lineN int

	for scanner.Scan() {
		lineN++
		line := scanner.Text()
		addr, err := hwaddr.ParseAddr(line)
		if err != nil {
			return AtInputPositionError{Position: lineN, Err: err}
		}

		converted := convertFunc(addr)
		_, err = writer.WriteString(converted + "\n")
		if err != nil {
			return berrors.WithStack(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return berrors.WithStack(err)
	}

	err := writer.Flush()
	if err != nil {
		return berrors.WithStack(err)
	}

	return nil
}
