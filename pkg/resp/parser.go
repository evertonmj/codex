package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser parses RESP protocol data
type Parser struct {
	reader *bufio.Reader
}

// NewParser creates a new RESP parser
func NewParser(reader *bufio.Reader) *Parser {
	return &Parser{reader: reader}
}

// Parse reads and parses a RESP value from the reader
func (p *Parser) Parse() (*Value, error) {
	// Peek at the first byte to determine type
	b, err := p.reader.Peek(1)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("unexpected EOF")
		}
		return nil, fmt.Errorf("read error: %w", err)
	}

	switch b[0] {
	case '+':
		return p.parseSimpleString()
	case '-':
		return p.parseError()
	case ':':
		return p.parseInteger()
	case '$':
		return p.parseBulkString()
	case '*':
		return p.parseArray()
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", b[0])
	}
}

// parseSimpleString parses a simple string (+OK\r\n)
func (p *Parser) parseSimpleString() (*Value, error) {
	// Discard the '+' prefix
	_, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Read until \r\n
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return NewSimpleString(line), nil
}

// parseError parses an error message (-ERR message\r\n)
func (p *Parser) parseError() (*Value, error) {
	// Discard the '-' prefix
	_, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Read until \r\n
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return NewError(line), nil
}

// parseInteger parses an integer (:1000\r\n)
func (p *Parser) parseInteger() (*Value, error) {
	// Discard the ':' prefix
	_, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Read until \r\n
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Parse the integer
	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %s", line)
	}

	return NewInteger(n), nil
}

// parseBulkString parses a bulk string ($6\r\nfoobar\r\n)
func (p *Parser) parseBulkString() (*Value, error) {
	// Discard the '$' prefix
	_, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Read the length line
	lengthLine, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Parse the length
	length, err := strconv.Atoi(lengthLine)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %s", lengthLine)
	}

	// Handle null bulk string ($-1)
	if length == -1 {
		return NewNull(), nil
	}

	// Read exactly 'length' bytes
	data := make([]byte, length)
	n, err := io.ReadFull(p.reader, data)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("unexpected EOF: expected %d bytes, got %d", length, n)
		}
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Read the trailing \r\n
	_, err = p.readLine()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return NewBulkString(data), nil
}

// parseArray parses an array (*2\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n)
func (p *Parser) parseArray() (*Value, error) {
	// Discard the '*' prefix
	_, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Read the count line
	countLine, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	// Parse the count
	count, err := strconv.Atoi(countLine)
	if err != nil {
		return nil, fmt.Errorf("invalid array count: %s", countLine)
	}

	// Handle empty array
	if count == 0 {
		return NewArray(), nil
	}

	// Handle null array (*-1)
	if count == -1 {
		return NewNull(), nil
	}

	// Parse each element
	array := make([]Value, count)
	for i := 0; i < count; i++ {
		val, err := p.Parse()
		if err != nil {
			return nil, fmt.Errorf("error parsing array element %d: %w", i, err)
		}
		array[i] = *val
	}

	return &Value{
		Type:  Array,
		Array: array,
	}, nil
}

// readLine reads until \r\n and returns the line without the \r\n
func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && len(line) > 0 {
			// Try to use what we have if it ends with \r
			line = strings.TrimSuffix(line, "\r")
			return line, nil
		}
		return "", err
	}

	// Remove both \r and \n
	line = strings.TrimSuffix(line, "\r\n")
	line = strings.TrimSuffix(line, "\n")

	return line, nil
}

// ParseCommand parses a RESP array into command name and arguments
// Returns the command name (uppercase) and string arguments
func ParseCommand(val *Value) (string, []string, error) {
	if val.Type != Array {
		return "", nil, fmt.Errorf("expected array, got %c", val.Type)
	}

	if len(val.Array) == 0 {
		return "", nil, fmt.Errorf("empty command array")
	}

	// First element should be the command name
	if val.Array[0].Type != BulkString {
		return "", nil, fmt.Errorf("expected bulk string command, got %c", val.Array[0].Type)
	}

	cmd := strings.ToUpper(string(val.Array[0].Bulk))

	// Parse remaining elements as arguments
	args := make([]string, len(val.Array)-1)
	for i := 1; i < len(val.Array); i++ {
		elem := val.Array[i]
		switch elem.Type {
		case BulkString:
			if elem.Null {
				args[i-1] = ""
			} else {
				args[i-1] = string(elem.Bulk)
			}
		case SimpleString:
			args[i-1] = elem.Str
		case Integer:
			args[i-1] = strconv.FormatInt(elem.Int, 10)
		default:
			return "", nil, fmt.Errorf("unexpected type in command array: %c", elem.Type)
		}
	}

	return cmd, args, nil
}
