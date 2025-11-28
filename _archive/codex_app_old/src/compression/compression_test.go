package compression

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"strings"
	"testing"
)

func TestCompressDecompress_Gzip(t *testing.T) {
	original := []byte("Hello, World! This is a test of gzip compression. " +
		"Compression works best with repetitive data. " +
		"Compression works best with repetitive data. " +
		"Compression works best with repetitive data.")

	compressed, err := Compress(original, Gzip, gzip.DefaultCompression)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if len(compressed) >= len(original) {
		t.Logf("Warning: Compressed size (%d) >= original size (%d)", len(compressed), len(original))
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data doesn't match original")
	}
}

func TestCompressDecompress_Zstd(t *testing.T) {
	original := []byte(strings.Repeat("This is test data for zstd compression! ", 100))

	compressed, err := Compress(original, Zstd, 3)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	t.Logf("Zstd: Original=%d, Compressed=%d, Ratio=%.2f, Savings=%.1f%%",
		len(original), len(compressed),
		CompressionRatio(len(original), len(compressed)),
		SpaceSavings(len(original), len(compressed)))

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data doesn't match original")
	}
}

func TestCompressDecompress_Snappy(t *testing.T) {
	original := []byte(strings.Repeat("Snappy is designed for speed! ", 50))

	compressed, err := Compress(original, Snappy, 0)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	t.Logf("Snappy: Original=%d, Compressed=%d, Ratio=%.2f",
		len(original), len(compressed),
		CompressionRatio(len(original), len(compressed)))

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data doesn't match original")
	}
}

func TestCompressDecompress_None(t *testing.T) {
	original := []byte("No compression test")

	compressed, err := Compress(original, None, 0)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	// With None algorithm, data has header but is uncompressed
	if len(compressed) != len(original)+2 {
		t.Errorf("None compression should add 2-byte header, got %d bytes for %d original",
			len(compressed), len(original))
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data doesn't match original")
	}
}

func TestCompress_EmptyData(t *testing.T) {
	original := []byte{}

	for _, algo := range []Algorithm{None, Gzip, Zstd, Snappy} {
		compressed, err := Compress(original, algo, 0)
		if err != nil {
			t.Errorf("Compress with %s failed for empty data: %v", algo, err)
			continue
		}

		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Errorf("Decompress with %s failed for empty data: %v", algo, err)
			continue
		}

		if !bytes.Equal(original, decompressed) {
			t.Errorf("Empty data not preserved with %s", algo)
		}
	}
}

func TestCompress_LargeData(t *testing.T) {
	// Create 1MB of repetitive data (should compress well)
	original := make([]byte, 1024*1024)
	for i := range original {
		original[i] = byte(i % 256)
	}

	for _, algo := range []Algorithm{Gzip, Zstd, Snappy} {
		t.Run(algo.String(), func(t *testing.T) {
			compressed, err := Compress(original, algo, 6)
			if err != nil {
				t.Fatalf("Compress failed: %v", err)
			}

			ratio := CompressionRatio(len(original), len(compressed))
			savings := SpaceSavings(len(original), len(compressed))

			t.Logf("%s: Original=%d bytes, Compressed=%d bytes, Ratio=%.2fx, Savings=%.1f%%",
				algo, len(original), len(compressed), ratio, savings)

			if ratio < 1.5 {
				t.Logf("Warning: Low compression ratio (%.2f) for %s", ratio, algo)
			}

			decompressed, err := Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress failed: %v", err)
			}

			if !bytes.Equal(original, decompressed) {
				t.Errorf("Large data not preserved correctly")
			}
		})
	}
}

func TestCompress_RandomData(t *testing.T) {
	// Random data should not compress well
	original := make([]byte, 10*1024) // 10KB
	if _, err := rand.Read(original); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	for _, algo := range []Algorithm{Gzip, Zstd, Snappy} {
		compressed, err := Compress(original, algo, 6)
		if err != nil {
			t.Fatalf("Compress with %s failed: %v", algo, err)
		}

		ratio := CompressionRatio(len(original), len(compressed))
		t.Logf("%s on random data: Ratio=%.2f (expected ~1.0 for incompressible data)", algo, ratio)

		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Fatalf("Decompress with %s failed: %v", algo, err)
		}

		if !bytes.Equal(original, decompressed) {
			t.Errorf("Random data not preserved with %s", algo)
		}
	}
}

