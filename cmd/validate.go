package cmd

import (
	"fmt"
	"os"
	"strconv"

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
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Printf("Error closing file: %v\n", err)
			}
		}(file)

		// Create reader with default config
		cfg := pkg.DefaultConfig()
		table, err := pkg.ReadTable(file, cfg)
		if err != nil {
			return fmt.Errorf("error reading table: %w", err)
		}

		var errors []string

		// Validate column types
		for _, header := range table.Headers {
			colType, _ := table.GetColumnType(header)
			col, _ := table.GetColumn(header)

			for i, val := range col {
				switch colType {
				case pkg.TypeInteger:
					if _, err := strconv.ParseInt(val, 10, 64); err != nil && val != "" {
						errors = append(errors, fmt.Sprintf("Row %d, Column %s: Invalid integer value %q",
							i+1, header, val))
					}
				case pkg.TypeFloat:
					if _, err := strconv.ParseFloat(val, 64); err != nil && val != "" {
						errors = append(errors, fmt.Sprintf("Row %d, Column %s: Invalid float value %q",
							i+1, header, val))
					}
				case pkg.TypeBoolean:
					if val != "" && val != "true" && val != "false" {
						errors = append(errors, fmt.Sprintf("Row %d, Column %s: Invalid boolean value %q",
							i+1, header, val))
					}
				default:
					panic("unhandled default case")
				}

				// In strict mode, check for empty fields
				if strict && val == "" {
					errors = append(errors, fmt.Sprintf("Row %d, Column %s: Empty field not allowed in strict mode",
						i+1, header))
				}
			}
		}

		// Display results
		fmt.Printf("File: %s\n", filePath)
		fmt.Printf("Rows processed: %d\n", len(table.Rows))
		fmt.Printf("Columns: %d\n", len(table.Headers))

		if len(errors) > 0 {
			fmt.Println("\nValidation Errors:")
			for _, err := range errors {
				fmt.Printf("- %s\n", err)
			}
			return fmt.Errorf("validation failed with %d errors", len(errors))
		}

		fmt.Println("\nValidation successful! No errors found.")

		// Show column type summary
		fmt.Println("\nColumn Type Summary:")
		for _, header := range table.Headers {
			colType, _ := table.GetColumnType(header)
			fmt.Printf("- %s: %v\n", header, colType)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&strict, "strict", "s", false,
		"Enable strict validation (no empty fields allowed)")
}
