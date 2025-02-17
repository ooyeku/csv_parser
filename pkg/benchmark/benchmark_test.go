package benchmark

import (
	"strings"
	"testing"

	"github.com/ooyeku/csv_parser/pkg"
)

func BenchmarkCSVParser(b *testing.B) {
	benchData := GenerateBenchmarkData()

	for _, data := range benchData {
		b.Run(data.Name, func(b *testing.B) {
			cfg := pkg.DefaultConfig()
			b.ResetTimer()
			b.SetBytes(data.FileSize)

			for i := 0; i < b.N; i++ {
				reader, err := pkg.NewReader(strings.NewReader(data.Content), cfg)
				if err != nil {
					b.Fatal(err)
				}

				var rowCount int
				for {
					_, err := reader.ReadRecord()
					if err != nil {
						break
					}
					rowCount++
				}
			}
		})
	}
}

func BenchmarkCSVParserWithConfig(b *testing.B) {
	// Test different configurations
	configs := map[string]pkg.Config{
		"default": pkg.DefaultConfig(),
		"with_null": {
			Delimiter: ',',
			Quote:     '"',
			Null:      "\\N",
		},
		"with_comments": {
			Delimiter: ',',
			Quote:     '"',
			Comment:   '#',
		},
		"semicolon_delimiter": {
			Delimiter: ';',
			Quote:     '"',
		},
		"trim_leading": {
			Delimiter:   ',',
			Quote:       '"',
			TrimLeading: true,
		},
	}

	// Use complex data for config testing
	data := generateComplexCSV(10000)

	for name, cfg := range configs {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			b.SetBytes(data.FileSize)

			for i := 0; i < b.N; i++ {
				reader, err := pkg.NewReader(strings.NewReader(data.Content), cfg)
				if err != nil {
					b.Fatal(err)
				}

				var rowCount int
				for {
					_, err := reader.ReadRecord()
					if err != nil {
						break
					}
					rowCount++
				}
			}
		})
	}
}

func BenchmarkCSVParserMemory(b *testing.B) {
	// Test memory allocation patterns
	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		data := generateSimpleCSV(size)
		b.Run(data.Name, func(b *testing.B) {
			cfg := pkg.DefaultConfig()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader, err := pkg.NewReader(strings.NewReader(data.Content), cfg)
				if err != nil {
					b.Fatal(err)
				}

				var rowCount int
				for {
					_, err := reader.ReadRecord()
					if err != nil {
						break
					}
					rowCount++
				}
			}
		})
	}
}
