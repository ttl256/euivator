package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/netip"
	"strings"

	"github.com/spf13/cobra"

	berrors "github.com/pkg/errors"

	"github.com/ttl256/euivator/internal/hwaddr"
)

var addr6Cmd = &cobra.Command{
	Use:          "addr6 [[prefix6, [eui48, eui64]] ...]",
	Short:        "Generate an IPv6 address based on a prefix and an EUI",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var r io.Reader
		if len(args) > 0 {
			if len(args)%2 != 0 {
				return fmt.Errorf("expected an even number of arguments, got %d in %v", len(args), args)
			}
			buf := &strings.Builder{}
			for i := 1; i < len(args); i += 2 {
				buf.WriteString(strings.Join([]string{args[i-1], args[i]}, " ") + "\n")
			}
			r = strings.NewReader(buf.String())
		} else {
			r = cmd.InOrStdin()
		}

		return addr6Action(cmd.OutOrStdout(), r)
	},
}

func init() {
	euiCmd.AddCommand(addr6Cmd)
}

func addr6Action(w io.Writer, r io.Reader) error {
	const numFields = 2
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)

	var lineN int

	for scanner.Scan() {
		lineN++
		line := scanner.Text()
		lineFields := strings.Fields(line)
		if len(lineFields) != numFields {
			return fmt.Errorf("expected %d fields, got %d in %q", numFields, len(lineFields), line)
		}

		prefixRaw := lineFields[0]
		euiRaw := lineFields[1]

		prefix, err := netip.ParsePrefix(prefixRaw)
		if err != nil {
			return AtInputPositionError{Position: lineN, Err: err}
		}
		eui, err := hwaddr.ParseAddr(euiRaw)
		if err != nil {
			return AtInputPositionError{Position: lineN, Err: err}
		}

		var eui64 hwaddr.EUI64

		if len(eui) == hwaddr.EUI48Len {
			eui48, _ := hwaddr.EUI48FromBytes(eui)
			eui64 = eui48.EUI64Modified()
		} else {
			eui64, err = hwaddr.EUI64FromBytes(eui)
			if err != nil {
				return AtInputPositionError{Position: lineN, Err: err}
			}
		}

		addr := hwaddr.AppendToPrefix(prefix, eui64)
		_, err = writer.WriteString(addr.String() + "\n")
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
