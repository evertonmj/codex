// Package compression provides multiple compression algorithms for reducing
// storage costs while maintaining data integrity.
//
// Supported algorithms:
//   - None: No compression (pass-through)
//   - Gzip: Balanced speed and compression ratio (5-10x effective)
//   - Zstd: Best compression ratio (10-20x effective)
//   - Snappy: Fastest with lower compression (2-4x effective)
//
// All compressed data includes a 2-byte header identifying the algorithm
// for automatic detection during decompression.
//
// Example:
//
//	data := []byte("repetitive data repetitive data...")
//	compressed, err := compression.Compress(data, compression.Gzip, 6)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	decompressed, err := compression.Decompress(compressed)
//	if err != nil {
//	    log.Fatal(err)
//	}
package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
)

// Algorithm represents the compression algorithm to use.
type Algorithm int

const (
	// None indicates no compression.
	None Algorithm = iota
	// Gzip uses the gzip compression algorithm (good balance).
	Gzip
	// Zstd uses the Zstandard compression algorithm (best compression).
	Zstd
	// Snappy uses the Snappy compression algorithm (fastest).
	Snappy
)

// String returns the string representation of the algorithm.
func (a Algorithm) String() string {
	switch a {
	case None:
		return "none"
	case Gzip:
		return "gzip"
	case Zstd:
		return "zstd"
	case Snappy:
		return "snappy"
	default:
		return "unknown"
	}
}

// Compress compresses data using the specified algorithm and level.
// Level is only used for Gzip and Zstd (1-9 for Gzip, ignored for Snappy).
// Returns the compressed data with a 2-byte header: [algorithm][level].
func Compress(data []byte, algo Algorithm, level int) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	// For None algorithm, still add header for consistency
	if algo == None {
		result := make([]byte, 2+len(data))
		result[0] = byte(None)
		result[1] = 0
		copy(result[2:], data)
		return result, nil
	}

	// Validate compression level for algorithms that use it
	if algo == Gzip {
		if level < gzip.DefaultCompression || level > gzip.BestCompression {
			level = gzip.DefaultCompression
		}
	} else if algo == Zstd {
		if level < 1 || level > 9 {
			level = 3 // Default zstd level
		}
	}

	var compressed bytes.Buffer

	switch algo {
	case Gzip:
		writer, err := gzip.NewWriterLevel(&compressed, level)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			writer.Close()
			return nil, fmt.Errorf("failed to compress with gzip: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}

	case Zstd:
		writer, err := zstd.NewWriter(&compressed, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			writer.Close()
			return nil, fmt.Errorf("failed to compress with zstd: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close zstd writer: %w", err)
		}

	case Snappy:
		compressed.Write(snappy.Encode(nil, data))

	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %d", algo)
	}

	// Prepend header: [algorithm byte][level byte][compressed data]
	result := make([]byte, 2+compressed.Len())
	result[0] = byte(algo)
	result[1] = byte(level)
	copy(result[2:], compressed.Bytes())

	return result, nil
}

// Decompress decompresses data that was compressed with Compress.
// Automatically detects the algorithm from the 2-byte header.
func Decompress(data []byte) ([]byte, error) {
	if len(data) < 2 {
		// No header, assume uncompressed
		return data, nil
	}

	// Read header
	algo := Algorithm(data[0])
	// level := int(data[1]) // Not needed for decompression
	compressedData := data[2:]

	if algo == None {
		return compressedData, nil
	}

	var decompressed bytes.Buffer

	switch algo {
	case Gzip:
		reader, err := gzip.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		if _, err := io.Copy(&decompressed, reader); err != nil {
			reader.Close()
			return nil, fmt.Errorf("failed to decompress with gzip: %w", err)
		}
		if err := reader.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip reader: %w", err)
		}

	case Zstd:
		reader, err := zstd.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer reader.Close()
		if _, err := io.Copy(&decompressed, reader); err != nil {
			return nil, fmt.Errorf("failed to decompress with zstd: %w", err)
		}

	case Snappy:
		decoded, err := snappy.Decode(nil, compressedData)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress with snappy: %w", err)
		}
		decompressed.Write(decoded)

	default:
		return nil, fmt.Errorf("unsupported compression algorithm in header: %d", algo)
	}

	return decompressed.Bytes(), nil
}

// CompressionRatio calculates the compression ratio (original/compressed).
// A ratio > 1.0 means compression was effective.
func CompressionRatio(originalSize, compressedSize int) float64 {
	if compressedSize == 0 {
		return 0
	}
	return float64(originalSize) / float64(compressedSize)
}

// SpaceSavings calculates the percentage of space saved by compression.
func SpaceSavings(originalSize, compressedSize int) float64 {
	if originalSize == 0 {
		return 0
	}
	return (1.0 - float64(compressedSize)/float64(originalSize)) * 100.0
}
