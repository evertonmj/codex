package resp

import (
	"fmt"
	"io"
)

// Writer writes RESP protocol values
type Writer struct {
	writer io.Writer
}

// NewWriter creates a new RESP writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

// WriteValue writes a RESP value
func (w *Writer) WriteValue(v *Value) error {
	if v == nil {
		return w.WriteNull()
	}

	switch v.Type {
	case SimpleString:
		return w.WriteSimpleString(v.Str)
	case Error:
		return w.WriteError(v.Str)
	case Integer:
		return w.WriteInteger(v.Int)
	case BulkString:
		if v.Null {
			return w.WriteNull()
		}
		return w.WriteBulkString(v.Bulk)
	case Array:
		return w.WriteArray(v.Array)
	default:
		return fmt.Errorf("unknown RESP type: %c", v.Type)
	}
}

// WriteSimpleString writes a simple string (+OK\r\n)
func (w *Writer) WriteSimpleString(s string) error {
	_, err := fmt.Fprintf(w.writer, "+%s\r\n", s)
	return err
}

// WriteError writes an error message (-ERR message\r\n)
func (w *Writer) WriteError(msg string) error {
	_, err := fmt.Fprintf(w.writer, "-%s\r\n", msg)
	return err
}

// WriteErrorf writes a formatted error message
func (w *Writer) WriteErrorf(format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return w.WriteError(msg)
}

// WriteInteger writes an integer (:1000\r\n)
func (w *Writer) WriteInteger(n int64) error {
	_, err := fmt.Fprintf(w.writer, ":%d\r\n", n)
	return err
}

// WriteBulkString writes a bulk string ($6\r\nfoobar\r\n)
func (w *Writer) WriteBulkString(data []byte) error {
	// Write length
	if _, err := fmt.Fprintf(w.writer, "$%d\r\n", len(data)); err != nil {
		return err
	}

	// Write data
	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	// Write trailing \r\n
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

// WriteBulkStringStr writes a bulk string from a string
func (w *Writer) WriteBulkStringStr(s string) error {
	return w.WriteBulkString([]byte(s))
}

// WriteNull writes a null bulk string ($-1\r\n)
func (w *Writer) WriteNull() error {
	_, err := w.writer.Write([]byte("$-1\r\n"))
	return err
}

// WriteArray writes an array of values
func (w *Writer) WriteArray(values []Value) error {
	// Write array length
	if _, err := fmt.Fprintf(w.writer, "*%d\r\n", len(values)); err != nil {
		return err
	}

	// Write each element
	for i := range values {
		if err := w.WriteValue(&values[i]); err != nil {
			return fmt.Errorf("error writing array element %d: %w", i, err)
		}
	}

	return nil
}

// WriteArrayPtr writes an array of value pointers
func (w *Writer) WriteArrayPtr(values []*Value) error {
	// Write array length
	if _, err := fmt.Fprintf(w.writer, "*%d\r\n", len(values)); err != nil {
		return err
	}

	// Write each element
	for i, val := range values {
		if err := w.WriteValue(val); err != nil {
			return fmt.Errorf("error writing array element %d: %w", i, err)
		}
	}

	return nil
}

// WriteOK writes +OK response
func (w *Writer) WriteOK() error {
	return w.WriteSimpleString("OK")
}

// WritePONG writes +PONG response
func (w *Writer) WritePONG() error {
	return w.WriteSimpleString("PONG")
}

// WriteCommandError writes a command-specific error
func (w *Writer) WriteCommandError(cmd string, msg string) error {
	return w.WriteErrorf("ERR %s: %s", cmd, msg)
}

// WriteWrongArgsError writes a wrong number of arguments error
func (w *Writer) WriteWrongArgsError(cmd string) error {
	return w.WriteErrorf("ERR wrong number of arguments for '%s' command", cmd)
}

// WriteAuthError writes an authentication error
func (w *Writer) WriteAuthError(msg string) error {
	return w.WriteErrorf("NOAUTH %s", msg)
}

// WriteInternalError writes an internal error
func (w *Writer) WriteInternalError(msg string) error {
	return w.WriteError(fmt.Sprintf("ERR %s", msg))
}
