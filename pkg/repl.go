package pkg

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// REPL represents the interactive CSV analysis environment
type REPL struct {
	currentTable *Table
	currentFile  string
	undoStack    []*Table
	redoStack    []*Table
	formats      map[string]FormatOptions
	history      []string
	lastResult   *Table
}

// NewREPL creates a new REPL instance
func NewREPL() *REPL {
	r := &REPL{
		undoStack: make([]*Table, 0),
		redoStack: make([]*Table, 0),
		formats:   make(map[string]FormatOptions),
		history:   make([]string, 0),
	}

	// Initialize default formats
	r.formats["default"] = FormatOptions{
		Style:          DefaultStyle,
		HeaderStyle:    Bold,
		HeaderColor:    Cyan,
		BorderColor:    Blue,
		AlternateRows:  true,
		AlternateColor: Dim,
		NumberedRows:   true,
		MaxColumnWidth: 20,
		WrapText:       true,
	}

	r.formats["fancy"] = FormatOptions{
		Style:          FancyStyle,
		HeaderStyle:    Bold + Underline,
		HeaderColor:    Yellow,
		BorderColor:    Green,
		AlternateRows:  true,
		AlternateColor: Dim,
		NumberedRows:   true,
		MaxColumnWidth: 30,
		WrapText:       true,
	}

	r.formats["compact"] = FormatOptions{
		Style:          RoundedStyle,
		HeaderStyle:    Bold,
		HeaderColor:    White,
		BorderColor:    Blue,
		CompactBorders: true,
		MaxColumnWidth: 15,
	}

	r.formats["stats"] = FormatOptions{
		Style:          FancyStyle,
		HeaderStyle:    Bold + Underline,
		HeaderColor:    Magenta,
		BorderColor:    Cyan,
		AlternateRows:  true,
		AlternateColor: Dim,
		Alignment:      []string{"left", "right", "right", "right", "right"},
	}

	return r
}

// Table Operations

func (r *REPL) pushUndo() {
	if r.currentTable == nil {
		return
	}
	tableCopy := r.currentTable.Copy()
	r.undoStack = append(r.undoStack, tableCopy)
	r.redoStack = nil // Clear redo stack on new operation
}

