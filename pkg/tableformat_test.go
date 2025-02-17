package pkg

import (
	"strings"
	"testing"
)

func TestDefaultFormat(t *testing.T) {
	format := DefaultFormat()

	if format.Style != RoundedStyle {
		t.Error("DefaultFormat() should use RoundedStyle")
	}

	if format.HeaderStyle != Bold {
		t.Error("DefaultFormat() should use Bold for headers")
	}

	if !format.AlternateRows {
		t.Error("DefaultFormat() should have AlternateRows enabled")
	}

	if format.MaxColumnWidth != 50 {
		t.Error("DefaultFormat() should have MaxColumnWidth of 50")
	}
}

func TestFormatCell(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		width     int
		alignment string
		want      string
	}{
		{
			name:      "left align short content",
			content:   "test",
			width:     10,
			alignment: "left",
			want:      "test      ",
		},
		{
			name:      "right align short content",
			content:   "test",
			width:     10,
			alignment: "right",
			want:      "      test",
		},
		{
			name:      "center align short content",
			content:   "test",
			width:     10,
			alignment: "center",
			want:      "   test   ",
		},
		{
			name:      "truncate long content",
			content:   "very long content",
			width:     10,
			alignment: "left",
			want:      "very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCell(tt.content, tt.width, tt.alignment)
			if got != tt.want {
				t.Errorf("formatCell() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{
			name:  "no wrap needed",
			text:  "short text",
			width: 10,
			want:  []string{"short text"},
		},
		{
			name:  "wrap on word boundary",
			text:  "this is a long text that needs wrapping",
			width: 10,
			want:  []string{"this is a", "long text", "that needs", "wrapping"},
		},
		{
			name:  "wrap long word",
			text:  "supercalifragilisticexpialidocious",
			width: 10,
			want:  []string{"supercalif", "ragilistic", "expialidoc", "ious"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("wrapText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTableFormat(t *testing.T) {
	table := NewTable([]string{"Name", "Age", "City"})
	table.AddRow([]string{"John Doe", "30", "New York"})
	table.AddRow([]string{"Jane Smith", "25", "Los Angeles"})

	tests := []struct {
		name    string
		opts    FormatOptions
		checks  []string
		exclude []string
	}{
		{
			name: "default style",
			opts: DefaultFormat(),
			checks: []string{
				"╭", "╮", // Top corners
				"│",                   // Vertical borders
				"Name", "Age", "City", // Headers
				"John Doe", "30", "New York", // Data
			},
		},
		{
			name: "fancy style",
			opts: FormatOptions{
				Style:       FancyStyle,
				HeaderStyle: Bold,
			},
			checks: []string{
				"╔", "╗", // Top corners
				"║",                   // Vertical borders
				"Name", "Age", "City", // Headers
			},
		},
		{
			name: "compact style",
			opts: FormatOptions{
				Style:          RoundedStyle,
				CompactBorders: true,
				MaxColumnWidth: 10,
			},
			checks: []string{
				"John Doe", "30", "New York",
			},
		},
		{
			name: "with row numbers",
			opts: FormatOptions{
				Style:        DefaultStyle,
				NumberedRows: true,
			},
			checks: []string{
				"1", "2", // Row numbers
				"John Doe", "30",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := table.Format(tt.opts)

			// Check required elements
			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf("Format() result should contain %q", check)
				}
			}

			// Check excluded elements
			for _, exclude := range tt.exclude {
				if strings.Contains(result, exclude) {
					t.Errorf("Format() result should not contain %q", exclude)
				}
			}
		})
	}
}

func TestBorderStyles(t *testing.T) {
	styles := []struct {
		name  string
		style BorderStyle
	}{
		{"Default", DefaultStyle},
		{"Fancy", FancyStyle},
		{"Rounded", RoundedStyle},
	}

	table := NewTable([]string{"Test"})
	table.AddRow([]string{"Data"})

	for _, style := range styles {
		t.Run(style.name, func(t *testing.T) {
			opts := FormatOptions{Style: style.style}
			result := table.Format(opts)

			// Verify top border
			if !strings.Contains(result, style.style.TopLeft) ||
				!strings.Contains(result, style.style.TopRight) {
				t.Errorf("%s style: missing top border characters", style.name)
			}

			// Verify vertical borders
			if !strings.Contains(result, style.style.Vertical) {
				t.Errorf("%s style: missing vertical border character", style.name)
			}

			// Verify bottom border
			if !strings.Contains(result, style.style.BottomLeft) ||
				!strings.Contains(result, style.style.BottomRight) {
				t.Errorf("%s style: missing bottom border characters", style.name)
			}
		})
	}
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
