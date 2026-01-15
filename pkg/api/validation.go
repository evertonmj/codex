package api

import "strings"

// ValidateKey validates that a key is valid
func ValidateKey(key string) bool {
	return strings.TrimSpace(key) != ""
}

// ValidateBatchOperations validates batch operations
func ValidateBatchOperations(ops []BatchOperation) error {
	if len(ops) == 0 {
		return BadRequest("No operations provided")
	}

	for _, op := range ops {
		if op.Op != "set" && op.Op != "delete" {
			return BadRequest("Invalid operation: " + op.Op)
		}

		if !ValidateKey(op.Key) {
			return BadRequest("Key cannot be empty")
		}

		if op.Op == "set" && len(op.Value) == 0 {
			return BadRequest("Value required for set operation")
		}
	}

	return nil
}

// ValidateBatchGetKeys validates batch get keys
func ValidateBatchGetKeys(keys []string) error {
	if len(keys) == 0 {
		return BadRequest("No keys provided")
	}

	for _, key := range keys {
		if !ValidateKey(key) {
			return BadRequest("Key cannot be empty")
		}
	}

	return nil
}