func (r *REPL) Undo() error {
	if len(r.undoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	r.redoStack = append(r.redoStack, r.currentTable.Copy())
	r.currentTable = r.undoStack[len(r.undoStack)-1]
	r.undoStack = r.undoStack[:len(r.undoStack)-1]
	return nil
}

func (r *REPL) Redo() error {
	if len(r.redoStack) == 0 {
		return fmt.Errorf("nothing to redo")
	}
	r.undoStack = append(r.undoStack, r.currentTable.Copy())
	r.currentTable = r.redoStack[len(r.redoStack)-1]
	r.redoStack = r.redoStack[:len(r.redoStack)-1]
	return nil
}

// Advanced Analysis

func (r *REPL) Summarize(columns []string) (*Table, error) {
	if r.currentTable == nil {
		return nil, fmt.Errorf("no table loaded")
	}

	if len(columns) == 0 {
		columns = r.currentTable.Headers
	}

	summary := NewTable([]string{"Column", "Type", "Min", "Max", "Mean", "Median", "StdDev", "Unique"})

	for _, col := range columns {
		values, err := r.currentTable.GetColumn(col)
		if err != nil {
			continue
		}

		colType, _ := r.currentTable.GetColumnType(col)
		stats := r.calculateStats(values, colType)
		summary.AddRow([]string{
			col,
			fmt.Sprintf("%v", colType),
			stats["min"],
			stats["max"],
			stats["mean"],
			stats["median"],
			stats["stddev"],
			stats["unique"],
		})
	}

	return summary, nil
}

func (r *REPL) Correlate(columns []string) (*Table, error) {
	if r.currentTable == nil {
		return nil, fmt.Errorf("no table loaded")
	}

	if len(columns) == 0 {
		// Only use numeric columns
		for _, h := range r.currentTable.Headers {
			colType, _ := r.currentTable.GetColumnType(h)
			if colType == TypeInteger || colType == TypeFloat {
				columns = append(columns, h)
			}
		}
	}

	corr := NewTable(append([]string{"Column"}, columns...))

	for _, row := range columns {
		rowVals, _ := r.currentTable.GetColumn(row)
		rowNums := toNumbers(rowVals)

		corrRow := []string{row}
		for _, col := range columns {
			colVals, _ := r.currentTable.GetColumn(col)
			colNums := toNumbers(colVals)

			correlation := calculateCorrelation(rowNums, colNums)
			corrRow = append(corrRow, fmt.Sprintf("%.3f", correlation))
		}
		corr.AddRow(corrRow)
	}

	return corr, nil
}

func (r *REPL) Pivot(rows, cols, values string, agg string) (*Table, error) {
	if r.currentTable == nil {
		return nil, fmt.Errorf("no table loaded")
	}

	// Get unique values for column headers
	colVals, err := r.currentTable.GetColumn(cols)
	if err != nil {
		return nil, err
	}
	uniqueCols := uniqueStrings(colVals)
	sort.Strings(uniqueCols)

	// Create headers for pivot table
	headers := []string{rows}
	headers = append(headers, uniqueCols...)

	pivot := NewTable(headers)

	// Get unique values for row dimension
	rowVals, err := r.currentTable.GetColumn(rows)
	if err != nil {
		return nil, err
	}
	uniqueRows := uniqueStrings(rowVals)
	sort.Strings(uniqueRows)

	// Build pivot table
	for _, rowVal := range uniqueRows {
		row := []string{rowVal}
		for _, colVal := range uniqueCols {
			filtered := r.currentTable.Filter(func(row []string) bool {
				rowIdx := r.currentTable.index[rows]
				colIdx := r.currentTable.index[cols]
				return row[rowIdx] == rowVal && row[colIdx] == colVal
			})

			// Get aggregated value
			valIdx := r.currentTable.index[values]
			vals := make([]string, 0)
			for _, r := range filtered.Rows {
				vals = append(vals, r[valIdx])
			}
			aggVal := aggregateValues(vals, agg)
			row = append(row, aggVal)
		}
		pivot.AddRow(row)
	}

	return pivot, nil
}

func (r *REPL) DateAnalysis(dateCol string) (*Table, error) {
	if r.currentTable == nil {
		return nil, fmt.Errorf("no table loaded")
	}

	dates, err := r.currentTable.GetColumn(dateCol)
	if err != nil {
		return nil, err
	}

	analysis := NewTable([]string{
		"Metric", "Value",
		"Sample", "Distribution",
	})

	var validDates []time.Time
	var minDate, maxDate time.Time
	format := "2006-01-02" // Default format

	// Try to parse dates and find min/max
	for _, d := range dates {
		if t, err := time.Parse(format, d); err == nil {
			validDates = append(validDates, t)
			if minDate.IsZero() || t.Before(minDate) {
				minDate = t
			}
			if maxDate.IsZero() || t.After(maxDate) {
				maxDate = t
			}
		}
	}

	if len(validDates) == 0 {
		return nil, fmt.Errorf("no valid dates found in column %s", dateCol)
	}

	// Calculate date metrics
	duration := maxDate.Sub(minDate)
	days := duration.Hours() / 24
	years := days / 365.25

	// Count by year
	yearCount := make(map[int]int)
	for _, d := range validDates {
		yearCount[d.Year()]++
	}

	// Count by month
	monthCount := make(map[time.Month]int)
	for _, d := range validDates {
		monthCount[d.Month()]++
	}

	// Count by weekday
	weekdayCount := make(map[time.Weekday]int)
	for _, d := range validDates {
		weekdayCount[d.Weekday()]++
	}

	// Add metrics to analysis table
	analysis.AddRow([]string{
		"Date Range",
		fmt.Sprintf("%v to %v", minDate.Format(format), maxDate.Format(format)),
		fmt.Sprintf("%.1f years", years),
		createDistributionBar(len(validDates), len(dates)),
	})

	// Add year distribution
	years = float64(len(yearCount))
	analysis.AddRow([]string{
		"Years",
		fmt.Sprintf("%d unique", len(yearCount)),
		fmt.Sprintf("%.1f dates/year", float64(len(validDates))/years),
		createYearDistribution(yearCount),
	})

	// Add month distribution
	analysis.AddRow([]string{
		"Months",
		fmt.Sprintf("%d months", len(monthCount)),
		fmt.Sprintf("%.1f dates/month", float64(len(validDates))/12),
		createMonthDistribution(monthCount),
	})

	// Add weekday distribution
	analysis.AddRow([]string{
		"Weekdays",
		fmt.Sprintf("%d days", len(weekdayCount)),
		fmt.Sprintf("%.1f dates/day", float64(len(validDates))/7),
		createWeekdayDistribution(weekdayCount),
	})

	return analysis, nil
}

// Helper Functions

func (r *REPL) calculateStats(values []string, colType ColumnType) map[string]string {
	stats := make(map[string]string)

	if colType == TypeInteger || colType == TypeFloat {
		nums := toNumbers(values)
		if len(nums) == 0 {
			return map[string]string{
				"min": "N/A", "max": "N/A", "mean": "N/A",
				"median": "N/A", "stddev": "N/A", "unique": "0",
			}
		}

		sort.Float64s(nums)

		stats["min"] = fmt.Sprintf("%.2f", nums[0])
		stats["max"] = fmt.Sprintf("%.2f", nums[len(nums)-1])
		stats["mean"] = fmt.Sprintf("%.2f", mean(nums))
		stats["median"] = fmt.Sprintf("%.2f", median(nums))
		stats["stddev"] = fmt.Sprintf("%.2f", stdDev(nums))
		stats["unique"] = fmt.Sprintf("%d", len(uniqueFloat64s(nums)))
	} else {
		unique := uniqueStrings(values)
		stats["min"] = "N/A"
		stats["max"] = "N/A"
		stats["mean"] = "N/A"
		stats["median"] = "N/A"
		stats["stddev"] = "N/A"
		stats["unique"] = fmt.Sprintf("%d", len(unique))
	}

	return stats
}

func toNumbers(values []string) []float64 {
	var nums []float64
	for _, v := range values {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			nums = append(nums, n)
		}
	}
	return nums
}

