package pkg

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// REPL represents the interactive CSV analysis environment
type REPL struct {
	currentTable *Table
	currentFile  string
	undoStack    []*Table
	redoStack    []*Table
	formats      map[string]FormatOptions
	history      []string
}

// NewREPL creates a new REPL instance
func NewREPL() *REPL {
	return &REPL{
		undoStack: make([]*Table, 0),
		redoStack: make([]*Table, 0),
		formats:   make(map[string]FormatOptions),
		history:   make([]string, 0),
	}
}

// pushUndo adds the current table state to the undo stack
func (r *REPL) pushUndo() {
	if r.currentTable != nil {
		r.undoStack = append(r.undoStack, r.currentTable.Copy())
		r.redoStack = nil // Clear redo stack when new action is performed
	}
}

// Start begins the REPL session
func (r *REPL) Start() {
	fmt.Println("Welcome to the CSV Parser REPL!")
	fmt.Println("Type 'help' for available commands or 'exit' to quit")

	scanner := bufio.NewScanner(os.Stdin)
	mainFormat := DefaultFormat()

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
			if r.currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			r.showInfo()
		case "preview":
			if r.currentTable == nil {
				fmt.Println("No file loaded. Use 'load <file>' first.")
				continue
			}
			n := 5
			if len(args) > 1 {
				if n_, err := strconv.Atoi(args[1]); err == nil {
					n = n_
				}
			}
			r.showPreview(n, mainFormat)
		}
	}
}

func (r *REPL) showHelp() {
	fmt.Println(`Available commands:
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
  exit                    - Exit the REPL`)
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

	r.pushUndo() // Save current state for undo
	r.currentTable = table
	r.currentFile = path
	return nil
}

func (r *REPL) showInfo() {
	fmt.Printf("File: %s\n", r.currentFile)
	fmt.Printf("Rows: %d\n", len(r.currentTable.Rows))
	fmt.Printf("Columns: %d\n\n", len(r.currentTable.Headers))

	fmt.Println("Column Information:")
	for i, header := range r.currentTable.Headers {
		colType, _ := r.currentTable.GetColumnType(header)
		fmt.Printf("%d. %s (%v)\n", i+1, header, colType)
	}
}

func (r *REPL) showPreview(n int, format FormatOptions) {
	preview := NewTable(r.currentTable.Headers)
	for i := 0; i < minimum(n, len(r.currentTable.Rows)); i++ {
		if err := preview.AddRow(r.currentTable.Rows[i]); err != nil {
			fmt.Printf("Error creating preview: %v\n", err)
			return
		}
	}
	fmt.Println(preview.Format(format))
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
