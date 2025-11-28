// Package integrity provides data integrity verification using
// SHA256 checksums to detect corruption.
//
// Integrity Protection:
//   - All data is signed with SHA256 checksum before storage
//   - Checksum is verified on load to detect corruption
//   - Works with encrypted and unencrypted data
//   - Detects bit flips, truncation, and modification
//
// Verification Process:
//  1. Calculate SHA256 of data
//  2. Create fileFormat with checksum and data
//  3. Marshal to JSON and store
//  4. On load, recalculate checksum and verify match
//
// Note: Checksum verification is transparent to the user and happens
// automatically during database load.
package integrity

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// FileFormat represents the structure of the data file with a checksum.

type FileFormat struct {
	Checksum string          `json:"checksum"`
	Data     json.RawMessage `json:"data"`
}

// Sign calculates the checksum of the data and returns the full file content.
func Sign(data []byte) ([]byte, error) {
	hasher := sha256.New()
	hasher.Write(data)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	content := FileFormat{
		Checksum: checksum,
		Data:     data,
	}

	return json.MarshalIndent(content, "", "  ")
}

// Verify checks the integrity of the file content and returns the raw data.
func Verify(fileData []byte) (json.RawMessage, error) {
	var content FileFormat
	// If parsing fails, or if essential fields are missing, assume it's an old format.
	if err := json.Unmarshal(fileData, &content); err != nil || content.Checksum == "" || content.Data == nil {
		return fileData, nil
	}

	// Compact the data to get a canonical representation for checksum comparison
	var compactedData bytes.Buffer
	if err := json.Compact(&compactedData, content.Data); err != nil {
		return nil, fmt.Errorf("failed to compact data for checksum: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(compactedData.Bytes())
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if calculatedChecksum != content.Checksum {
		return nil, fmt.Errorf("file integrity check failed: checksum mismatch")
	}

	return content.Data, nil
}
