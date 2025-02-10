package registry_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ttl256/euivator/internal/registry"
)

func TestParseRecordsFromCSV(t *testing.T) {
	cases := []struct {
		input string
		want  []registry.Record
	}{
		{`Registry,Assignment,Organization Name,Organization Address
MA-S,8C1F64ABA,"COOL DEVICES, INC",32 NORTHWESTERN HH AA US 01079 
MA-S,8C1B649B9,"EVEN COOLER DEVICES, S.L.",Av. Onze de Setembre 13 Reus Tarragona ES 49203 
MA-S,8C1F6480A,ASDF Corporation,"Address: 20F.-1, No. 8, County TW 30244 "
`, []registry.Record{
			{
				Registry:   registry.NameMAS,
				Assignment: "8C1F64ABA",
				OrgName:    "COOL DEVICES, INC",
				OrgAddress: "32 NORTHWESTERN HH AA US 01079",
			},
			{
				Registry:   registry.NameMAS,
				Assignment: "8C1B649B9",
				OrgName:    "EVEN COOLER DEVICES, S.L.",
				OrgAddress: "Av. Onze de Setembre 13 Reus Tarragona ES 49203",
			},
			{
				Registry:   registry.NameMAS,
				Assignment: "8C1F6480A",
				OrgName:    "ASDF Corporation",
				OrgAddress: "Address: 20F.-1, No. 8, County TW 30244",
			},
		},
		},
	}
	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := registry.ParseRecordsFromCSV(strings.NewReader(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
