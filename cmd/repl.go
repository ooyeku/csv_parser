package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ooyeku/csv_parser/pkg"
	"github.com/spf13/cobra"
)

var (
	currentTable *pkg.Table
	currentFile  string
)

// replCmd represents the REPL command
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive CSV parsing session",
	Long: `Start an interactive session for parsing and analyzing CSV files.
Available commands:
  load <file>              - Load a CSV file
  info                     - Show information about the current table
  preview [n]              - Show first n rows (default: 5)
  stats                    - Show column statistics
  summarize [cols]         - Show detailed statistics for columns
  correlate [cols]         - Show correlation matrix for numeric columns
  pivot <row> <col> <val> - Create pivot table with aggregation
  dates <col>             - Analyze dates in a column
  undo                    - Undo last operation
  redo                    - Redo last undone operation
  help                    - Show this help message
  exit                    - Exit the REPL`,
	Run: func(cmd *cobra.Command, args []string) {
		repl := pkg.NewREPL()
		repl.Start()
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
}

func startREPL(cmd *cobra.Command, args []string) {
	fmt.Println("Welcome to the CSV Parser REPL!")
	fmt.Println("Type 'help' for available commands or 'exit' to quit")

	scanner := bufio.NewScanner(os.Stdin)
	mainFormat := getDefaultFormat()

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		command := strings.ToLower(args[0])
		switch command {
		case "exit":
			fmt.Println("Goodbye!")
			return

		case "help":
			fmt.Println(cmd.Long)

		case "load":
			if len(args) < 2 {
				fmt.Println("Usage: load <file>")
				continue
			}
			filePath := args[1]
			if err := loadFile(filePath); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Loaded %d rows from %s\n", len(currentTable.Rows), filePath)

		case "info":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			showTableInfo()

		case "preview":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			n := 5
			if len(args) > 1 {
				if n_, err := strconv.Atoi(args[1]); err == nil {
					n = n_
				}
			}
			showPreview(n, mainFormat)

		case "stats":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			showTableStats(mainFormat)

		case "filter":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			if len(args) < 4 {
				fmt.Println("Usage: filter <column> <operator> <value>")
				continue
			}
			if filtered, err := filterTable(args[1], args[2], args[3]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				currentTable = filtered
				fmt.Printf("Filtered to %d rows\n", len(currentTable.Rows))
				showPreview(5, mainFormat)
			}

		case "sort":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			if len(args) < 2 {
				fmt.Println("Usage: sort <column> [desc]")
				continue
			}
			desc := len(args) > 2 && strings.ToLower(args[2]) == "desc"
			if err := sortTable(args[1], desc); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Table sorted")
				showPreview(5, mainFormat)
			}

		case "group":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			if len(args) < 3 {
				fmt.Println("Usage: group <column> <aggregation>")
				continue
			}
			if grouped, err := groupTable(args[1], args[2]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(grouped.Format(getStatsFormat()))
			}

		case "format":
			if len(args) < 2 {
				fmt.Println("Usage: format <style> (default,fancy,rounded)")
				continue
			}
			mainFormat = getFormatByStyle(args[1])
			if currentTable != nil {
				showPreview(5, mainFormat)
			}

		case "export":
			if currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			if len(args) < 2 {
				fmt.Println("Usage: export <file>")
				continue
			}
			if err := exportTable(args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Table exported to %s\n", args[1])
			}

		default:
			fmt.Printf("Unknown command: %s\nType 'help' for available commands\n", command)
		}
	}
}

func loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	table, err := pkg.ReadTable(file, pkg.DefaultConfig())
	if err != nil {
		return fmt.Errorf("error reading table: %w", err)
	}

	currentTable = table
	currentFile = path
	return nil
}

func showTableInfo() {
	fmt.Printf("File: %s\n", currentFile)
	fmt.Printf("Rows: %d\n", len(currentTable.Rows))
	fmt.Printf("Columns: %d\n\n", len(currentTable.Headers))

	fmt.Println("Column Information:")
	for i, header := range currentTable.Headers {
		colType, _ := currentTable.GetColumnType(header)
		fmt.Printf("%d. %s (%v)\n", i+1, header, colType)
	}
}

