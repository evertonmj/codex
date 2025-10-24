package main

import (
	"os"
	"testing"
)

func TestCompressionExample(t *testing.T) {
	// Clean up any existing files
	defer func() {
		os.Remove("compression_none.db")
		os.Remove("compression_gzip.db")
		os.Remove("compression_zstd.db")
		os.Remove("compression_snappy.db")
		os.Remove("compression_encrypted.db")
		os.Remove("compression_multi.db")
		os.Remove("compression_batch.db")
	}()

	// Run the main function
	main()

	// Test passes if main() completes without panic
}