func mean(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0.0
	for _, n := range nums {
		sum += n
	}
	return sum / float64(len(nums))
}

func median(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	mid := len(nums) / 2
	if len(nums)%2 == 0 {
		return (nums[mid-1] + nums[mid]) / 2
	}
	return nums[mid]
}

func stdDev(nums []float64) float64 {
	if len(nums) < 2 {
		return 0
	}
	m := mean(nums)
	var sum float64
	for _, n := range nums {
		sum += (n - m) * (n - m)
	}
	return math.Sqrt(sum / float64(len(nums)-1))
}

func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	mx, my := mean(x), mean(y)
	var num, dx, dy float64

	for i := range x {
		dx += (x[i] - mx) * (x[i] - mx)
		dy += (y[i] - my) * (y[i] - my)
		num += (x[i] - mx) * (y[i] - my)
	}

	if dx == 0 || dy == 0 {
		return 0
	}
	return num / math.Sqrt(dx*dy)
}

func uniqueStrings(values []string) []string {
	unique := make(map[string]struct{})
	for _, v := range values {
		unique[v] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for v := range unique {
		result = append(result, v)
	}
	return result
}

func uniqueFloat64s(values []float64) []float64 {
	unique := make(map[float64]struct{})
	for _, v := range values {
		unique[v] = struct{}{}
	}
	result := make([]float64, 0, len(unique))
	for v := range unique {
		result = append(result, v)
	}
	return result
}

func aggregateValues(values []string, agg string) string {
	nums := toNumbers(values)
	if len(nums) == 0 {
		return ""
	}

	switch strings.ToLower(agg) {
	case "sum":
		return fmt.Sprintf("%.2f", mean(nums)*float64(len(nums)))
	case "avg":
		return fmt.Sprintf("%.2f", mean(nums))
	case "min":
		sort.Float64s(nums)
		return fmt.Sprintf("%.2f", nums[0])
	case "max":
		sort.Float64s(nums)
		return fmt.Sprintf("%.2f", nums[len(nums)-1])
	case "count":
		return fmt.Sprintf("%d", len(nums))
	default:
		return ""
	}
}

func createDistributionBar(valid, total int) string {
	if total == 0 {
		return ""
	}
	percent := float64(valid) / float64(total) * 100
	bars := int(percent / 5)
	return fmt.Sprintf("%s %.1f%%", strings.Repeat("█", bars), percent)
}

func createYearDistribution(counts map[int]int) string {
	years := make([]int, 0, len(counts))
	for y := range counts {
		years = append(years, y)
	}
	sort.Ints(years)

	max := 0
	for _, count := range counts {
		if count > max {
			max = count
		}
	}

	var dist strings.Builder
	for _, year := range years {
		bars := int(float64(counts[year]) / float64(max) * 10)
		dist.WriteString(fmt.Sprintf("%d%s ", year%100, strings.Repeat("▇", bars)))
	}
	return dist.String()
}

func createMonthDistribution(counts map[time.Month]int) string {
	max := 0
	for _, count := range counts {
		if count > max {
			max = count
		}
	}

	var dist strings.Builder
	for m := time.January; m <= time.December; m++ {
		if count, ok := counts[m]; ok {
			bars := int(float64(count) / float64(max) * 5)
			dist.WriteString(fmt.Sprintf("%s%s ", m.String()[:3], strings.Repeat("▇", bars)))
		}
	}
	return dist.String()
}

func createWeekdayDistribution(counts map[time.Weekday]int) string {
	max := 0
	for _, count := range counts {
		if count > max {
			max = count
		}
	}

	var dist strings.Builder
	for d := time.Sunday; d <= time.Saturday; d++ {
		if count, ok := counts[d]; ok {
			bars := int(float64(count) / float64(max) * 5)
			dist.WriteString(fmt.Sprintf("%s%s ", d.String()[:3], strings.Repeat("▇", bars)))
		}
	}
	return dist.String()
}

// Start begins the REPL session
func (r *REPL) Start() {
	fmt.Println("Welcome to the CSV Parser REPL!")
	fmt.Println("Type 'help' for available commands or 'exit' to quit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		r.history = append(r.history, input)

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
			r.showHelp()
		case "load":
			if len(args) < 2 {
				fmt.Println("Usage: load <file>")
				continue
			}
			if err := r.loadFile(args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "info":
			if err := r.showInfo(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "preview":
			n := 5
			if len(args) > 1 {
				if n_, err := strconv.Atoi(args[1]); err == nil {
					n = n_
				}
			}
			if err := r.showPreview(n); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "stats":
			if err := r.showStats(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "summarize":
			var cols []string
			if len(args) > 1 {
				cols = strings.Split(args[1], ",")
			}
			if summary, err := r.Summarize(cols); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(summary.Format(r.formats["stats"]))
			}
		case "correlate":
			var cols []string
			if len(args) > 1 {
				cols = strings.Split(args[1], ",")
			}
			if corr, err := r.Correlate(cols); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(corr.Format(r.formats["stats"]))
			}
		case "pivot":
			if len(args) < 4 {
				fmt.Println("Usage: pivot <row_col> <col_col> <value_col> [agg=sum]")
				continue
			}
			agg := "sum"
			if len(args) > 4 {
				agg = args[4]
			}
			if pivot, err := r.Pivot(args[1], args[2], args[3], agg); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(pivot.Format(r.formats["default"]))
			}
		case "dates":
			if len(args) < 2 {
				fmt.Println("Usage: dates <date_column>")
				continue
			}
			if analysis, err := r.DateAnalysis(args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(analysis.Format(r.formats["stats"]))
			}
		case "undo":
			if err := r.Undo(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Operation undone")
				r.showPreview(5)
			}
		case "redo":
			if err := r.Redo(); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Operation redone")
				r.showPreview(5)
			}
		default:
			fmt.Printf("Unknown command: %s\nType 'help' for available commands\n", command)
		}
	}
}

