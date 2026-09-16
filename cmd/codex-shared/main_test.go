package main

import "testing"

func TestNullRequestAllocation(t *testing.T) {
	response := CodexCall(nil)
	if response == nil {
		t.Fatal("null request should return an allocated protocol error")
	}
	CodexFree(response)
	CodexFree(nil)
}
