package web

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadPlaceIDsFromJobCSV reads place_id values from a completed job CSV.
func (s *Service) LoadPlaceIDsFromJobCSV(jobID string) (map[string]struct{}, error) {
	if strings.Contains(jobID, "/") || strings.Contains(jobID, "\\") || strings.Contains(jobID, "..") {
		return nil, fmt.Errorf("id de job inválido")
	}

	path := filepath.Join(s.dataFolder, jobID+".csv")

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("CSV do job de exclusão não encontrado")
		}

		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler cabeçalho do CSV: %w", err)
	}

	placeIdx := -1
	for i, h := range headers {
		if strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")) == "place_id" {
			placeIdx = i
			break
		}
	}

	if placeIdx < 0 {
		return nil, fmt.Errorf("coluna place_id não encontrada no CSV")
	}

	out := make(map[string]struct{})

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		if placeIdx >= len(row) {
			continue
		}

		id := strings.TrimSpace(row[placeIdx])
		if id == "" {
			continue
		}

		out[id] = struct{}{}
	}

	return out, nil
}
