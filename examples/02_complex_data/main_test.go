package main

import (
	"os"
	"testing"
	"time"

	"go-file-persistence/codex"
)

func TestComplexDataExample(t *testing.T) {
	// Clean up
	defer os.Remove("complex_data.db")

	store, err := codex.New("complex_data.db")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test storing complex user object
	user := User{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		CreatedAt: time.Now().Round(time.Second),
		Tags:      []string{"developer", "golang", "database"},
		Settings: map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"language":      "en",
		},
	}

	if err := store.Set("user:1", user); err != nil {
		t.Fatalf("Failed to set user: %v", err)
	}

	// Test retrieving user
	var retrievedUser User
	if err := store.Get("user:1", &retrievedUser); err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.Username != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%s'", retrievedUser.Username)
	}
	if retrievedUser.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", retrievedUser.Email)
	}
	if len(retrievedUser.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(retrievedUser.Tags))
	}

	// Test storing blog posts
	posts := []Post{
		{
			ID:        1,
			Title:     "Introduction to Codex",
			Content:   "Codex is a simple key-value store...",
			Author:    "john_doe",
			Comments:  []string{"Great post!", "Very helpful"},
			Published: true,
		},
		{
			ID:        2,
			Title:     "Advanced Features",
			Content:   "Let's explore encryption and ledger mode...",
			Author:    "john_doe",
			Comments:  []string{},
			Published: false,
		},
	}

	for _, post := range posts {
		key := "post:" + string(rune(post.ID+'0'))
		if err := store.Set(key, post); err != nil {
			t.Fatalf("Failed to set post: %v", err)
		}
	}

	// Test nested data structures
	config := map[string]interface{}{
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "admin",
			"pool": map[string]int{
				"min": 5,
				"max": 20,
			},
		},
		"server": map[string]interface{}{
			"port":    8080,
			"timeout": 30,
			"ssl":     true,
		},
	}

	if err := store.Set("app:config", config); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	var retrievedConfig map[string]interface{}
	if err := store.Get("app:config", &retrievedConfig); err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	if len(retrievedConfig) != 2 {
		t.Errorf("Expected 2 config sections, got %d", len(retrievedConfig))
	}

	// Test collections
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if err := store.Set("numbers", numbers); err != nil {
		t.Fatalf("Failed to set numbers: %v", err)
	}

	words := []string{"hello", "world", "golang", "rocks"}
	if err := store.Set("words", words); err != nil {
		t.Fatalf("Failed to set words: %v", err)
	}

	var retrievedNumbers []interface{}
	if err := store.Get("numbers", &retrievedNumbers); err != nil {
		t.Fatalf("Failed to get numbers: %v", err)
	}

	if len(retrievedNumbers) != 10 {
		t.Errorf("Expected 10 numbers, got %d", len(retrievedNumbers))
	}
}
