package pkg

import (
	"io"
	"strings"
	"testing"
)

func TestNewReader(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		cfg         Config
		wantErr     bool
		errContains string
	}{
		{
			name:  "valid config",
			input: "a,b,c",
			cfg:   DefaultConfig(),
		},
		{
			name: "invalid config - same delimiter and quote",
			cfg: Config{
				Delimiter: ',',
				Quote:     ',',
			},
			wantErr:     true,
			errContains: "delimiter, quote, and comment must be distinct",
		},
		{
			name: "invalid config - same delimiter and comment",
			cfg: Config{
				Delimiter: ',',
				Comment:   ',',
			},
			wantErr:     true,
			errContains: "delimiter, quote, and comment must be distinct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			_, err := NewReader(reader, tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("NewReader() error = %v, want error containing %v", err, tt.errContains)
			}
		})
	}
}

func TestReadRecord(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		cfg     Config
		want    [][]string
		wantErr bool
	}{
		{
			name:  "simple csv",
			input: "a,b,c\n1,2,3",
			cfg:   DefaultConfig(),
			want: [][]string{
				{"a", "b", "c"},
				{"1", "2", "3"},
			},
		},
		{
			name:  "quoted fields",
			input: `"a,a","b""b",c` + "\n" + `"1,1","2""2",3`,
			cfg:   DefaultConfig(),
			want: [][]string{
				{"a,a", `b"b`, "c"},
				{"1,1", `2"2`, "3"},
			},
		},
		{
			name:  "custom delimiter",
			input: "a;b;c\n1;2;3",
			cfg: Config{
				Delimiter: ';',
				Quote:     '"',
			},
			want: [][]string{
				{"a", "b", "c"},
				{"1", "2", "3"},
			},
		},
		{
			name:  "with comments",
			input: "# header\na,b,c\n#skip this\n1,2,3",
			cfg: Config{
				Delimiter: ',',
				Quote:     '"',
				Comment:   '#',
			},
			want: [][]string{
				{"a", "b", "c"},
				{"1", "2", "3"},
			},
		},
		{
			name:  "with null values",
			input: `a,\N,c`,
			cfg: Config{
				Delimiter: ',',
				Quote:     '"',
				Null:      "\\N",
			},
			want: [][]string{
				{"a", "", "c"},
			},
		},
		{
			name:  "trim leading whitespace",
			input: "a, b ,c\n1, 2 ,3",
			cfg: Config{
				Delimiter:   ',',
				Quote:       '"',
				TrimLeading: true,
			},
			want: [][]string{
				{"a", "b ", "c"},
				{"1", "2 ", "3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := NewReader(strings.NewReader(tt.input), tt.cfg)
			if err != nil {
				t.Fatalf("NewReader() error = %v", err)
			}

			var got [][]string
			for {
				record, err := reader.ReadRecord()
				if err == io.EOF {
					break
				}
				if err != nil {
					if !tt.wantErr {
						t.Errorf("ReadRecord() error = %v", err)
					}
					return
				}
				got = append(got, record)
			}

			if len(got) != len(tt.want) {
				t.Errorf("ReadRecord() got %d records, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if len(got[i]) != len(tt.want[i]) {
					t.Errorf("Record %d: got %d fields, want %d", i, len(got[i]), len(tt.want[i]))
					continue
				}
				for j := range got[i] {
					if got[i][j] != tt.want[i][j] {
						t.Errorf("Record %d, Field %d: got %q, want %q", i, j, got[i][j], tt.want[i][j])
					}
				}
			}
		})
	}
}

func TestPosition(t *testing.T) {
	input := "a,b,c\n1,2,3"
	reader, err := NewReader(strings.NewReader(input), DefaultConfig())
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}

	// Read first record
	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("ReadRecord() error = %v", err)
	}
	if got := reader.CurrentRow(); got != 1 {
		t.Errorf("CurrentRow() = %v, want %v", got, 1)
	}
	if got := len(record); got != 3 {
		t.Errorf("len(record) = %v, want %v", got, 3)
	}
}

func BenchmarkReadRecord(b *testing.B) {
	input := strings.Repeat("field1,field2,field3,field4,field5\n", 1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader, _ := NewReader(strings.NewReader(input), DefaultConfig())
		b.StartTimer()

		for {
			_, err := reader.ReadRecord()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkReadRecordWithQuotes(b *testing.B) {
	input := strings.Repeat(`"field,1","field,2","field,3","field,4","field,5"`+"\n", 1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader, _ := NewReader(strings.NewReader(input), DefaultConfig())
		b.StartTimer()

		for {
			_, err := reader.ReadRecord()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
