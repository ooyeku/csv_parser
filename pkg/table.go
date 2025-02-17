package pkg

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Table represents a data table with headers and rows
type Table struct {
	Headers []string
	Rows    [][]string
	types   []ColumnType
	index   map[string]int // Header to column index mapping
}

// ColumnType represents the detected type of a column
type ColumnType int

const (
	TypeString ColumnType = iota
	TypeInteger
	TypeFloat
	TypeBoolean
	TypeNull
)

// NewTable creates a new table with the given headers
func NewTable(headers []string) *Table {
	index := make(map[string]int, len(headers))
	for i, h := range headers {
		index[h] = i
	}
	return &Table{
		Headers: headers,
		Rows:    make([][]string, 0),
		types:   make([]ColumnType, len(headers)),
		index:   index,
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) error {
	if len(row) != len(t.Headers) {
		return fmt.Errorf("row length %d does not match headers length %d", len(row), len(t.Headers))
	}
	t.Rows = append(t.Rows, row)
	t.updateTypes(row)
	return nil
}

// updateTypes updates the detected types for each column based on the new row
func (t *Table) updateTypes(row []string) {
	for i, val := range row {
		if t.types[i] == TypeNull {
			t.types[i] = detectType(val)
			continue
		}
		newType := detectType(val)
		if newType != t.types[i] {
			// If types conflict, fall back to string
			t.types[i] = TypeString
		}
	}
}

// detectType attempts to determine the type of a value
func detectType(val string) ColumnType {
	if val == "" || strings.EqualFold(val, "null") || strings.EqualFold(val, "\\N") {
		return TypeNull
	}
	if strings.EqualFold(val, "true") || strings.EqualFold(val, "false") {
		return TypeBoolean
	}
	if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		return TypeInteger
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return TypeFloat
	}
	return TypeString
}

// GetColumn returns all values in a column by header name
func (t *Table) GetColumn(header string) ([]string, error) {
	idx, ok := t.index[header]
	if !ok {
		return nil, fmt.Errorf("column %q not found", header)
	}
	col := make([]string, len(t.Rows))
	for i, row := range t.Rows {
		col[i] = row[idx]
	}
	return col, nil
}

// GetColumnType returns the detected type of a column
func (t *Table) GetColumnType(header string) (ColumnType, error) {
	idx, ok := t.index[header]
	if !ok {
		return TypeString, fmt.Errorf("column %q not found", header)
	}
	return t.types[idx], nil
}

// Filter returns a new table containing only rows that match the predicate
func (t *Table) Filter(predicate func(row []string) bool) *Table {
	newTable := NewTable(t.Headers)
	for _, row := range t.Rows {
		if predicate(row) {
			newTable.AddRow(row)
		}
	}
	return newTable
}

// Sort sorts the table by the specified columns
// columns should be in the format: ["name:asc", "age:desc"]
func (t *Table) Sort(columns []string) error {
	type sortKey struct {
		col  string
		desc bool
	}

	// Parse sort keys
	keys := make([]sortKey, len(columns))
	indices := make([]int, len(columns))

	for i, col := range columns {
		parts := strings.Split(col, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid sort format for %q, expected 'column:asc' or 'column:desc'", col)
		}

		idx, ok := t.index[parts[0]]
		if !ok {
			return fmt.Errorf("column %q not found", parts[0])
		}

		keys[i] = sortKey{
			col:  parts[0],
			desc: strings.EqualFold(parts[1], "desc"),
		}
		indices[i] = idx
	}

	// Sort rows
	sort.SliceStable(t.Rows, func(i, j int) bool {
		for k, key := range keys {
			idx := indices[k]
			a, b := t.Rows[i][idx], t.Rows[j][idx]
			if a == b {
				continue
			}
			less := a < b
			if key.desc {
				less = !less
			}
			return less
		}
		return false
	})

	return nil
}

