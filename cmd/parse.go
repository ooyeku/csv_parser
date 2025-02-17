package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ooyeku/csv_parser/pkg"
	"github.com/spf13/cobra"
)

var (
	delimiter string
	quote     string
	trim      bool
)

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Parse and display CSV file contents",
	Long: `Parse and display the contents of a CSV file with customizable options for
delimiter, quote character, and whitespace trimming.

Example:
  csv_parser parse data.csv
  csv_parser parse --delimiter=";" --quote="'" data.csv`,
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

		// Create config
		cfg := pkg.Config{
			Delimiter:   []rune(delimiter)[0],
			Quote:       []rune(quote)[0],
			TrimLeading: trim,
		}

		// Create reader
		reader, err := pkg.NewReader(file, cfg)
		if err != nil {
			return fmt.Errorf("error creating reader: %w", err)
		}

		// Read and display records
		for {
			record, err := reader.ReadRecord()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("error reading record: %w", err)
			}

			// Print the record
			for i, field := range record {
				if i > 0 {
					fmt.Print("\t")
				}
				fmt.Print(field)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	// Add flags
	parseCmd.Flags().StringVarP(&delimiter, "delimiter", "d", ",", "Field delimiter character")
	parseCmd.Flags().StringVarP(&quote, "quote", "q", "\"", "Quote character")
	parseCmd.Flags().BoolVarP(&trim, "trim", "t", false, "Trim leading whitespace in unquoted fields")
}
