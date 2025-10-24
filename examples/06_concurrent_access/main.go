package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"go-file-persistence/codex"
)

func main() {
	fmt.Println("=== Concurrent Access Example ===")

	store, err := codex.New("concurrent.db")
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Example 1: Multiple writers
	fmt.Println("1. Multiple concurrent writers...")
	var wg sync.WaitGroup
	numWriters := 5
	writesPerWorker := 10

	startTime := time.Now()
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < writesPerWorker; j++ {
				key := fmt.Sprintf("worker_%d_item_%d", workerID, j)
				value := fmt.Sprintf("data_%d_%d", workerID, j)
				store.Set(key, value)
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("   %d writes completed in %v\n", numWriters*writesPerWorker, elapsed)

	// Example 2: Concurrent readers and writers
	fmt.Println("\n2. Concurrent readers and writers...")
	numReaders := 3
	numWriters = 2
	opsPerWorker := 20

	startTime = time.Now()

	// Readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				keys := store.Keys()
				_ = keys
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", writerID, j)
				store.Set(key, j)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	elapsed = time.Since(startTime)
	fmt.Printf("   Mixed operations completed in %v\n", elapsed)

	// Example 3: Counter with concurrent increments
	fmt.Println("\n3. Concurrent counter increments...")
	store.Set("counter", 0)

	numIncrementers := 10
	incrementsPerWorker := 100

	startTime = time.Now()
	var mu sync.Mutex

	for i := 0; i < numIncrementers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerWorker; j++ {
				mu.Lock()
				var counter int
				store.Get("counter", &counter)
				counter++
				store.Set("counter", counter)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	elapsed = time.Since(startTime)

	var finalCounter int
	store.Get("counter", &finalCounter)
	expected := numIncrementers * incrementsPerWorker
	fmt.Printf("   Expected: %d, Got: %d, Time: %v\n", expected, finalCounter, elapsed)

	if finalCounter == expected {
		fmt.Println("   ✓ Counter is correct!")
	} else {
		fmt.Println("   ✗ Counter mismatch!")
	}

	// Example 4: Producer-Consumer pattern
	fmt.Println("\n4. Producer-Consumer pattern...")
	const numMessages = 50

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numMessages; i++ {
			key := fmt.Sprintf("message_%d", i)
			message := fmt.Sprintf("Hello from producer, message #%d", i)
			store.Set(key, message)
			store.Set("latest_message_id", i)
			time.Sleep(time.Millisecond * 10)
		}
	}()

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		processedCount := 0
		lastID := -1

		for processedCount < numMessages {
			var latestID int
			if err := store.Get("latest_message_id", &latestID); err != nil {
				time.Sleep(time.Millisecond * 5)
				continue
			}

			if latestID > lastID {
				key := fmt.Sprintf("message_%d", latestID)
				var message string
				if err := store.Get(key, &message); err == nil {
					processedCount++
					lastID = latestID
				}
			}
			time.Sleep(time.Millisecond * 5)
		}
	}()

	wg.Wait()
	fmt.Printf("   ✓ Produced and consumed %d messages\n", numMessages)

	// Final stats
	fmt.Println("\n5. Final statistics:")
	keys := store.Keys()
	fmt.Printf("   Total keys in store: %d\n", len(keys))

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nNote: Codex handles concurrent access safely with internal locking.")
}
