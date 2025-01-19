//go:generate go-enum --names --values

package registry

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
)

// ENUM(MA-L, MA-M, MA-S, CID).
type Name string

type Record struct {
	Assignment []byte
	Registry   Name
	OrgName    string
	OrgAddress string
}

func ParseRecordsFromCSV(r io.Reader) ([]Record, error) {
	var records []Record

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 4
	reader.LazyQuotes = true

	var n int

	for {
		line, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("error reading CSV data: %w", err)
		}
		n++

		// Skip the headers line
		if n < 2 { //nolint: mnd // ok
			continue
		}
		registryName, err := ParseName(line[0])
		if err != nil {
			return nil, err
		}

		record := Record{Registry: registryName, Assignment: []byte(line[1]), OrgName: line[2], OrgAddress: line[3]}
		records = append(records, record)
	}

	return records, nil
}

type RecordS []Record

func (r RecordS) Len() int {
	return len(r)
}

func (r RecordS) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RecordS) Less(i, j int) bool {
	if cmp := bytes.Compare(r[i].Assignment, r[j].Assignment); cmp != 0 {
		return cmp < 0
	}

	if r[i].OrgName != r[j].OrgName {
		return r[i].OrgName < r[j].OrgName
	}

	return r[i].OrgAddress < r[j].OrgAddress
}
