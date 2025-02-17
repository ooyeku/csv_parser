.PHONY: build test coverage lint clean help bench bench-all bench-cpu bench-mem bench-profile

# Default target
all: build

# Build the binary
build:
	go build -v -o csv_parser

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	go vet ./...
	test -z $(gofmt -l .)

# Clean build artifacts
clean:
	rm -f csv_parser
	rm -f coverage.out
	rm -f coverage.html
	rm -f cpu.prof
	rm -f mem.prof
	rm -rf testdata/bench

# Install development dependencies
setup:
	go mod download
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/perf/cmd/benchstat@latest

# Run quick benchmarks
bench:
	@echo "Running quick benchmarks..."
	@go test -bench=. -benchmem ./pkg/...

# Run comprehensive benchmarks with the CLI tool
bench-cli:
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
bench-cpu:
	@echo "Running CPU profiling benchmarks..."
	@go test -bench=. -cpuprofile=cpu.prof ./pkg/...
	@echo "To analyze CPU profile:"
	@echo "go tool pprof cpu.prof"
	@echo "Or for web view: go tool pprof -http=:8080 cpu.prof"

# Run memory profiling benchmarks
bench-mem:
	@echo "Running memory profiling benchmarks..."
	@go test -bench=. -memprofile=mem.prof ./pkg/...
	@echo "To analyze memory profile:"
	@echo "go tool pprof mem.prof"
	@echo "Or for web view: go tool pprof -http=:8080 mem.prof"

# Run benchmarks with custom profile
bench-profile:
	@echo "Running benchmarks with all profiles..."
	@go test -bench=. \
		-benchmem \
		-cpuprofile=cpu.prof \
		-memprofile=mem.prof \
		-blockprofile=block.prof \
		-mutexprofile=mutex.prof \
		./pkg/...

# Help command
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  setup        - Install development dependencies"
	@echo "  bench        - Run quick benchmarks"
	@echo "  bench-cli    - Run benchmarks using CLI tool"
	@echo "  bench-all    - Run comprehensive benchmarks"
	@echo "  bench-cpu    - Run CPU profiling benchmarks"
	@echo "  bench-mem    - Run memory profiling benchmarks"
	@echo "  bench-profile- Run benchmarks with all profiles"
	@echo "  help         - Show this help message" 