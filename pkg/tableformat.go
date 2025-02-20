package pkg

import (
	"fmt"
	"strings"
)

// Color codes for terminal output
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	Black   = "\033[30m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgGreen = "\033[42m"
	BgBlue  = "\033[44m"
)

// BorderStyle defines the characters used for table borders
type BorderStyle struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	TopT        string
	BottomT     string
	LeftT       string
	RightT      string
	Cross       string
	Horizontal  string
	Vertical    string
}

var (
	// DefaultStyle uses simple ASCII characters
	DefaultStyle = BorderStyle{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		TopT:        "+",
		BottomT:     "+",
		LeftT:       "+",
		RightT:      "+",
		Cross:       "+",
		Horizontal:  "-",
		Vertical:    "|",
	}

	// FancyStyle uses Unicode box-drawing characters
	FancyStyle = BorderStyle{
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		TopT:        "╦",
		BottomT:     "╩",
		LeftT:       "╠",
		RightT:      "╣",
		Cross:       "╬",
		Horizontal:  "═",
		Vertical:    "║",
	}

	// RoundedStyle uses Unicode rounded box-drawing characters
	RoundedStyle = BorderStyle{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		TopT:        "┬",
		BottomT:     "┴",
		LeftT:       "├",
		RightT:      "┤",
		Cross:       "┼",
		Horizontal:  "─",
		Vertical:    "│",
	}
)

// FormatOptions defines the styling options for table formatting
type FormatOptions struct {
	Style           BorderStyle
	HeaderStyle     string   // ANSI style for headers
	HeaderColor     string   // ANSI color for headers
	BorderColor     string   // ANSI color for borders
	AlternateRows   bool     // Whether to color alternate rows
	AlternateColor  string   // Color for alternate rows
	NumberedRows    bool     // Whether to add row numbers
	MaxColumnWidth  int      // Maximum width for any column (0 for unlimited)
	Alignment       []string // Alignment for each column ("left", "right", "center")
	FooterSeparator bool     // Whether to add separator before footer
	WrapText        bool     // Whether to wrap text in cells
	HideHeaders     bool     // Whether to hide headers
	CompactBorders  bool     // Whether to use compact borders
}

// DefaultFormat returns the default formatting options
func DefaultFormat() FormatOptions {
	return FormatOptions{
		Style:          RoundedStyle,
		HeaderStyle:    Bold,
		HeaderColor:    Cyan,
		BorderColor:    Blue,
		AlternateRows:  true,
		AlternateColor: Dim,
		MaxColumnWidth: 50,
		WrapText:       true,
	}
}

// Format returns a formatted string representation of the table
func (t *Table) Format(opts FormatOptions) string {
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
			if opts.MaxColumnWidth > 0 && len(cell) > opts.MaxColumnWidth {
				if len(cell) > widths[i] {
					widths[i] = opts.MaxColumnWidth
				}
			} else if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Write top border
	writeHorizontalBorder(&sb, widths, opts, true)
	sb.WriteString("\n")

	// Write headers
	if !opts.HideHeaders {
		sb.WriteString(opts.Style.Vertical)
		if opts.NumberedRows {
			sb.WriteString(" # ")
			sb.WriteString(opts.Style.Vertical)
		}
		for i, h := range t.Headers {
			sb.WriteString(" ")
			cell := FormatCell(h, widths[i], getAlignment(opts.Alignment, i, "center"))
			sb.WriteString(opts.HeaderColor + opts.HeaderStyle + cell + Reset)
			sb.WriteString(" " + opts.Style.Vertical)
		}
		sb.WriteString("\n")
		writeHorizontalBorder(&sb, widths, opts, false)
		sb.WriteString("\n")
	}

	// Write rows
	for rowIdx, row := range t.Rows {
		// Handle text wrapping
		if opts.WrapText {
			wrappedCells := make([][]string, len(row))
			maxLines := 1
			for i, cell := range row {
				if opts.MaxColumnWidth > 0 && len(cell) > opts.MaxColumnWidth {
					wrappedCells[i] = WrapText(cell, opts.MaxColumnWidth)
					if len(wrappedCells[i]) > maxLines {
						maxLines = len(wrappedCells[i])
					}
				} else {
					wrappedCells[i] = []string{cell}
				}
			}

			// Write each line of the wrapped cells
			for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
				writeRowBorder(&sb, opts)
				if opts.NumberedRows {
					if lineIdx == 0 {
						sb.WriteString(fmt.Sprintf(" %2d ", rowIdx+1))
					} else {
						sb.WriteString("    ")
					}
					sb.WriteString(opts.Style.Vertical)
				}

				for i := range row {
					sb.WriteString(" ")
					if lineIdx < len(wrappedCells[i]) {
						cell := FormatCell(wrappedCells[i][lineIdx], widths[i], getAlignment(opts.Alignment, i, "left"))
						if opts.AlternateRows && rowIdx%2 == 1 {
							cell = opts.AlternateColor + cell + Reset
						}
						sb.WriteString(cell)
					} else {
						sb.WriteString(strings.Repeat(" ", widths[i]))
					}
					sb.WriteString(" " + opts.Style.Vertical)
				}
				sb.WriteString("\n")
			}
		} else {
			writeRowBorder(&sb, opts)
			if opts.NumberedRows {
				sb.WriteString(fmt.Sprintf(" %2d ", rowIdx+1))
				sb.WriteString(opts.Style.Vertical)
			}

			for i, cell := range row {
				sb.WriteString(" ")
				formattedCell := FormatCell(cell, widths[i], getAlignment(opts.Alignment, i, "left"))
				if opts.AlternateRows && rowIdx%2 == 1 {
					formattedCell = opts.AlternateColor + formattedCell + Reset
				}
				sb.WriteString(formattedCell)
				sb.WriteString(" " + opts.Style.Vertical)
			}
			sb.WriteString("\n")
		}
	}

	// Write bottom border
	writeHorizontalBorder(&sb, widths, opts, false)
	sb.WriteString("\n")

	return sb.String()
}

