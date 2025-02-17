.PHONY: build test test-verbose test-race test-short coverage coverage-func coverage-html lint clean help bench bench-all bench-cpu bench-mem bench-profile analyze-cpu analyze-mem analyze-block analyze-mutex test-pkg fmt

# Default target
all: build

# Build the binary
build:
	go build -v -o csv_parser

# Run all tests
test: test-race
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Run short tests only
test-short:
	go test -short ./...

# Run tests for specific package
test-pkg:
	@if [ "$(pkg)" = "" ]; then \
		echo "Usage: make test-pkg pkg=<package_name>"; \
		echo "Example: make test-pkg pkg=table"; \
		exit 1; \
	fi
	go test -v ./pkg/$(pkg)

# Run tests with coverage
coverage: coverage-html coverage-func

# Generate coverage report
coverage-func:
	@echo "Coverage by function:"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html:
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run all tests and generate coverage badge
coverage-badge: coverage
	@echo "Generating coverage badge..."
	@go test -covermode=atomic -coverprofile=coverage.out ./...
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+') ;\
	echo "Total coverage: $$COVERAGE%"

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w -s .
	@echo "Code formatting complete"

# Run linter
lint: fmt
	@echo "Running go vet..."
	@go vet ./... || (echo "go vet failed"; exit 1)
	@if command -v golint >/dev/null 2>&1; then \
		echo "Running golint..."; \
		golint -set_exit_status ./... || (echo "golint failed"; exit 1); \
	else \
		echo "Warning: golint not installed. Run 'make setup' to install it."; \
	fi
	@if command -v staticcheck >/dev/null 2>&1; then \
		echo "Running staticcheck..."; \
		staticcheck ./... || (echo "staticcheck failed"; exit 1); \
	else \
		echo "Warning: staticcheck not installed. Run 'make setup' to install it."; \
	fi
	@echo "All lint checks passed!"

# Clean build artifacts
clean:
	rm -f csv_parser
	rm -f coverage.out
	rm -f coverage.html
	rm -f bench_results.txt
	rm -rf testdata/bench
	rm -rf profiles/
	rm -rf testdata/
	rm -f *.prof
	rm -f *.test
	rm -f *.out
	rm -f *.html
	rm -f *.txt

# Install development dependencies
setup:
	@echo "Installing development dependencies..."
	go mod download
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/perf/cmd/benchstat@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Installing Graphviz for profile visualization..."
	@if command -v apt-get >/dev/null; then \
		sudo apt-get update && sudo apt-get install -y graphviz; \
	elif command -v brew >/dev/null; then \
		brew install graphviz; \
	elif command -v pacman >/dev/null; then \
		sudo pacman -S graphviz; \
	elif command -v yum >/dev/null; then \
		sudo yum install -y graphviz; \
	elif command -v choco >/dev/null; then \
		choco install graphviz; \
	else \
		echo "Please install Graphviz manually: https://graphviz.org/download/"; \
		exit 1; \
	fi
	@echo "Setup complete!"

# Run quick benchmarks
bench:
	@echo "Running quick benchmarks..."
	@go test -bench=. -benchmem ./pkg/...

# Run comprehensive benchmarks with the CLI tool
bench-cli: build
	@echo "Generating benchmark data..."
	@mkdir -p testdata/bench
	@./csv_parser bench --generate
	@echo "Running benchmarks..."
	@./csv_parser bench

# Run all benchmarks (comprehensive)
bench-all: bench-cli
	@echo "Running comprehensive benchmarks..."
	@go test -bench=. -benchmem -count=5 ./pkg/... | tee bench_results.txt
	@benchstat bench_results.txt

# Run CPU profiling benchmarks
bench-cpu: profiles
	@echo "Running CPU profiling benchmarks..."
	@echo "Profiling main package..."
	@go test -run=^$$ -bench=. -cpuprofile=profiles/pkg_cpu.prof ./pkg
	@echo "Profiling benchmark package..."
	@go test -run=^$$ -bench=. -cpuprofile=profiles/benchmark_cpu.prof ./pkg/benchmark
	@echo "\nTo analyze CPU profiles:"
	@echo "Main package:    go tool pprof -http=:8080 profiles/pkg_cpu.prof"
	@echo "Benchmark pkg:   go tool pprof -http=:8081 profiles/benchmark_cpu.prof"

# Run memory profiling benchmarks
bench-mem: profiles
	@echo "Running memory profiling benchmarks..."
	@echo "Profiling main package..."
	@go test -run=^$$ -bench=. -memprofile=profiles/pkg_mem.prof ./pkg
	@echo "Profiling benchmark package..."
	@go test -run=^$$ -bench=. -memprofile=profiles/benchmark_mem.prof ./pkg/benchmark
	@echo "\nTo analyze memory profiles:"
	@echo "Main package:    go tool pprof -http=:8080 profiles/pkg_mem.prof"
	@echo "Benchmark pkg:   go tool pprof -http=:8081 profiles/benchmark_mem.prof"

# Create profiles directory
profiles:
	@mkdir -p profiles

# Run benchmarks with profiling for each package separately
bench-profile: profiles
	@echo "Running benchmarks with profiles..."
	@echo "Profiling main package..."
	go test -run=^$$ -bench=. \
		-cpuprofile=profiles/cpu.prof \
		-memprofile=profiles/mem.prof \
		-blockprofile=profiles/block.prof \
		-mutexprofile=profiles/mutex.prof \
		./pkg

	@echo "Profiling benchmark package..."
	go test -run=^$$ -bench=. \
		-cpuprofile=profiles/benchmark_cpu.prof \
		-memprofile=profiles/benchmark_mem.prof \
		-blockprofile=profiles/benchmark_block.prof \
		-mutexprofile=profiles/benchmark_mutex.prof \
		./pkg/benchmark

# Analysis targets
analyze-pkg-cpu:
	go tool pprof -http=:8080 profiles/pkg_cpu.prof

analyze-benchmark-cpu:
	go tool pprof -http=:8081 profiles/benchmark_cpu.prof

analyze-pkg-mem:
	go tool pprof -http=:8080 profiles/pkg_mem.prof

analyze-benchmark-mem:
	go tool pprof -http=:8081 profiles/benchmark_mem.prof

# Help command
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test          - Run all tests with race detection"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-short    - Run only short tests"
	@echo "  test-pkg      - Run tests for specific package (usage: make test-pkg pkg=table)"
	@echo "  coverage      - Generate both HTML and func coverage reports"
	@echo "  coverage-html - Generate HTML coverage report"
	@echo "  coverage-func - Generate func coverage report"
	@echo "  coverage-badge- Generate coverage badge"
	@echo "  lint          - Run all linters"
	@echo "  clean         - Clean build artifacts"
	@echo "  setup         - Install development dependencies"
	@echo "  bench         - Run quick benchmarks"
	@echo "  bench-cli     - Run benchmarks using CLI tool"
	@echo "  bench-all     - Run comprehensive benchmarks"
	@echo "  bench-cpu     - Run CPU profiling benchmarks"
	@echo "  bench-mem     - Run memory profiling benchmarks"
	@echo "  bench-profile - Run benchmarks with all profiles"
	@echo "  analyze-cpu   - Analyze CPU profile"
	@echo "  analyze-mem   - Analyze memory profile"
	@echo "  analyze-block - Analyze block profile"
	@echo "  analyze-mutex - Analyze mutex profile"
	@echo "  fmt           - Format code"
	@echo "  help          - Show this help message" 