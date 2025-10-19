package cmd

import (
	"bufio"
	"io"
	"strings"

	"github.com/spf13/cobra"

	berrors "github.com/pkg/errors"

	"github.com/ttl256/euivator/pkg/hwaddr"
)

var modifiedCmd = &cobra.Command{
	Use:          "modified [eui48 ...]",
	Short:        "Generate an EUI64 modified from an EUI48",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r io.Reader
		if len(args) > 0 {
			r = strings.NewReader(strings.Join(args, "\n"))
		} else {
			r = cmd.InOrStdin()
		}

		return modifiedAction(cmd.OutOrStdout(), r, flagEUIFormat)
	},
}

func init() {
	euiCmd.AddCommand(modifiedCmd)
}

func modifiedAction(w io.Writer, r io.Reader, format EUIFormat) error {
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

		eui48, err := hwaddr.EUI48FromBytes(addr)
		if err != nil {
			return AtInputPositionError{Position: lineN, Err: err}
		}

		eui64 := eui48.EUI64Modified()

		_, err = writer.WriteString(convertFunc(eui64[:]) + "\n")
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
