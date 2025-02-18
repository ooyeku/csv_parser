package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ooyeku/csv_parser/pkg"
	"github.com/spf13/cobra"
)

var (
	format string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [input.csv] [output.json|html]",
	Short: "Export CSV data to different formats",
	Long: `Export CSV data to different formats (JSON, HTML).
Automatically detects output format from file extension.

Example:
  csv_parser export data.csv output.json
  csv_parser export data.csv output.html
  csv_parser export --format=json data.csv output.txt`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		outputFile := args[1]

		// Determine format
		exportFormat := format
		if exportFormat == "" {
			ext := strings.ToLower(filepath.Ext(outputFile))
			switch ext {
			case ".json":
				exportFormat = "json"
			case ".html":
				exportFormat = "html"
			default:
				return fmt.Errorf("unknown output format: %s", ext)
			}
		}

		// Read input CSV
		input, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer input.Close()

		// Parse CSV
		table, err := pkg.ReadTable(input, pkg.DefaultConfig())
		if err != nil {
			return fmt.Errorf("error reading CSV: %w", err)
		}

		// Create output file
		output, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer output.Close()

		// Export based on format
		switch exportFormat {
		case "json":
			if err := table.ExportToJSON(output); err != nil {
				return fmt.Errorf("error exporting to JSON: %w", err)
			}
		case "html":
			if err := table.ExportToHTML(output); err != nil {
				return fmt.Errorf("error exporting to HTML: %w", err)
			}
		default:
			return fmt.Errorf("unsupported format: %s", exportFormat)
		}

		fmt.Printf("Successfully exported to %s\n", outputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&format, "format", "f", "", "Export format (json, html)")
}
