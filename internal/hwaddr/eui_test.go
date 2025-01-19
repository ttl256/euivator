package hwaddr_test

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ttl256/euivator/internal/hwaddr"
)

func TestParseAddr(t *testing.T) {
	t.Parallel()
	validAddrCases := []struct {
		input string
		want  []byte
	}{
		{"00:AA:11:BB:22:CC", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00-AA-11-BB-22-CC", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00AA.11BB.22CC", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00AA11BB22CC", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00:00:00:00:00:00", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"FF:FF:FF:FF:FF:FF", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"00:AA:11:BB:22:CC:33:DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00-AA-11-BB-22-CC-33-DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00AA.11BB.22CC.33DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00AA11BB22CC33DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00:00:00:00:00:00:00:00", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"FF:FF:FF:FF:FF:FF:FF:FF", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"00:aa:11:BB:22:cc", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00-AA-11-bb-22-CC", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00aa.11BB.22cc", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00aa11BB22cc", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}},
		{"00:aa:11:BB:22:cc:33:DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00-AA-11-bb-22-CC-33-dd", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00aa.11BB.22cc.33DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
		{"00aa11BB22cc33DD", []byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}},
	}

	for _, tt := range validAddrCases {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			addr, err := hwaddr.ParseAddr(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, addr)
		})
	}
}

func TestParseAddrInvalid(t *testing.T) {
	invalidAddrCases := []struct {
		input string
		msg   string
		err   error
	}{
		{"0", "short", hwaddr.ErrInputTooShort},
		{"AA-AA-AA-AA-AA-AA-AA-AA-AA", "long", hwaddr.ErrInputTooLong},
		{"00:00:00:00:00:00:", "unbalanced", hwaddr.ErrInputUnbalanced},
		{"00:00:00:00:00:00:00", "odd number of bytes", hwaddr.ErrInputUnexpectedNumBytes},
		{"0000.0000.0000.", "unbalanced", hwaddr.ErrInputUnbalanced},
		{"0000.0000.0000.00", "unbalanced", hwaddr.ErrInputUnbalanced},
		{"0000000000000", "unbalanced", hwaddr.ErrInputUnbalanced},
		{"00000000000000", "odd number of bytes", hwaddr.ErrInputUnexpectedNumBytes},
	}

	for _, tt := range invalidAddrCases {
		t.Run(tt.msg+tt.input, func(t *testing.T) {
			_, err := hwaddr.ParseAddr(tt.input)
			require.ErrorAs(t, err, new(hwaddr.ParseError))
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			}
		})
	}

	invalidByteCases := []struct {
		input string
		msg   string
		err   hex.InvalidByteError
	}{
		{"00:00:00:0T:00:00", "contains non-hex", hex.InvalidByteError(0)},
		{"0000.000T.0000", "contains non-hex", hex.InvalidByteError(0)},
		{"0000000T0000", "contains non-hex", hex.InvalidByteError(0)},
	}

	for _, tt := range invalidByteCases {
		t.Run(tt.msg+tt.input, func(t *testing.T) {
			_, err := hwaddr.ParseAddr(tt.input)
			require.ErrorAs(t, err, new(hwaddr.ParseError))
			require.ErrorAs(t, err, &tt.err)
		})
	}
}

func TestToString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input []byte
		sep   []byte
		count int
		want  string
	}{
		{[]byte{}, []byte{':'}, 1, ""},
		{[]byte{0x00}, []byte{':'}, 1, "00"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}, []byte{':'}, 1, "00:aa:11:bb:22:cc"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}, []byte{'-'}, 1, "00-aa-11-bb-22-cc"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}, []byte{'.'}, 2, "00aa.11bb.22cc"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}, []byte{}, 0, "00aa11bb22cc"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, []byte{':'}, 1, "00:00:00:00:00:00"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, []byte{':'}, 1, "ff:ff:ff:ff:ff:ff"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}, []byte{':'}, 1, "00:aa:11:bb:22:cc:33:dd"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}, []byte{'-'}, 1, "00-aa-11-bb-22-cc-33-dd"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}, []byte{'.'}, 2, "00aa.11bb.22cc.33dd"},
		{[]byte{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD}, []byte{}, 0, "00aa11bb22cc33dd"},
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, []byte{':'}, 1, "00:00:00:00:00:00:00:00"},
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, []byte{':'}, 1, "ff:ff:ff:ff:ff:ff:ff:ff"},
	}

	for _, tt := range cases {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			got := hwaddr.ToString(tt.input, tt.sep, tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEUI64Modified(t *testing.T) {
	t.Parallel()

	inputTests := []struct {
		input hwaddr.EUI48
		want  hwaddr.EUI64
	}{
		{hwaddr.EUI48{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			hwaddr.EUI64{0x02, 0x00, 0x00, 0xFF, 0xFE, 0x00, 0x00, 0x00}},
		{hwaddr.EUI48{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			hwaddr.EUI64{0xFD, 0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFF}},
		{hwaddr.EUI48{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			hwaddr.EUI64{0xA8, 0xBB, 0xCC, 0xFF, 0xFE, 0xDD, 0xEE, 0xFF}},
	}

	for i, tt := range inputTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			got := tt.input.EUI64Modified()
			assert.Equal(t, tt.want, got)
		})
	}
}