func (r *REPL) showHelp() {
	help := NewTable([]string{"Command", "Description"})
	help.AddRow([]string{"load <file>", "Load a CSV file"})
	help.AddRow([]string{"info", "Show table information"})
	help.AddRow([]string{"preview [n]", "Show first n rows (default: 5)"})
	help.AddRow([]string{"stats", "Show basic column statistics"})
	help.AddRow([]string{"summarize [cols]", "Show detailed statistics for columns"})
	help.AddRow([]string{"correlate [cols]", "Show correlation matrix for numeric columns"})
	help.AddRow([]string{"pivot <row> <col> <val> [agg]", "Create pivot table with aggregation"})
	help.AddRow([]string{"dates <col>", "Analyze dates in a column"})
	help.AddRow([]string{"undo", "Undo last operation"})
	help.AddRow([]string{"redo", "Redo last undone operation"})
	help.AddRow([]string{"help", "Show this help message"})
	help.AddRow([]string{"exit", "Exit the REPL"})

	fmt.Println(help.Format(r.formats["compact"]))
}

func (r *REPL) loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	table, err := ReadTable(file, DefaultConfig())
	if err != nil {
		return fmt.Errorf("error reading table: %w", err)
	}

	r.currentTable = table
	r.currentFile = path
	fmt.Printf("Loaded %d rows from %s\n", len(table.Rows), path)
	return nil
}

