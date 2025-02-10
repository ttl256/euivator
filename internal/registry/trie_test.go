package registry_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ttl256/euivator/internal/registry"
)

func TestTrie(t *testing.T) {
	cases := []struct {
		input []registry.Record
		want  []registry.Record
	}{
		{
			input: []registry.Record{},
			want:  []registry.Record(nil),
		},
		{
			input: []registry.Record{
				{Assignment: "F", Registry: registry.NameMAL, OrgName: "A", OrgAddress: "A"},
				{Assignment: "F", Registry: registry.NameMAL, OrgName: "B", OrgAddress: "B"},
				{Assignment: "FA", Registry: registry.NameMAL, OrgName: "C", OrgAddress: "C"},
				{Assignment: "FB", Registry: registry.NameMAL, OrgName: "D", OrgAddress: "D"},
				{Assignment: "FAB", Registry: registry.NameMAL, OrgName: "F", OrgAddress: "F"},
			},
			want: []registry.Record{
				{Assignment: "F", Registry: registry.NameMAL, OrgName: "A", OrgAddress: "A"},
				{Assignment: "F", Registry: registry.NameMAL, OrgName: "B", OrgAddress: "B"},
				{Assignment: "FA", Registry: registry.NameMAL, OrgName: "C", OrgAddress: "C"},
				{Assignment: "FB", Registry: registry.NameMAL, OrgName: "D", OrgAddress: "D"},
				{Assignment: "FAB", Registry: registry.NameMAL, OrgName: "F", OrgAddress: "F"},
			},
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			trie := registry.NewTrie()
			trie.InsertMany(tt.input)
			got := trie.Traverse()

			sort.Sort(registry.RecordS(got))
			sort.Sort(registry.RecordS(tt.want))

			assert.Equal(t, tt.want, got)
		})
	}
}
