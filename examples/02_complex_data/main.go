package main

import (
	"fmt"
	"log"
	"time"

	"go-file-persistence/codex"
)

type User struct {
	ID        int
	Username  string
	Email     string
	CreatedAt time.Time
	Tags      []string
	Settings  map[string]interface{}
}

type Post struct {
	ID        int
	Title     string
	Content   string
	Author    string
	Comments  []string
	Published bool
}

func main() {
	fmt.Println("=== Complex Data Types Example ===")

	store, err := codex.New("complex_data.db")
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Store complex structs
	fmt.Println("1. Storing complex user object...")
	user := User{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		Tags:      []string{"developer", "golang", "database"},
		Settings: map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"language":      "en",
		},
	}

	if err := store.Set("user:1", user); err != nil {
		log.Fatalf("Failed to set user: %v", err)
	}
	fmt.Println("   User stored successfully")

	// Retrieve the user
	fmt.Println("\n2. Retrieving user object...")
	var retrievedUser User
	if err := store.Get("user:1", &retrievedUser); err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	fmt.Printf("   Username: %s\n", retrievedUser.Username)
	fmt.Printf("   Email: %s\n", retrievedUser.Email)
	fmt.Printf("   Tags: %v\n", retrievedUser.Tags)
	fmt.Printf("   Theme: %v\n", retrievedUser.Settings["theme"])

	// Store blog posts
	fmt.Println("\n3. Storing multiple blog posts...")
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
		key := fmt.Sprintf("post:%d", post.ID)
		store.Set(key, post)
	}
	fmt.Printf("   Stored %d posts\n", len(posts))

	// Store nested data structures
	fmt.Println("\n4. Storing nested data structures...")
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

	store.Set("app:config", config)

	var retrievedConfig map[string]interface{}
	store.Get("app:config", &retrievedConfig)
	fmt.Printf("   Config keys: %v\n", getMapKeys(retrievedConfig))

	// Store arrays and slices
	fmt.Println("\n5. Storing collections...")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	store.Set("numbers", numbers)

	words := []string{"hello", "world", "golang", "rocks"}
	store.Set("words", words)

	var retrievedNumbers []interface{}
	store.Get("numbers", &retrievedNumbers)
	fmt.Printf("   Numbers count: %d\n", len(retrievedNumbers))

	fmt.Println("\n=== Example Complete ===")
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
