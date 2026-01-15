package resp

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWriteSimpleString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "OK response",
			input: "OK",
			want:  "+OK\r\n",
		},
		{
			name:  "message",
			input: "Hello World",
			want:  "+Hello World\r\n",
		},
		{
			name:  "empty string",
			input: "",
			want:  "+\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)
			err := writer.WriteSimpleString(tt.input)

			if err != nil {
				t.Errorf("WriteSimpleString() error = %v", err)
				return
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("WriteSimpleString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unknown command",
			input: "ERR unknown command",
			want:  "-ERR unknown command\r\n",
		},
		{
			name:  "auth error",
			input: "NOAUTH authentication required",
			want:  "-NOAUTH authentication required\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)
			err := writer.WriteError(tt.input)

			if err != nil {
				t.Errorf("WriteError() error = %v", err)
				return
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("WriteError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteInteger(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{
			name:  "positive",
			input: 1000,
			want:  ":1000\r\n",
		},
		{
			name:  "negative",
			input: -1,
			want:  ":-1\r\n",
		},
		{
			name:  "zero",
			input: 0,
			want:  ":0\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)
			err := writer.WriteInteger(tt.input)

			if err != nil {
				t.Errorf("WriteInteger() error = %v", err)
				return
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("WriteInteger() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteBulkString(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "simple",
			input: []byte("hello"),
			want:  "$5\r\nhello\r\n",
		},
		{
			name:  "with spaces",
			input: []byte("hello world"),
			want:  "$11\r\nhello world\r\n",
		},
		{
			name:  "empty",
			input: []byte(""),
			want:  "$0\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)
			err := writer.WriteBulkString(tt.input)

			if err != nil {
				t.Errorf("WriteBulkString() error = %v", err)
				return
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("WriteBulkString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteNull(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	err := writer.WriteNull()

	if err != nil {
		t.Errorf("WriteNull() error = %v", err)
		return
	}

	want := "$-1\r\n"
	if got := buf.String(); got != want {
		t.Errorf("WriteNull() = %q, want %q", got, want)
	}
}

func TestWriteArray(t *testing.T) {
	tests := []struct {
		name  string
		input []Value
		want  string
	}{
		{
			name:  "empty array",
			input: []Value{},
			want:  "*0\r\n",
		},
		{
			name: "bulk strings",
			input: []Value{
				{Type: BulkString, Bulk: []byte("key1")},
				{Type: BulkString, Bulk: []byte("key2")},
			},
			want: "*2\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n",
		},
		{
			name: "mixed types",
			input: []Value{
				{Type: SimpleString, Str: "OK"},
				{Type: Integer, Int: 42},
				{Type: BulkString, Bulk: []byte("hello")},
			},
			want: "*3\r\n+OK\r\n:42\r\n$5\r\nhello\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)
			err := writer.WriteArray(tt.input)

			if err != nil {
				t.Errorf("WriteArray() error = %v", err)
				return
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("WriteArray() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteOK(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	err := writer.WriteOK()

	if err != nil {
		t.Errorf("WriteOK() error = %v", err)
		return
	}

	want := "+OK\r\n"
	if got := buf.String(); got != want {
		t.Errorf("WriteOK() = %q, want %q", got, want)
	}
}

func TestWritePONG(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	err := writer.WritePONG()

	if err != nil {
		t.Errorf("WritePONG() error = %v", err)
		return
	}

	want := "+PONG\r\n"
	if got := buf.String(); got != want {
		t.Errorf("WritePONG() = %q, want %q", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can write and read back the same values
	tests := []struct {
		name string
		val  *Value
	}{
		{
			name: "simple string",
			val:  NewSimpleString("OK"),
		},
		{
			name: "error",
			val:  NewError("ERR unknown command"),
		},
		{
			name: "integer",
			val:  NewInteger(42),
		},
		{
			name: "bulk string",
			val:  NewBulkString([]byte("hello")),
		},
		{
			name: "null bulk string",
			val:  NewNull(),
		},
		{
			name: "array",
			val:  NewArray(
				*NewBulkString([]byte("key1")),
				*NewBulkString([]byte("key2")),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write
			var buf bytes.Buffer
			writer := NewWriter(&buf)
			err := writer.WriteValue(tt.val)
			if err != nil {
				t.Fatalf("WriteValue() error = %v", err)
			}

			// Read back
			reader := NewParser(bufio.NewReader(bytes.NewReader(buf.Bytes())))
			got, err := reader.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare
			if !valuesEqual(tt.val, got) {
				t.Errorf("Round trip failed: original = %v, got = %v", tt.val, got)
			}
		})
	}
}

// valuesEqual checks if two RESP values are equal
func valuesEqual(a, b *Value) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case SimpleString, Error:
		return a.Str == b.Str
	case Integer:
		return a.Int == b.Int
	case BulkString:
		if a.Null != b.Null {
			return false
		}
		return bytes.Equal(a.Bulk, b.Bulk)
	case Array:
		if len(a.Array) != len(b.Array) {
			return false
		}
		for i := range a.Array {
			if !valuesEqual(&a.Array[i], &b.Array[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func BenchmarkWriteSimpleString(b *testing.B) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = writer.WriteSimpleString("OK")
	}
}

func BenchmarkWriteBulkString(b *testing.B) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	data := []byte("hello world")
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = writer.WriteBulkString(data)
	}
}

func BenchmarkWriteArray(b *testing.B) {
	array := []Value{
		{Type: BulkString, Bulk: []byte("key1")},
		{Type: BulkString, Bulk: []byte("key2")},
		{Type: BulkString, Bulk: []byte("key3")},
	}
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = writer.WriteArray(array)
	}
}
