package benchmark

import (
	"fmt"
	"os"
	"strings"
)

// BenchData represents a benchmark dataset
type BenchData struct {
	Name     string
	Content  string
	FileSize int64
}

// GenerateBenchmarkData creates benchmark datasets of various sizes and complexities
func GenerateBenchmarkData() []BenchData {
	return []BenchData{
		generateSimpleCSV(1000),      // 1K rows
		generateSimpleCSV(100000),    // 100K rows
		generateSimpleCSV(1000000),   // 1M rows
		generateQuotedCSV(1000),      // 1K rows with quotes
		generateQuotedCSV(100000),    // 100K rows with quotes
		generateComplexCSV(1000),     // 1K rows with mixed content
		generateComplexCSV(100000),   // 100K rows with mixed content
		generateWideCSV(1000, 100),   // 1K rows x 100 columns
		generateWideCSV(100000, 100), // 100K rows x 100 columns
	}
}

// SaveBenchmarkData saves benchmark data to files in the specified directory
func SaveBenchmarkData(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create benchmark directory: %w", err)
	}

	for _, data := range GenerateBenchmarkData() {
		filename := fmt.Sprintf("%s/bench_%s.csv", dir, strings.ReplaceAll(data.Name, " ", "_"))
		if err := os.WriteFile(filename, []byte(data.Content), 0644); err != nil {
			return fmt.Errorf("failed to write benchmark file %s: %w", filename, err)
		}
	}

	return nil
}

// generateSimpleCSV generates a simple CSV with numeric data
func generateSimpleCSV(rows int) BenchData {
	var sb strings.Builder
	sb.WriteString("id,value1,value2,value3,value4,value5\n")

	for i := 0; i < rows; i++ {
		sb.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d\n",
			i, i*2, i*3, i*4, i*5, i*6))
	}

	content := sb.String()
	return BenchData{
		Name:     fmt.Sprintf("simple_%dk", rows/1000),
		Content:  content,
		FileSize: int64(len(content)),
	}
}

// generateQuotedCSV generates a CSV with quoted fields containing commas
func generateQuotedCSV(rows int) BenchData {
	var sb strings.Builder
	sb.WriteString("id,description,data,notes\n")

	for i := 0; i < rows; i++ {
		sb.WriteString(fmt.Sprintf("%d,\"Description, with comma\",\"Data, with, multiple, commas\",\"Note %d\"\n",
			i, i))
	}

	content := sb.String()
	return BenchData{
		Name:     fmt.Sprintf("quoted_%dk", rows/1000),
		Content:  content,
		FileSize: int64(len(content)),
	}
}

// generateComplexCSV generates a CSV with mixed content types and special cases
func generateComplexCSV(rows int) BenchData {
	var sb strings.Builder
	sb.WriteString("id,text,quoted,null,comment,empty\n")

	for i := 0; i < rows; i++ {
		// Mix of normal text, quoted text with commas, NULL values, and empty fields
		sb.WriteString(fmt.Sprintf("%d,normal text,\"quoted, with \"\"escaped\"\" quotes\",\\N,#comment,\n",
			i))
	}

	content := sb.String()
	return BenchData{
		Name:     fmt.Sprintf("complex_%dk", rows/1000),
		Content:  content,
		FileSize: int64(len(content)),
	}
}

// generateWideCSV generates a CSV with many columns
func generateWideCSV(rows, cols int) BenchData {
	var sb strings.Builder

	// Generate header
	for i := 0; i < cols; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("col%d", i))
	}
	sb.WriteString("\n")

	// Generate rows
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if j > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("value_%d_%d", i, j))
		}
		sb.WriteString("\n")
	}

	content := sb.String()
	return BenchData{
		Name:     fmt.Sprintf("wide_%dk_%dcols", rows/1000, cols),
		Content:  content,
		FileSize: int64(len(content)),
	}
}
