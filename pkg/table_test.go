package pkg

import (
	"reflect"
	"testing"
)

func TestNewTable(t *testing.T) {
	headers := []string{"id", "name", "age"}
	table := NewTable(headers)

	if !reflect.DeepEqual(table.Headers, headers) {
		t.Errorf("NewTable() headers = %v, want %v", table.Headers, headers)
	}

	if len(table.Rows) != 0 {
		t.Errorf("NewTable() rows = %v, want empty", table.Rows)
	}

	if len(table.types) != len(headers) {
		t.Errorf("NewTable() types length = %d, want %d", len(table.types), len(headers))
	}

	for header, idx := range table.index {
		if headers[idx] != header {
			t.Errorf("NewTable() index mapping incorrect for %s", header)
		}
	}
}

func TestAddRow(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		row     []string
		wantErr bool
	}{
		{
			name:    "valid row",
			headers: []string{"id", "name", "age"},
			row:     []string{"1", "John", "25"},
			wantErr: false,
		},
		{
			name:    "row too short",
			headers: []string{"id", "name", "age"},
			row:     []string{"1", "John"},
			wantErr: true,
		},
		{
			name:    "row too long",
			headers: []string{"id", "name", "age"},
			row:     []string{"1", "John", "25", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTable(tt.headers)
			err := table.AddRow(tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRow() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && !reflect.DeepEqual(table.Rows[0], tt.row) {
				t.Errorf("AddRow() row = %v, want %v", table.Rows[0], tt.row)
			}
		})
	}
}

func TestDetectType(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want ColumnType
	}{
		{"empty string", "", TypeNull},
		{"null string", "null", TypeNull},
		{"\\N string", "\\N", TypeNull},
		{"integer", "123", TypeInteger},
		{"negative integer", "-123", TypeInteger},
		{"float", "123.45", TypeFloat},
		{"negative float", "-123.45", TypeFloat},
		{"scientific notation", "1.23e-4", TypeFloat},
		{"boolean true", "true", TypeBoolean},
		{"boolean false", "false", TypeBoolean},
		{"string", "hello", TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectType(tt.val); got != tt.want {
				t.Errorf("detectType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetColumn(t *testing.T) {
	table := NewTable([]string{"id", "name", "age"})
	err := table.AddRow([]string{"1", "John", "25"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"2", "Jane", "30"})
	if err != nil {
		return
	}

	tests := []struct {
		name    string
		header  string
		want    []string
		wantErr bool
	}{
		{
			name:    "existing column",
			header:  "name",
			want:    []string{"John", "Jane"},
			wantErr: false,
		},
		{
			name:    "non-existent column",
			header:  "invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := table.GetColumn(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetColumn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	table := NewTable([]string{"id", "name", "age"})
	err := table.AddRow([]string{"1", "John", "25"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"2", "Jane", "30"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"3", "Bob", "25"})
	if err != nil {
		return
	}

	filtered := table.Filter(func(row []string) bool {
		return row[2] == "25" // Filter by age
	})

	if len(filtered.Rows) != 2 {
		t.Errorf("Filter() got %d rows, want 2", len(filtered.Rows))
	}

	for _, row := range filtered.Rows {
		if row[2] != "25" {
			t.Errorf("Filter() included row with age %s, want 25", row[2])
		}
	}
}

func TestSort(t *testing.T) {
	table := NewTable([]string{"id", "name", "age"})
	err := table.AddRow([]string{"2", "Jane", "30"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"1", "John", "25"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"3", "Bob", "25"})
	if err != nil {
		return
	}

	tests := []struct {
		name    string
		columns []string
		wantErr bool
		check   func(*Table) bool
	}{
		{
			name:    "sort by age ascending",
			columns: []string{"age:asc"},
			wantErr: false,
			check: func(t *Table) bool {
				return t.Rows[0][2] <= t.Rows[1][2] && t.Rows[1][2] <= t.Rows[2][2]
			},
		},
		{
			name:    "sort by age desc, name asc",
			columns: []string{"age:desc", "name:asc"},
			wantErr: false,
			check: func(t *Table) bool {
				return t.Rows[0][2] >= t.Rows[1][2] &&
					(t.Rows[0][2] != t.Rows[1][2] || t.Rows[0][1] <= t.Rows[1][1])
			},
		},
		{
			name:    "invalid column",
			columns: []string{"invalid:asc"},
			wantErr: true,
			check:   func(t *Table) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableCopy := NewTable(table.Headers)
			for _, row := range table.Rows {
				err := tableCopy.AddRow(row)
				if err != nil {
					return
				}
			}

			err := tableCopy.Sort(tt.columns)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.check(tableCopy) {
				t.Errorf("Sort() result failed check")
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	table := NewTable([]string{"id", "dept", "salary"})
	err := table.AddRow([]string{"1", "IT", "1000"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"2", "IT", "2000"})
	if err != nil {
		return
	}
	err = table.AddRow([]string{"3", "HR", "1500"})
	if err != nil {
		return
	}

	tests := []struct {
		name      string
		groupCols []string
		aggs      map[string]string
		wantErr   bool
		checkFn   func(*Table) bool
	}{
		{
			name:      "group by dept with sum",
			groupCols: []string{"dept"},
			aggs:      map[string]string{"salary": "sum"},
			wantErr:   false,
			checkFn: func(t *Table) bool {
				return len(t.Rows) == 2 // Should have 2 departments
			},
		},
		{
			name:      "invalid column",
			groupCols: []string{"invalid"},
			aggs:      map[string]string{"salary": "sum"},
			wantErr:   true,
			checkFn:   func(t *Table) bool { return true },
		},
		{
			name:      "invalid aggregation",
			groupCols: []string{"dept"},
			aggs:      map[string]string{"salary": "invalid"},
			wantErr:   true,
			checkFn:   func(t *Table) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := table.GroupBy(tt.groupCols, tt.aggs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GroupBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.checkFn(result) {
				t.Errorf("GroupBy() result failed check")
			}
		})
	}
}

func TestCopy(t *testing.T) {
	original := NewTable([]string{"id", "name"})
	err := original.AddRow([]string{"1", "John"})
	if err != nil {
		return
	}
	err = original.AddRow([]string{"2", "Jane"})
	if err != nil {
		return
	}

	table := original.Copy()

	// Verify headers
	if !reflect.DeepEqual(table.Headers, original.Headers) {
		t.Errorf("Copy() headers = %v, want %v", table.Headers, original.Headers)
	}

	// Verify rows
	if !reflect.DeepEqual(table.Rows, original.Rows) {
		t.Errorf("Copy() rows = %v, want %v", table.Rows, original.Rows)
	}

	// Verify types
	if !reflect.DeepEqual(table.types, original.types) {
		t.Errorf("Copy() types = %v, want %v", table.types, original.types)
	}

	// Verify index
	if !reflect.DeepEqual(table.index, original.index) {
		t.Errorf("Copy() index = %v, want %v", table.index, original.index)
	}

	// Verify deep table by modifying original
	err = original.AddRow([]string{"3", "Bob"})
	if err != nil {
		return
	}
	if len(table.Rows) == len(original.Rows) {
		t.Error("Copy() did not create a deep table")
	}
}
