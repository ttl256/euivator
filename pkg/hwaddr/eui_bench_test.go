package hwaddr_test

import (
	"encoding/hex"
	"testing"

	"github.com/ttl256/euivator/pkg/hwaddr"
)

var resultString string //nolint: gochecknoglobals // avoid compiler optimization

func BenchmarkToString(b *testing.B) {
	cases := [][]byte{
		{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC},
		{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC, 0x33, 0xDD},
	}

	for _, tt := range cases {
		b.Run(hex.EncodeToString(tt), func(b *testing.B) {
			for range b.N {
				resultString = hwaddr.ToString(tt, []byte{':'}, 1)
			}
		})
	}
}

func stringsForBenchmark() []string {
	return []string{
		"00:AA:11:BB:22:CC",
		"00-AA-11-BB-22-CC",
		"00AA.11BB.22CC",
		"00AA11BB22CC",
	}
}

var SliceResult []byte //nolint: gochecknoglobals // avoid compiler optimization

func BenchmarkParse(b *testing.B) {
	for _, input := range stringsForBenchmark() {
		b.Run(input, func(b *testing.B) {
			for range b.N {
				SliceResult, _ = hwaddr.ParseAddr(input)
			}
		})
	}
}

var EUI64Result hwaddr.EUI64 //nolint: gochecknoglobals // avoid compiler optimization

func BenchmarkEUI64Modified(b *testing.B) {
	m := hwaddr.EUI48{0x00, 0xAA, 0x11, 0xBB, 0x22, 0xCC}

	for range b.N {
		EUI64Result = m.EUI64Modified()
	}
}
