package sheet

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
)

type Record map[string]string

type Fetcher struct {
	url string
}

func NewFetcher(url string) *Fetcher {
	return &Fetcher{url: url}
}

func (f *Fetcher) GetRecords() ([]Record, error) {
	resp, err := http.Get(f.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sheet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var records []Record

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read csv row: %w", err)
		}
		record := make(Record)
		for i, value := range row {
			if i < len(header) {
				record[header[i]] = value
			}
		}
		records = append(records, record)
	}

	return records, nil
}
