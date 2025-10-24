package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go-file-persistence/codex"
)

func TestConcurrentAccessExample(t *testing.T) {
	// Clean up
	defer os.Remove("concurrent.db")

	store, err := codex.New("concurrent.db")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Test 1: Multiple concurrent writers
	t.Run("multiple_writers", func(t *testing.T) {
		var wg sync.WaitGroup
		numWriters := 5
		writesPerWorker := 10

		for i := 0; i < numWriters; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < writesPerWorker; j++ {
					key := fmt.Sprintf("worker_%d_item_%d", workerID, j)
					value := fmt.Sprintf("data_%d_%d", workerID, j)
					if err := store.Set(key, value); err != nil {
						t.Errorf("Worker %d failed to set: %v", workerID, err)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify all writes succeeded
		keys := store.Keys()
		expectedWrites := numWriters * writesPerWorker
		if len(keys) < expectedWrites {
			t.Errorf("Expected at least %d keys, got %d", expectedWrites, len(keys))
		}
	})

	// Test 2: Concurrent readers and writers
	t.Run("mixed_operations", func(t *testing.T) {
		var wg sync.WaitGroup
		numReaders := 3
		numWriters := 2
		opsPerWorker := 20

		// Readers
		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func(readerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					_ = store.Keys()
					time.Sleep(time.Microsecond * 100)
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
					time.Sleep(time.Microsecond * 100)
				}
			}(i)
		}

		wg.Wait()
	})

	// Test 3: Counter with concurrent increments
	t.Run("concurrent_counter", func(t *testing.T) {
		store.Set("counter", 0)

		var wg sync.WaitGroup
		var mu sync.Mutex
		numIncrementers := 10
		incrementsPerWorker := 100

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

		var finalCounter int
		if err := store.Get("counter", &finalCounter); err != nil {
			t.Fatalf("Failed to get counter: %v", err)
		}

		expected := numIncrementers * incrementsPerWorker
		if finalCounter != expected {
			t.Errorf("Expected counter %d, got %d", expected, finalCounter)
		}
	})

	// Test 4: Producer-Consumer pattern
	t.Run("producer_consumer", func(t *testing.T) {
		var wg sync.WaitGroup
		const numMessages = 20

		// Producer
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numMessages; i++ {
				key := fmt.Sprintf("message_%d", i)
				message := fmt.Sprintf("Hello from producer, message #%d", i)
				store.Set(key, message)
				store.Set("latest_message_id", i)
				time.Sleep(time.Microsecond * 100)
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
					time.Sleep(time.Microsecond * 50)
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
				time.Sleep(time.Microsecond * 50)
			}
		}()

		wg.Wait()

		// Verify all messages were produced
		for i := 0; i < numMessages; i++ {
			key := fmt.Sprintf("message_%d", i)
			if !store.Has(key) {
				t.Errorf("Expected message_%d to exist", i)
			}
		}
	})
}
