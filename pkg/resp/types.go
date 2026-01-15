package resp

import "fmt"

// RESPType represents the RESP data type marker
type RESPType byte

const (
	// SimpleString represents a simple string response (+)
	SimpleString RESPType = '+'

	// Error represents an error message (-)
	Error RESPType = '-'

	// Integer represents an integer (:)
	Integer RESPType = ':'

	// BulkString represents a bulk string ($)
	BulkString RESPType = '$'

	// Array represents an array of values (*)
	Array RESPType = '*'
)

// Value represents a RESP protocol value
type Value struct {
	Type  RESPType
	Str   string   // For SimpleString, Error
	Int   int64    // For Integer
	Bulk  []byte   // For BulkString
	Array []Value  // For Array
	Null  bool     // True if this is a null value (used with BulkString)
}

// NewSimpleString creates a simple string value
func NewSimpleString(s string) *Value {
	return &Value{
		Type: SimpleString,
		Str:  s,
	}
}

// NewError creates an error value
func NewError(msg string) *Value {
	return &Value{
		Type: Error,
		Str:  msg,
	}
}

// NewErrorf creates a formatted error value
func NewErrorf(format string, args ...interface{}) *Value {
	return &Value{
		Type: Error,
		Str:  fmt.Sprintf(format, args...),
	}
}

// NewInteger creates an integer value
func NewInteger(n int64) *Value {
	return &Value{
		Type: Integer,
		Int:  n,
	}
}

// NewBulkString creates a bulk string value
func NewBulkString(b []byte) *Value {
	return &Value{
		Type: BulkString,
		Bulk: b,
	}
}

// NewBulkStringStr creates a bulk string value from a string
func NewBulkStringStr(s string) *Value {
	return &Value{
		Type: BulkString,
		Bulk: []byte(s),
	}
}

// NewNull creates a null bulk string value ($-1)
func NewNull() *Value {
	return &Value{
		Type: BulkString,
		Null: true,
	}
}

// NewArray creates an array value
func NewArray(values ...Value) *Value {
	return &Value{
		Type:  Array,
		Array: values,
	}
}

// NewArrayPtr creates an array value from pointers
func NewArrayPtr(values ...*Value) *Value {
	arr := make([]Value, len(values))
	for i, v := range values {
		if v != nil {
			arr[i] = *v
		}
	}
	return &Value{
		Type:  Array,
		Array: arr,
	}
}

// IsNull returns true if this is a null bulk string
func (v *Value) IsNull() bool {
	return v.Type == BulkString && v.Null
}

// String returns a string representation of the value
func (v *Value) String() string {
	switch v.Type {
	case SimpleString:
		return fmt.Sprintf("+%s", v.Str)
	case Error:
		return fmt.Sprintf("-%s", v.Str)
	case Integer:
		return fmt.Sprintf(":%d", v.Int)
	case BulkString:
		if v.Null {
			return "$-1"
		}
		return fmt.Sprintf("$%d\r\n%s", len(v.Bulk), string(v.Bulk))
	case Array:
		return fmt.Sprintf("*%d", len(v.Array))
	default:
		return "unknown"
	}
}
