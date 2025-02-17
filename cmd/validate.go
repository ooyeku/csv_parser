package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ooyeku/csv_parser/pkg"
	"github.com/spf13/cobra"
)

var strict bool

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate CSV file structure",
	Long: `Validate the structure of a CSV file by checking:
- Consistent number of columns across all rows
- Proper quote and delimiter usage
- No malformed rows

Example:
  csv_parser validate data.csv
  csv_parser validate --strict data.csv`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()

		// Create reader with default config
		cfg := pkg.DefaultConfig()
		reader, err := pkg.NewReader(file, cfg)
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}

		var (
			rowCount    int
			columnCount int
			errors      []string
		)

		// Validate records
		for {
			record, err := reader.ReadRecord()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("error reading record: %w", err)
			}

			rowCount++

			// Check column consistency
			if rowCount == 1 {
				columnCount = len(record)
			} else if len(record) != columnCount {
				errors = append(errors, fmt.Sprintf("Row %d: Expected %d columns, got %d",
					rowCount, columnCount, len(record)))
				if !strict {
					continue
				}
				break
			}

			// In strict mode, check for empty fields
			if strict {
				for i, field := range record {
					if field == "" {
						errors = append(errors, fmt.Sprintf("Row %d, Column %d: Empty field",
							rowCount, i+1))
					}
				}
			}
		}

		// Display results
		fmt.Printf("File: %s\n", filePath)
		fmt.Printf("Rows processed: %d\n", rowCount)
		fmt.Printf("Columns per row: %d\n", columnCount)

		if len(errors) > 0 {
			fmt.Println("\nValidation Errors:")
			for _, err := range errors {
				fmt.Printf("- %s\n", err)
			}
			return fmt.Errorf("validation failed with %d errors", len(errors))
		}

		fmt.Println("\nValidation successful! No errors found.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&strict, "strict", "s", false,
		"Enable strict validation (no empty fields allowed)")
}
