package probe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveReport saves the probe report to a JSONL file
func SaveReport(report *ProbeReport, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(outputDir, fmt.Sprintf("probe_report_%s.jsonl", timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, result := range report.Results {
		if err := encoder.Encode(result); err != nil {
			return "", fmt.Errorf("failed to encode result: %w", err)
		}
	}

	return filename, nil
}

// LoadReport loads a probe report from a JSONL file
func LoadReport(filename string) (*ProbeReport, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open report file: %w", err)
	}
	defer file.Close()

	report := &ProbeReport{
		Results: make([]ProbeResult, 0),
	}

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var result ProbeResult
		if err := decoder.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
		report.Results = append(report.Results, result)

		// Update statistics
		report.Total++
		if result.StatusCode == 0 {
			report.NetFail++
		} else if result.StatusCode == 200 {
			report.Success++
		} else {
			report.Invalid++
		}
	}

	return report, nil
}