func showPreview(n int, format pkg.FormatOptions) {
	preview := pkg.NewTable(currentTable.Headers)
	for i := 0; i < minInt(n, len(currentTable.Rows)); i++ {
		preview.AddRow(currentTable.Rows[i])
	}
	fmt.Println(preview.Format(format))
}

func showTableStats(format pkg.FormatOptions) {
	stats := pkg.NewTable([]string{"Column", "Type", "Unique Values", "Null Count"})

	for _, header := range currentTable.Headers {
		col, _ := currentTable.GetColumn(header)
		colType, _ := currentTable.GetColumnType(header)

		// Count unique values and nulls
		unique := make(map[string]struct{})
		nullCount := 0
		for _, val := range col {
			if val == "" {
				nullCount++
			} else {
				unique[val] = struct{}{}
			}
		}

		stats.AddRow([]string{
			header,
			fmt.Sprintf("%v", colType),
			strconv.Itoa(len(unique)),
			strconv.Itoa(nullCount),
		})
	}

	fmt.Println(stats.Format(getStatsFormat()))
}

func filterTable(column, op, value string) (*pkg.Table, error) {
	filtered := pkg.NewTable(currentTable.Headers)
	colIdx := -1
	for i, h := range currentTable.Headers {
		if h == column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return nil, fmt.Errorf("column %s not found", column)
	}

	for _, row := range currentTable.Rows {
		if matchesFilter(row[colIdx], op, value) {
			filtered.AddRow(row)
		}
	}
	return filtered, nil
}

func matchesFilter(val, op, target string) bool {
	switch op {
	case "=", "==":
		return val == target
	case "!=":
		return val != target
	case ">", "<", ">=", "<=":
		v1, err1 := strconv.ParseFloat(val, 64)
		v2, err2 := strconv.ParseFloat(target, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		switch op {
		case ">":
			return v1 > v2
		case "<":
			return v1 < v2
		case ">=":
			return v1 >= v2
		case "<=":
			return v1 <= v2
		}
	}
	return false
}

func sortTable(column string, desc bool) error {
	colIdx := -1
	for i, h := range currentTable.Headers {
		if h == column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return fmt.Errorf("column %s not found", column)
	}

	sort.Slice(currentTable.Rows, func(i, j int) bool {
		if desc {
			i, j = j, i
		}
		return currentTable.Rows[i][colIdx] < currentTable.Rows[j][colIdx]
	})
	return nil
}

func groupTable(column, agg string) (*pkg.Table, error) {
	return currentTable.GroupBy(
		[]string{column},
		map[string]string{"*": agg},
	)
}

func exportTable(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write headers
	fmt.Fprintln(file, strings.Join(currentTable.Headers, ","))

	// Write rows
	for _, row := range currentTable.Rows {
		fmt.Fprintln(file, strings.Join(row, ","))
	}
	return nil
}

func getDefaultFormat() pkg.FormatOptions {
	return pkg.FormatOptions{
		Style:          pkg.RoundedStyle,
		HeaderStyle:    pkg.Bold,
		HeaderColor:    pkg.Cyan,
		BorderColor:    pkg.Blue,
		AlternateRows:  true,
		AlternateColor: pkg.Dim,
		NumberedRows:   true,
		MaxColumnWidth: 20,
		WrapText:       true,
	}
}

func getStatsFormat() pkg.FormatOptions {
	return pkg.FormatOptions{
		Style:          pkg.FancyStyle,
		HeaderStyle:    pkg.Bold + pkg.Underline,
		HeaderColor:    pkg.Yellow,
		BorderColor:    pkg.Green,
		AlternateRows:  true,
		AlternateColor: pkg.Dim,
		Alignment:      []string{"left", "left", "right", "right"},
	}
}

func getFormatByStyle(style string) pkg.FormatOptions {
	format := getDefaultFormat()
	switch strings.ToLower(style) {
	case "fancy":
		format.Style = pkg.FancyStyle
	case "default":
		format.Style = pkg.DefaultStyle
	case "rounded":
		format.Style = pkg.RoundedStyle
	}
	return format
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
