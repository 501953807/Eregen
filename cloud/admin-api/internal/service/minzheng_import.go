package service

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// TemplateAdapter defines the interface for parsing government data templates.
type TemplateAdapter interface {
	// ParseCSV parses CSV data into rows of key-value maps.
	ParseCSV(data []byte) ([]map[string]string, error)
	// ValidateRow checks if a row has all required fields.
	ValidateRow(row map[string]string) error
	// ExtractFields converts a CSV row into a CommunityElder model.
	ExtractFields(row map[string]string) model.CommunityElder
}

// DefaultAdapter is the self-defined template used by default.
type DefaultAdapter struct{}

var _ TemplateAdapter = (*DefaultAdapter)(nil)

func (a *DefaultAdapter) ParseCSV(data []byte) ([]map[string]string, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("csv has no data rows")
	}
	headers := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		if len(record) != len(headers) {
			continue // skip malformed row
		}
		row := make(map[string]string)
		for i, h := range headers {
			row[strings.TrimSpace(h)] = strings.TrimSpace(record[i])
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (a *DefaultAdapter) ValidateRow(row map[string]string) error {
	required := []string{"name", "id_card"}
	for _, field := range required {
		if row[field] == "" {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

func (a *DefaultAdapter) ExtractFields(row map[string]string) model.CommunityElder {
	var gender int
	switch strings.TrimSpace(row["gender"]) {
	case "1", "男":
		gender = 1
	case "2", "女":
		gender = 2
	default:
		gender = 0
	}
	age := 0
	if v := row["age"]; v != "" {
		fmt.Sscanf(v, "%d", &age)
	}
	return model.CommunityElder{
		Name:             row["name"],
		IDCard:           row["id_card"],
		Gender:           gender,
		Age:              age,
		Address:          row["address"],
		EmergencyContact: row["emergency_contact"],
		BankAccount:      row["bank_account"],
		Status:           "active",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

// CivilAffairsAdapter is a placeholder for the official civil affairs bureau template.
type CivilAffairsAdapter struct{}

var _ TemplateAdapter = (*CivilAffairsAdapter)(nil)

func (a *CivilAffairsAdapter) ParseCSV(data []byte) ([]map[string]string, error) {
	return nil, fmt.Errorf("not yet implemented — use DefaultAdapter")
}

func (a *CivilAffairsAdapter) ValidateRow(row map[string]string) error {
	return nil
}

func (a *CivilAffairsAdapter) ExtractFields(row map[string]string) model.CommunityElder {
	return model.CommunityElder{}
}
