package integrity

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	t.Run("it signs and verifies correctly", func(t *testing.T) {
		rawData := []byte(`{"key":"value","number":123}`)

		signedData, err := Sign(rawData)
		if err != nil {
			t.Fatalf("Sign() failed: %v", err)
		}

		verifiedData, err := Verify(signedData)
		if err != nil {
			t.Fatalf("Verify() failed: %v", err)
		}

		// Unmarshal both to maps to compare content, ignoring formatting
		var expected, actual map[string]interface{}
		if err := json.Unmarshal(rawData, &expected); err != nil {
			t.Fatalf("failed to unmarshal rawData: %v", err)
		}
		if err := json.Unmarshal(verifiedData, &actual); err != nil {
			t.Fatalf("failed to unmarshal verifiedData: %v", err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("mismatch: expected %v, got %v", expected, actual)
		}
	})

	t.Run("it returns error on corrupted data", func(t *testing.T) {
		// Manually create a corrupted file format
		corruptedContent := FileFormat{
			Checksum: "invalidchecksum",
			Data:     json.RawMessage(`{"key":"value"}`),
		}
		corruptedData, _ := json.Marshal(corruptedContent)

		_, err := Verify(corruptedData)
		if err == nil {
			t.Fatal("Verify() did not return error for corrupted data")
		}
		if err.Error() != "file integrity check failed: checksum mismatch" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	t.Run("it handles old format without checksum", func(t *testing.T) {
		oldFormatData := []byte(`{"key":"value"}`)

		verifiedData, err := Verify(oldFormatData)
		if err != nil {
			t.Fatalf("Verify() returned an unexpected error for old format: %v", err)
		}

		if string(verifiedData) != string(oldFormatData) {
			t.Errorf("mismatch for old format: expected %s, got %s", string(oldFormatData), string(verifiedData))
		}
	})
}
