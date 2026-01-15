package resp

import (
	"bufio"
	"bytes"
	"testing"
)

func TestParseSimpleString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantStr string
		wantErr bool
	}{
		{
			name:    "simple OK",
			input:   "+OK\r\n",
			wantStr: "OK",
			wantErr: false,
		},
		{
			name:    "simple message",
			input:   "+Hello World\r\n",
			wantStr: "Hello World",
			wantErr: false,
		},
		{
			name:    "empty simple string",
			input:   "+\r\n",
			wantStr: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := NewParser(reader)
			val, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSimpleString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if val == nil && !tt.wantErr {
				t.Error("ParseSimpleString() returned nil value")
				return
			}

			if !tt.wantErr && val.Str != tt.wantStr {
				t.Errorf("ParseSimpleString() = %q, want %q", val.Str, tt.wantStr)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantStr string
		wantErr bool
	}{
		{
			name:    "simple error",
			input:   "-ERR unknown command\r\n",
			wantStr: "ERR unknown command",
			wantErr: false,
		},
		{
			name:    "error with spaces",
			input:   "-NOAUTH authentication required\r\n",
			wantStr: "NOAUTH authentication required",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := NewParser(reader)
			val, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && val.Str != tt.wantStr {
				t.Errorf("ParseError() = %q, want %q", val.Str, tt.wantStr)
			}
		})
	}
}

func TestParseInteger(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantInt int64
		wantErr bool
	}{
		{
			name:    "positive integer",
			input:   ":1000\r\n",
			wantInt: 1000,
			wantErr: false,
		},
		{
			name:    "negative integer",
			input:   ":-1\r\n",
			wantInt: -1,
			wantErr: false,
		},
		{
			name:    "zero",
			input:   ":0\r\n",
			wantInt: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := NewParser(reader)
			val, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInteger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && val.Int != tt.wantInt {
				t.Errorf("ParseInteger() = %d, want %d", val.Int, tt.wantInt)
			}
		})
	}
}

func TestParseBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData []byte
		wantNull bool
		wantErr  bool
	}{
		{
			name:     "simple bulk string",
			input:    "$5\r\nhello\r\n",
			wantData: []byte("hello"),
			wantNull: false,
			wantErr:  false,
		},
		{
			name:     "bulk string with spaces",
			input:    "$11\r\nhello world\r\n",
			wantData: []byte("hello world"),
			wantNull: false,
			wantErr:  false,
		},
		{
			name:     "empty bulk string",
			input:    "$0\r\n\r\n",
			wantData: []byte(""),
			wantNull: false,
			wantErr:  false,
		},
		{
			name:     "null bulk string",
			input:    "$-1\r\n",
			wantData: nil,
			wantNull: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := NewParser(reader)
			val, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBulkString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if val.Null != tt.wantNull {
					t.Errorf("ParseBulkString() null = %v, want %v", val.Null, tt.wantNull)
				}
				if !bytes.Equal(val.Bulk, tt.wantData) {
					t.Errorf("ParseBulkString() = %q, want %q", val.Bulk, tt.wantData)
				}
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantLen    int
		wantNull   bool
		wantErr    bool
		validate   func(*Value) bool
	}{
		{
			name:    "empty array",
			input:   "*0\r\n",
			wantLen: 0,
			wantErr: false,
			validate: func(v *Value) bool {
				return len(v.Array) == 0
			},
		},
		{
			name:    "array with bulk strings",
			input:   "*2\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n",
			wantLen: 2,
			wantErr: false,
			validate: func(v *Value) bool {
				return string(v.Array[0].Bulk) == "key1" && string(v.Array[1].Bulk) == "key2"
			},
		},
		{
			name:    "command array",
			input:   "*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			wantLen: 3,
			wantErr: false,
			validate: func(v *Value) bool {
				return string(v.Array[0].Bulk) == "CDX.SET" &&
					string(v.Array[1].Bulk) == "key" &&
					string(v.Array[2].Bulk) == "value"
			},
		},
		{
			name:     "null array",
			input:    "*-1\r\n",
			wantNull: true,
			wantErr:  false,
			validate: func(v *Value) bool {
				return v.Null
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := NewParser(reader)
			val, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if val.Null != tt.wantNull {
					t.Errorf("ParseArray() null = %v, want %v", val.Null, tt.wantNull)
				}
				if !tt.wantNull && len(val.Array) != tt.wantLen {
					t.Errorf("ParseArray() len = %d, want %d", len(val.Array), tt.wantLen)
				}
				if !tt.validate(val) {
					t.Error("ParseArray() validation failed")
				}
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCmd     string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name:     "SET command",
			input:    "*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			wantCmd:  "CDX.SET",
			wantArgs: []string{"key", "value"},
			wantErr:  false,
		},
		{
			name:     "GET command",
			input:    "*2\r\n$7\r\nCDX.GET\r\n$3\r\nkey\r\n",
			wantCmd:  "CDX.GET",
			wantArgs: []string{"key"},
			wantErr:  false,
		},
		{
			name:     "PING command",
			input:    "*1\r\n$8\r\nCDX.PING\r\n",
			wantCmd:  "CDX.PING",
			wantArgs: []string{},
			wantErr:  false,
		},
		{
			name:     "command lowercase",
			input:    "*1\r\n$8\r\ncdx.ping\r\n",
			wantCmd:  "CDX.PING",
			wantArgs: []string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := NewParser(reader)
			val, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			cmd, args, err := ParseCommand(val)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if cmd != tt.wantCmd {
				t.Errorf("ParseCommand() cmd = %q, want %q", cmd, tt.wantCmd)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("ParseCommand() args len = %d, want %d", len(args), len(tt.wantArgs))
				return
			}

			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("ParseCommand() arg[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func BenchmarkParseSimpleString(b *testing.B) {
	input := "+OK\r\n"
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(bytes.NewBufferString(input))
		parser := NewParser(reader)
		_, _ = parser.Parse()
	}
}

func BenchmarkParseBulkString(b *testing.B) {
	input := "$5\r\nhello\r\n"
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(bytes.NewBufferString(input))
		parser := NewParser(reader)
		_, _ = parser.Parse()
	}
}

func BenchmarkParseCommand(b *testing.B) {
	input := "*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(bytes.NewBufferString(input))
		parser := NewParser(reader)
		val, _ := parser.Parse()
		_, _, _ = ParseCommand(val)
	}
}
