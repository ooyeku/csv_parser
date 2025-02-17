package cmd

import (
	"fmt"
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

		// Display information
		fmt.Printf("File: %s\n", filePath)
		fmt.Printf("Total Rows: %d\n", len(table.Rows))
		fmt.Printf("Total Columns: %d\n", len(table.Headers))

		fmt.Println("\nColumn Information:")
		for i, header := range table.Headers {
			colType, _ := table.GetColumnType(header)
			col, _ := table.GetColumn(header)

			// Get sample of unique values
			uniqueVals := make(map[string]struct{})
			for _, v := range col[:m(len(col), 5)] {
				uniqueVals[v] = struct{}{}
			}
			samples := make([]string, 0, len(uniqueVals))
			for v := range uniqueVals {
				samples = append(samples, v)
			}

			fmt.Printf("%d. %s\n", i+1, header)
			fmt.Printf("   Type: %v\n", colType)
			fmt.Printf("   Sample Values: %v\n", samples)
		}

		// Show preview of data
		fmt.Println("\nData Preview:")
		fmt.Println(previewTable(table))

		return nil
	},
}

func previewTable(t *pkg.Table) string {
	preview := pkg.NewTable(t.Headers)
	for i := 0; i < m(5, len(t.Rows)); i++ {
		err := preview.AddRow(t.Rows[i])
		if err != nil {
			return ""
		}
	}
	return preview.String()
}

func m(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