func TestCompress_DifferentLevels(t *testing.T) {
	original := []byte(strings.Repeat("Compression level test data! ", 1000))

	levels := []int{gzip.BestSpeed, gzip.DefaultCompression, gzip.BestCompression}

	for _, level := range levels {
		compressed, err := Compress(original, Gzip, level)
		if err != nil {
			t.Fatalf("Compress with level %d failed: %v", level, err)
		}

		t.Logf("Gzip level %d: %d bytes (%.1f%% savings)",
			level, len(compressed), SpaceSavings(len(original), len(compressed)))

		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Fatalf("Decompress failed: %v", err)
		}

		if !bytes.Equal(original, decompressed) {
			t.Errorf("Data corrupted at level %d", level)
		}
	}
}

func TestDecompress_InvalidData(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"corrupted_gzip", []byte{byte(Gzip), 6, 0xFF, 0xFF, 0xFF}},
		{"corrupted_zstd", []byte{byte(Zstd), 3, 0xFF, 0xFF, 0xFF}},
		{"corrupted_snappy", []byte{byte(Snappy), 0, 0xFF, 0xFF, 0xFF}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decompress(tc.data)
			if err == nil {
				t.Error("Expected error for corrupted data")
			} else {
				t.Logf("Correctly detected corruption: %v", err)
			}
		})
	}
}

func TestAlgorithmString(t *testing.T) {
	tests := []struct {
		algo Algorithm
		want string
	}{
		{None, "none"},
		{Gzip, "gzip"},
		{Zstd, "zstd"},
		{Snappy, "snappy"},
		{Algorithm(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.algo.String()
		if got != tt.want {
			t.Errorf("Algorithm(%d).String() = %q, want %q", tt.algo, got, tt.want)
		}
	}
}

func TestCompressionMetrics(t *testing.T) {
	originalSize := 1000
	compressedSize := 250

	ratio := CompressionRatio(originalSize, compressedSize)
	if ratio != 4.0 {
		t.Errorf("CompressionRatio = %.2f, want 4.0", ratio)
	}

	savings := SpaceSavings(originalSize, compressedSize)
	if savings != 75.0 {
		t.Errorf("SpaceSavings = %.1f%%, want 75.0%%", savings)
	}

	// Edge cases
	if CompressionRatio(100, 0) != 0 {
		t.Error("CompressionRatio should be 0 when compressed size is 0")
	}

	if SpaceSavings(0, 100) != 0 {
		t.Error("SpaceSavings should be 0 when original size is 0")
	}
}

func TestCompressDecompress_AllAlgorithms(t *testing.T) {
	testData := []byte(strings.Repeat("Test data for all algorithms! ", 100))

	algorithms := []Algorithm{None, Gzip, Zstd, Snappy}

	for _, algo := range algorithms {
		t.Run(algo.String(), func(t *testing.T) {
			compressed, err := Compress(testData, algo, 6)
			if err != nil {
				t.Fatalf("Compress failed: %v", err)
			}

			decompressed, err := Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompress failed: %v", err)
			}

			if !bytes.Equal(testData, decompressed) {
				t.Error("Data mismatch after compress/decompress cycle")
			}

			if algo != None {
				t.Logf("%s: %d -> %d bytes (%.1f%% savings)",
					algo, len(testData), len(compressed),
					SpaceSavings(len(testData), len(compressed)))
			}
		})
	}
}

// Benchmark tests
func BenchmarkCompress_Gzip(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark data for compression! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress(data, Gzip, gzip.DefaultCompression)
	}
}

func BenchmarkCompress_Zstd(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark data for compression! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress(data, Zstd, 3)
	}
}

func BenchmarkCompress_Snappy(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark data for compression! ", 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress(data, Snappy, 0)
	}
}

func BenchmarkDecompress_Gzip(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark data for compression! ", 1000))
	compressed, _ := Compress(data, Gzip, gzip.DefaultCompression)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decompress(compressed)
	}
}

func BenchmarkDecompress_Zstd(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark data for compression! ", 1000))
	compressed, _ := Compress(data, Zstd, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decompress(compressed)
	}
}

func BenchmarkDecompress_Snappy(b *testing.B) {
	data := []byte(strings.Repeat("Benchmark data for compression! ", 1000))
	compressed, _ := Compress(data, Snappy, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decompress(compressed)
	}
}
