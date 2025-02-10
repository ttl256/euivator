//go:generate go-enum --names --values

package registry

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ENUM(MA-L, MA-M, MA-S, CID).
type Name string

type Record struct {
	Assignment string `json:"assignment"`
	Registry   Name   `json:"registry"`
	OrgName    string `json:"org_name"`
	OrgAddress string `json:"org_address"`
}

func ParseRecordsFromCSV(r io.Reader) ([]Record, error) {
	var records []Record

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 4
	reader.LazyQuotes = true

	// Skip the headers line
	_, err := reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return records, nil
		}
		return nil, fmt.Errorf("error reading CSV data: %w", err)
	}

	var (
		line   []string
		record Record
	)

	for {
		line, err = reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("error reading CSV data: %w", err)
		}

		record, err = ParseLine(line)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil
}

func ParseLine(lineFields []string) (Record, error) {
	line := make([]string, len(lineFields))
	copy(line, lineFields)

	for i := 0; i < len(line); i++ {
		line[i] = strings.TrimSpace(line[i])
	}

	registryName, err := ParseName(line[0])
	if err != nil {
		return Record{}, err
	}

	return Record{Registry: registryName, Assignment: line[1], OrgName: line[2], OrgAddress: line[3]}, nil
}

type RecordS []Record

func (r RecordS) Len() int {
	return len(r)
}

func (r RecordS) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RecordS) Less(i, j int) bool {
	if cmp := strings.Compare(r[i].Assignment, r[j].Assignment); cmp != 0 {
		return cmp < 0
	}

	if r[i].OrgName != r[j].OrgName {
		return r[i].OrgName < r[j].OrgName
	}

	return r[i].OrgAddress < r[j].OrgAddress
}
