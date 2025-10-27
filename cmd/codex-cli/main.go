package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"go-file-persistence/codex"
)

func main() {
	// Define global flags
	filePath := flag.String("file", "", "Path to the database file.")
	useHome := flag.Bool("home", false, "Create database in home directory (~/.codex/). Use with optional database name.")
	dbName := flag.String("name", "", "Database name (used with --home flag). Format: NAME_TIMESTAMP_HASH.db")
	ledgerMode := flag.Bool("ledger", false, "Enable append-only ledger mode.")

	flag.Parse()

	// Get command and arguments
	args := flag.Args()
	if len(args) < 1 {
		fatalf("Usage: codex-cli [--file path | --home [--name dbname]] [--ledger] <command> [args]\nCommands: set, get, delete, keys, has, clear, interactive")
	}

	// Read encryption key from environment variable for security
	keyStr := os.Getenv("CODEX_KEY")
	var keyBytes []byte
	if keyStr != "" {
		keyBytes = []byte(keyStr)
	}

	opts := codex.Options{
		LedgerMode:    *ledgerMode,
		EncryptionKey: keyBytes,
	}

	// Create or open the store
	var store *codex.Store
	var err error

	if *useHome {
		// Use home directory for database
		store, err = codex.NewHomeWithOptions(*dbName, opts)
	} else if *filePath != "" {
		// Use specified file path
		store, err = codex.NewWithOptions(*filePath, opts)
	} else {
		fatalf("Error: must specify either --file <path> or --home [--name <dbname>]")
	}

	if err != nil {
		fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Print database path when using --home
	if *useHome {
		fmt.Fprintf(os.Stderr, "Database: %s\n", store.Path())
	}

	command := args[0]
	cmdArgs := args[1:]

	if command == "interactive" {
		runInteractive(store)
	} else {
		if err := executeCommand(store, command, cmdArgs); err != nil {
			fatalf("%v", err)
		}
	}
}

func runInteractive(store *codex.Store) {
	fmt.Println("CodexDB Interactive Mode. Type 'exit' or 'quit' to leave.")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("codex > ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		if command == "exit" || command == "quit" {
			break
		}

		if err := executeCommand(store, command, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

func executeCommand(store *codex.Store, command string, args []string) error {
	switch command {
	case "set":
		if len(args) < 2 {
			return fmt.Errorf("usage: set <key> <json_value>")
		}
		key := args[0]
		valueStr := strings.Join(args[1:], " ")

		var value interface{}
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			return fmt.Errorf("invalid JSON value: %v", err)
		}
		if err := store.Set(key, value); err != nil {
			return fmt.Errorf("set failed: %v", err)
		}
		fmt.Println("OK")

	case "get":
		if len(args) != 1 {
			return fmt.Errorf("usage: get <key>")
		}
		var value interface{}
		if err := store.Get(args[0], &value); err != nil {
			return fmt.Errorf("get failed: %v", err)
		}
		jsonVal, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %v", err)
		}
		fmt.Println(string(jsonVal))

	case "delete":
		if len(args) != 1 {
			return fmt.Errorf("usage: delete <key>")
		}
		if err := store.Delete(args[0]); err != nil {
			return fmt.Errorf("delete failed: %v", err)
		}
		fmt.Println("OK")

	case "keys":
		if len(args) != 0 {
			return fmt.Errorf("usage: keys")
		}
		keys := store.Keys()
		fmt.Println(strings.Join(keys, "\n"))

	case "has":
		if len(args) != 1 {
			return fmt.Errorf("usage: has <key>")
		}
		if store.Has(args[0]) {
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}

	case "clear":
		if len(args) != 0 {
			return fmt.Errorf("usage: clear")
		}
		if err := store.Clear(); err != nil {
			return fmt.Errorf("clear failed: %v", err)
		}
		fmt.Println("OK")

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}