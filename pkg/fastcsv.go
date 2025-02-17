package pkg

import (
	"bufio"
	_ "errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Config holds the settings for our CSV parser.
type Config struct {
	Delimiter   rune   // e.g. ',' or ';'
	Quote       rune   // e.g. '"'
	TrimLeading bool   // trim leading whitespace of unquoted fields
	Null        string // e.g. "\N" or "NULL"
	Comment     rune   // Comment character for line skipping
}

// DefaultConfig returns a default config with comma delimiter, double-quote, etc.
func DefaultConfig() Config {
	return Config{
		Delimiter:   ',',
		Quote:       '"',
		TrimLeading: false,
		Null:        "", // No null string by default
		Comment:     0,  // No comment character by default
	}
}

// Reader provides a streaming CSV parser.
type Reader struct {
	r     *bufio.Reader
	cfg   Config
	field []byte
	err   error

	// State
	inQuotes         bool
	endOfField       bool
	endOfRecord      bool
	lastCharWasQuote bool

	// Statistics
	record        []string
	currentRecord []string
	currentRowNum int64
	currentColNum int
	bytesRead     int64
}

// Pool for record slices
var recordPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 16) // Initial capacity
	},
}

// Add field buffer pooling
var fieldPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 256)
	},
}

// NewReader creates a new Reader with the given io.Reader and config.
func NewReader(rd io.Reader, cfg Config) (*Reader, error) {
	if cfg.Delimiter == cfg.Quote || cfg.Delimiter == cfg.Comment {
		return nil, fmt.Errorf("delimiter, quote, and comment must be distinct")
	}
	if cfg.Quote == 0 {
		cfg.Quote = '"' // Force default quote if disabled
	}
	return &Reader{
		r:             bufio.NewReaderSize(rd, 64*1024), // 64KB buffer, can be tuned
		cfg:           cfg,
		currentRowNum: 0,
		currentColNum: 0,
		bytesRead:     0,
	}, nil
}

// ReadRecord reads one record (a slice of string fields) from the CSV stream.
// It returns nil, io.EOF at the end of the stream, or an error.
func (cr *Reader) ReadRecord() ([]string, error) {
	if cr.err != nil {
		return nil, cr.err
	}

	// Reset state
	cr.field = cr.field[:0]
	cr.record = recordPool.Get().([]string)[:0]
	cr.currentColNum = 0

	for {
		b, err := cr.r.ReadByte()
		if err == io.EOF {
			// If we have some data in the field buffer, finalize that field.
			if len(cr.field) > 0 || cr.endOfField || cr.inQuotes {
				cr.commitField()
			}
			// We have reached the end of file
			if len(cr.record) == 0 {
				// No more records
				return nil, io.EOF
			}
			cr.currentRecord = cr.record
			cr.currentRowNum++
			return cr.record, nil
		}
		if err != nil {
			cr.err = err
			return nil, err
		}

		cr.bytesRead++

		// Handle comments
		if cr.cfg.Comment != 0 && b == byte(cr.cfg.Comment) && !cr.inQuotes && len(cr.field) == 0 && len(cr.record) == 0 {
			// Skip until end of line
			for {
				b, err := cr.r.ReadByte()
				if err != nil || b == '\n' || b == '\r' {
					if b == '\r' {
						// Check for \n in Windows line endings
						if next, err := cr.r.Peek(1); err == nil && len(next) > 0 && next[0] == '\n' {
							_, _ = cr.r.ReadByte()
						}
					}
					break
				}
			}
			continue
		}

		switch {
		case b == byte(cr.cfg.Delimiter) && !cr.inQuotes:
			cr.commitField()
		case b == byte(cr.cfg.Quote):
			if !cr.inQuotes {
				// If we're not currently in quotes, entering a quote
				// Only do so if the field is empty or we've just started
				if len(cr.field) == 0 {
					cr.inQuotes = true
					continue
				}
			} else {
				// We are in quotes
				// Check next character to see if it's an escaped quote
				peekByte, err := cr.r.Peek(1)
				if err == nil && len(peekByte) > 0 && peekByte[0] == byte(cr.cfg.Quote) {
					// Escaped quote, consume it and add a quote to the field
					_, _ = cr.r.ReadByte() // consume next
					cr.field = append(cr.field, byte(cr.cfg.Quote))
					continue
				} else {
					// End quote
					cr.inQuotes = false
					cr.lastCharWasQuote = true
					continue
				}
			}
			// If we get here, it's just a normal character
			fallthrough

		case (b == '\n' || b == '\r') && !cr.inQuotes:
			// End of record
			// If we read '\r', check for the next one being '\n' to handle Windows line endings
			if b == '\r' {
				if next, err := cr.r.Peek(1); err == nil && len(next) > 0 && next[0] == '\n' {
					_, _ = cr.r.ReadByte() // consume '\n'
				}
			}
			cr.commitField()
			cr.currentRecord = cr.record
			cr.currentRowNum++
			return cr.record, nil

		default:
			// Regular character
			// Optionally handle trimming if TrimLeading is set
			if cr.cfg.TrimLeading && len(cr.field) == 0 && !cr.inQuotes && (b == ' ' || b == '\t') {
				// skip leading whitespace if not in quotes
				continue
			}
			cr.field = append(cr.field, b)
			cr.lastCharWasQuote = false
		}
	}
}

// New field commit logic
func (cr *Reader) commitField() {
	// Save the buffer and return it to pool
	buf := cr.field
	defer func() {
		buf = buf[:0]      // reset slice
		fieldPool.Put(buf) // return to pool
	}()

	str := string(buf)

	if cr.cfg.TrimLeading {
		str = strings.TrimLeft(str, " \t")
	}
	if cr.cfg.Null != "" && str == cr.cfg.Null {
		str = ""
	}

	cr.record = append(cr.record, str)
	cr.field = fieldPool.Get().([]byte)[:0] // Get new buffer from pool
}

// FieldCount returns the number of fields in the current record
func (cr *Reader) FieldCount() int {
	if cr.currentRecord == nil {
		return 0
	}
	return len(cr.currentRecord)
}

// CurrentRow returns the current row number (1-based)
func (cr *Reader) CurrentRow() int64 {
	return cr.currentRowNum
}

// CurrentColumn returns the current column number (1-based)
func (cr *Reader) CurrentColumn() int {
	return cr.currentColNum + 1
}

// BytesRead returns the total number of bytes read from the input
func (cr *Reader) BytesRead() int64 {
	return cr.bytesRead
}

// Position returns the current parsing position for error reporting
func (cr *Reader) Position() string {
	return fmt.Sprintf("row %d, column %d", cr.currentRowNum, cr.currentColNum+1)
}