// GroupBy groups rows by the specified columns and applies aggregations
func (t *Table) GroupBy(groupCols []string, aggs map[string]string) (*Table, error) {
	// Validate group columns
	groupIndices := make([]int, len(groupCols))
	for i, col := range groupCols {
		idx, ok := t.index[col]
		if !ok {
			return nil, fmt.Errorf("group column %q not found", col)
		}
		groupIndices[i] = idx
	}

	// Create result headers
	headers := make([]string, 0, len(groupCols)+len(aggs))
	headers = append(headers, groupCols...)
	for col := range aggs {
		headers = append(headers, col)
	}

	// Group rows
	groups := make(map[string][][]string)
	for _, row := range t.Rows {
		key := make([]string, len(groupIndices))
		for i, idx := range groupIndices {
			key[i] = row[idx]
		}
		groupKey := strings.Join(key, "\x00")
		groups[groupKey] = append(groups[groupKey], row)
	}

	// Apply aggregations
	result := NewTable(headers)
	for groupKey, rows := range groups {
		groupVals := strings.Split(groupKey, "\x00")
		newRow := make([]string, len(headers))
		copy(newRow, groupVals)

		// Calculate aggregations
		i := len(groupVals)
		for col, agg := range aggs {
			idx, ok := t.index[col]
			if !ok {
				return nil, fmt.Errorf("aggregation column %q not found", col)
			}

			vals := make([]string, len(rows))
			for j, row := range rows {
				vals[j] = row[idx]
			}

			aggVal, err := aggregate(vals, agg)
			if err != nil {
				return nil, fmt.Errorf("aggregation error for %q: %w", col, err)
			}
			newRow[i] = aggVal
			i++
		}

		result.AddRow(newRow)
	}

	return result, nil
}

// aggregate performs the specified aggregation on values
func aggregate(vals []string, agg string) (string, error) {
	switch strings.ToLower(agg) {
	case "count":
		return strconv.Itoa(len(vals)), nil

	case "sum":
		var sum float64
		for _, v := range vals {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return "", fmt.Errorf("invalid number %q for sum", v)
			}
			sum += f
		}
		return strconv.FormatFloat(sum, 'f', -1, 64), nil

	case "avg":
		if len(vals) == 0 {
			return "0", nil
		}
		var sum float64
		for _, v := range vals {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return "", fmt.Errorf("invalid number %q for average", v)
			}
			sum += f
		}
		avg := sum / float64(len(vals))
		return strconv.FormatFloat(avg, 'f', -1, 64), nil

	case "min":
		if len(vals) == 0 {
			return "", nil
		}
		min := vals[0]
		for _, v := range vals[1:] {
			if v < min {
				min = v
			}
		}
		return min, nil

	case "max":
		if len(vals) == 0 {
			return "", nil
		}
		max := vals[0]
		for _, v := range vals[1:] {
			if v > max {
				max = v
			}
		}
		return max, nil

	default:
		return "", fmt.Errorf("unknown aggregation %q", agg)
	}
}

// String returns a string representation of the table
func (t *Table) String() string {
	if len(t.Headers) == 0 {
		return "empty table"
	}

	// Calculate column widths
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build table string
	var sb strings.Builder

	// Write headers
	for i, h := range t.Headers {
		if i > 0 {
			sb.WriteString(" | ")
		}
		fmt.Fprintf(&sb, "%-*s", widths[i], h)
	}
	sb.WriteString("\n")

	// Write separator
	for i, w := range widths {
		if i > 0 {
			sb.WriteString("-+-")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteString("\n")

	// Write rows
	for _, row := range t.Rows {
		for i, cell := range row {
			if i > 0 {
				sb.WriteString(" | ")
			}
			fmt.Fprintf(&sb, "%-*s", widths[i], cell)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Copy creates a deep copy of the table
func (t *Table) Copy() *Table {
	newTable := NewTable(append([]string{}, t.Headers...))
	newTable.types = append([]ColumnType{}, t.types...)
	for k, v := range t.index {
		newTable.index[k] = v
	}
	for _, row := range t.Rows {
		newRow := append([]string{}, row...)
		newTable.Rows = append(newTable.Rows, newRow)
	}
	return newTable
}
