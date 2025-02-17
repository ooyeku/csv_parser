package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ooyeku/csv_parser/pkg"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Short: "Display information about a CSV file",
	Long: `Display basic information about a CSV file including:
- Number of rows
- Number of columns
- Sample of first few rows
- Detected delimiter (if different from default)

Example:
  csv_parser info data.csv`,
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
			firstRow    []string
		)

		// Read records to gather information
		for i := 0; ; i++ {
			record, err := reader.ReadRecord()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("error reading record: %w", err)
			}

			rowCount++

			// Store column count from first row
			if i == 0 {
				columnCount = len(record)
				firstRow = record
			}

			// Only process first few rows for performance
			if i >= 5 {
				// Count remaining rows
				for {
					_, err := reader.ReadRecord()
					if err != nil {
						if err == io.EOF {
							break
						}
						return fmt.Errorf("error reading record: %w", err)
					}
					rowCount++
				}
				break
			}
		}

		// Display information
		fmt.Printf("File: %s\n", filePath)
		fmt.Printf("Total Rows: %d\n", rowCount)
		fmt.Printf("Columns: %d\n", columnCount)

		if len(firstRow) > 0 {
			fmt.Println("\nColumn Headers:")
			for i, header := range firstRow {
				fmt.Printf("%d. %s\n", i+1, header)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
