# Fast CSV Parser

[![CI](https://github.com/ooyeku/csv_parser/actions/workflows/ci.yml/badge.svg)](https://github.com/ooyeku/csv_parser/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ooyeku/csv_parser)](https://goreportcard.com/report/github.com/ooyeku/csv_parser)
[![codecov](https://codecov.io/gh/ooyeku/csv_parser/branch/main/graph/badge.svg)](https://codecov.io/gh/ooyeku/csv_parser)
[![GoDoc](https://godoc.org/github.com/ooyeku/csv_parser?status.svg)](https://godoc.org/github.com/ooyeku/csv_parser)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, feature-rich CSV parser written in Go. Designed for speed and flexibility when handling large CSV files.

## Features

- 🚀 High-performance streaming parser
- 📦 Memory efficient with 64KB buffer
- 🔧 Configurable delimiters and quote characters
- 💡 Smart handling of:
  - Comments (skip lines starting with #)
  - NULL values (\N or custom)
  - Windows/Unix line endings
  - Quoted fields with escapes
  - Leading whitespace trimming

## Installation

```bash
go get github.com/ooyeku/csv_parser
```

## CLI Usage

### Parse CSV File

```bash
# Basic usage
csv_parser parse data.csv

# Custom delimiter and quote character
csv_parser parse --delimiter=";" --quote="'" data.csv

# Trim leading whitespace
csv_parser parse --trim data.csv
```

### Get CSV Information

```bash
csv_parser info data.csv
```

Shows:

- Total rows and columns
- Column headers
- File statistics

### Validate CSV Structure

```bash
# Basic validation
csv_parser validate data.csv

# Strict validation (no empty fields)
csv_parser validate --strict data.csv
```

## Development Commands

This section demonstrates all available make commands and their outputs.

### Basic Commands

```bash
# Build the binary
$ make build
go build -v -o csv_parser

# Run tests
$ make test
go test -v ./...
=== RUN   TestNewReader
--- PASS: TestNewReader (0.00s)
=== RUN   TestReadRecord
--- PASS: TestReadRecord (0.00s)
...

# Generate test coverage
$ make coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
# Opens coverage report in browser showing:
# - Line-by-line coverage
# - Overall coverage percentage
# - Uncovered lines highlighted

# Run linter
$ make lint
go vet ./...
test -z $(gofmt -l .)
# No output means no issues found
```

### Benchmark Commands

```bash
# Quick benchmarks
$ make bench
Running quick benchmarks...
goos: darwin
goarch: amd64
pkg: github.com/ooyeku/csv_parser/pkg
BenchmarkReadRecord-8             10000            123456 ns/op           1234 B/op          12 allocs/op
BenchmarkReadRecordWithQuotes-8    8000            234567 ns/op           2345 B/op          23 allocs/op

# Run CLI benchmarks
$ make bench-cli
Generating benchmark data...
Running benchmarks...
File: bench_simple_1k.csv
  Size: 0.05 MB
  Rows: 1000
  Time: 15.2ms
  Speed: 3.29 MB/s
  Rows/s: 65789

File: bench_quoted_100k.csv
  Size: 8.58 MB
  Rows: 100000
  Time: 890.1ms
  Speed: 9.64 MB/s
  Rows/s: 112345

# Comprehensive benchmarks with statistics
$ make bench-all
Running comprehensive benchmarks...
name                  old time/op    new time/op    delta
SimpleCSV-8            123µs ± 2%     121µs ± 1%    ~   (p=0.286 n=5)
QuotedCSV-8           234µs ± 3%     232µs ± 2%    ~   (p=0.286 n=5)

# CPU profiling
$ make bench-cpu
Running CPU profiling benchmarks...
# Creates cpu.prof
# To analyze:
go tool pprof -http=:8080 cpu.prof
# Opens web interface showing:
# - CPU flame graph
# - Hot spots
# - Call graph

# Memory profiling
$ make bench-mem
Running memory profiling benchmarks...
# Creates mem.prof
# To analyze:
go tool pprof -http=:8080 mem.prof
# Shows:
# - Memory allocation hot spots
# - Heap profile
# - Object allocation graph

# Full profiling
$ make bench-profile
Running benchmarks with all profiles...
# Creates:
# - cpu.prof (CPU usage)
# - mem.prof (memory allocations)
# - block.prof (goroutine blocking)
# - mutex.prof (mutex contention)
```

### Utility Commands

```bash
# Clean all artifacts
$ make clean
rm -f csv_parser
rm -f coverage.out
rm -f coverage.html
rm -f cpu.prof
rm -f mem.prof
rm -rf testdata/bench

# Install development dependencies
$ make setup
go mod download
go install golang.org/x/lint/golint@latest
go install golang.org/x/perf/cmd/benchstat@latest

# Show help
$ make help
Available targets:
  build        - Build the binary
  test         - Run tests
  coverage     - Run tests with coverage
  lint         - Run linter
  clean        - Clean build artifacts
  setup        - Install development dependencies
  bench        - Run quick benchmarks
  bench-cli    - Run benchmarks using CLI tool
  bench-all    - Run comprehensive benchmarks
  bench-cpu    - Run CPU profiling benchmarks
  bench-mem    - Run memory profiling benchmarks
  bench-profile- Run benchmarks with all profiles
```

## Programmatic Usage

```go
import "github.com/ooyeku/csv_parser/pkg"

// Create custom config
cfg := pkg.Config{
    Delimiter:   ',',
    Quote:       '"',
    TrimLeading: true,
    Null:        "\\N",    // Optional: custom NULL string
    Comment:     '#',      // Optional: skip comment lines
}

// Create reader
reader, err := pkg.NewReader(file, cfg)
if err != nil {
    log.Fatal(err)
}

// Read records
for {
    record, err := reader.ReadRecord()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    // Process record...
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Inspired by the Go standard library's encoding/csv package
- Built with performance and flexibility in mind