func (r *REPL) showInfo() error {
	if r.currentTable == nil {
		return fmt.Errorf("no table loaded")
	}

	info := NewTable([]string{"Property", "Value"})
	info.AddRow([]string{"File", r.currentFile})
	info.AddRow([]string{"Rows", fmt.Sprintf("%d", len(r.currentTable.Rows))})
	info.AddRow([]string{"Columns", fmt.Sprintf("%d", len(r.currentTable.Headers))})

	fmt.Println(info.Format(r.formats["compact"]))

	cols := NewTable([]string{"#", "Column", "Type"})
	for i, header := range r.currentTable.Headers {
		colType, _ := r.currentTable.GetColumnType(header)
		cols.AddRow([]string{
			fmt.Sprintf("%d", i+1),
			header,
			fmt.Sprintf("%v", colType),
		})
	}

	fmt.Println("\nColumns:")
	fmt.Println(cols.Format(r.formats["compact"]))
	return nil
}

func (r *REPL) showPreview(n int) error {
	if r.currentTable == nil {
		return fmt.Errorf("no table loaded")
	}

	preview := NewTable(r.currentTable.Headers)
	for i := 0; i < min(n, len(r.currentTable.Rows)); i++ {
		preview.AddRow(r.currentTable.Rows[i])
	}
	fmt.Println(preview.Format(r.formats["default"]))
	return nil
}

func (r *REPL) showStats() error {
	if r.currentTable == nil {
		return fmt.Errorf("no table loaded")
	}

	stats := NewTable([]string{"Column", "Type", "Unique Values", "Null Count"})
	for _, header := range r.currentTable.Headers {
		col, _ := r.currentTable.GetColumn(header)
		colType, _ := r.currentTable.GetColumnType(header)

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
			fmt.Sprintf("%d", len(unique)),
			fmt.Sprintf("%d", nullCount),
		})
	}

	fmt.Println(stats.Format(r.formats["stats"]))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
