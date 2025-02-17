package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ooyeku/csv_parser/pkg"
	"github.com/ooyeku/csv_parser/pkg/benchmark"
	"github.com/spf13/cobra"
)

var (
	benchDir string
	generate bool
)

// benchCmd represents the bench command
var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run benchmarks on CSV parser",
	Long: `Run comprehensive benchmarks on the CSV parser using various datasets.
Includes tests for different file sizes, content types, and configurations.

Example:
  csv_parser bench
  csv_parser bench --generate  # Generate new benchmark data
  csv_parser bench --dir=/path/to/data  # Use custom benchmark data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if generate {
			fmt.Println("Generating benchmark data...")
			if err := benchmark.SaveBenchmarkData(benchDir); err != nil {
				return fmt.Errorf("failed to generate benchmark data: %w", err)
			}
		}

		// Ensure benchmark directory exists
		if _, err := os.Stat(benchDir); os.IsNotExist(err) {
			return fmt.Errorf("benchmark directory %s does not exist. Use --generate to create benchmark data", benchDir)
		}

		// Run benchmarks on each file
		files, err := filepath.Glob(filepath.Join(benchDir, "bench_*.csv"))
		if err != nil {
			return fmt.Errorf("failed to list benchmark files: %w", err)
		}

		fmt.Printf("\nRunning benchmarks...\n\n")
		for _, file := range files {
			if err := benchmarkFile(file); err != nil {
				fmt.Printf("Error benchmarking %s: %v\n", file, err)
				continue
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(benchCmd)

	benchCmd.Flags().StringVarP(&benchDir, "dir", "d", "testdata/bench", "Directory containing benchmark data")
	benchCmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate new benchmark data")
}

func benchmarkFile(file string) error {
	start := time.Now()

	// Open file
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	fileInfo, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create reader with default config
	reader, err := pkg.NewReader(f, pkg.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}

	var rowCount int
	for {
		_, err := reader.ReadRecord()
		if err != nil {
			break
		}
		rowCount++
	}

	duration := time.Since(start)
	bytesPerSecond := float64(fileInfo.Size()) / duration.Seconds()
	rowsPerSecond := float64(rowCount) / duration.Seconds()

	fmt.Printf("File: %s\n", filepath.Base(file))
	fmt.Printf("  Size: %.2f MB\n", float64(fileInfo.Size())/1024/1024)
	fmt.Printf("  Rows: %d\n", rowCount)
	fmt.Printf("  Time: %v\n", duration)
	fmt.Printf("  Speed: %.2f MB/s\n", bytesPerSecond/1024/1024)
	fmt.Printf("  Rows/s: %.0f\n\n", rowsPerSecond)

	return nil
}
