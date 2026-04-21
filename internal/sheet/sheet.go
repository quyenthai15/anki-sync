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

func (f *Fetcher) GetHeaders() ([]string, error) {
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

	return header, nil
}

func (f *Fetcher) GetFirstColumnName() (string, error) {
	headers, err := f.GetHeaders()
	if err != nil {
		return "", err
	}
	if len(headers) == 0 {
		return "", fmt.Errorf("sheet has no columns")
	}
	return headers[0], nil
}

func (f *Fetcher) GetRecords() ([]Record, error) {
	var records []Record
	err := f.ForEachRecord(func(r Record) error {
		records = append(records, r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (f *Fetcher) ForEachRecord(callback func(Record) error) error {
	resp, err := http.Get(f.url)
	if err != nil {
		return fmt.Errorf("failed to fetch sheet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read csv row: %w", err)
		}
		record := make(Record)
		for i, value := range row {
			if i < len(header) {
				record[header[i]] = value
			}
		}
		if err := callback(record); err != nil {
			return err
		}
	}

	return nil
}