// Helper functions

func writeHorizontalBorder(sb *strings.Builder, widths []int, opts FormatOptions, isTop bool) {
	if isTop {
		sb.WriteString(opts.BorderColor + opts.Style.TopLeft + Reset)
	} else {
		sb.WriteString(opts.BorderColor + opts.Style.BottomLeft + Reset)
	}

	if opts.NumberedRows {
		sb.WriteString(opts.BorderColor + strings.Repeat(opts.Style.Horizontal, 4) + Reset)
		if isTop {
			sb.WriteString(opts.BorderColor + opts.Style.TopT + Reset)
		} else {
			sb.WriteString(opts.BorderColor + opts.Style.BottomT + Reset)
		}
	}

	for i, width := range widths {
		sb.WriteString(opts.BorderColor + strings.Repeat(opts.Style.Horizontal, width+2) + Reset)
		if i < len(widths)-1 {
			if isTop {
				sb.WriteString(opts.BorderColor + opts.Style.TopT + Reset)
			} else {
				sb.WriteString(opts.BorderColor + opts.Style.BottomT + Reset)
			}
		}
	}

	if isTop {
		sb.WriteString(opts.BorderColor + opts.Style.TopRight + Reset)
	} else {
		sb.WriteString(opts.BorderColor + opts.Style.BottomRight + Reset)
	}
}

func writeRowBorder(sb *strings.Builder, opts FormatOptions) {
	sb.WriteString(opts.BorderColor + opts.Style.Vertical + Reset)
}

func FormatCell(content string, width int, alignment string) string {
	if len(content) > width {
		return content[:width-3] + "..."
	}

	switch alignment {
	case "right":
		return fmt.Sprintf("%*s", width, content)
	case "center":
		padding := width - len(content)
		leftPad := padding / 2
		rightPad := padding - leftPad
		return fmt.Sprintf("%*s%s%*s", leftPad, "", content, rightPad, "")
	default: // "left"
		return fmt.Sprintf("%-*s", width, content)
	}
}

func getAlignment(alignments []string, index int, defaultAlign string) string {
	if index < len(alignments) {
		return strings.ToLower(alignments[index])
	}
	return defaultAlign
}

func WrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	line := ""
	words := strings.Fields(text)

	for _, word := range words {
		if len(line)+len(word)+1 <= width {
			if line != "" {
				line += " "
			}
			line += word
		} else {
			if line != "" {
				lines = append(lines, line)
			}
			if len(word) > width {
				// Word is longer than width, need to split it
				for len(word) > width {
					lines = append(lines, word[:width])
					word = word[width:]
				}
				if word != "" {
					line = word
				} else {
					line = ""
				}
			} else {
				line = word
			}
		}
	}

	if line != "" {
		lines = append(lines, line)
	}

	return lines
}
