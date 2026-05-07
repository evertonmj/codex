package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	codex "github.com/evertonmj/codex/app"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type options struct {
	file             string
	homeName         string
	ledger           bool
	encryptionKey    string
	numBackups       int
	compression      string
	compressionLevel int
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	var opts options

	fs := flag.NewFlagSet("codexdb", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.StringVar(&opts.file, "file", "", "database file path")
	fs.StringVar(&opts.homeName, "home", "", "create/open database in ~/.codex/ with the given name")
	fs.BoolVar(&opts.ledger, "ledger", false, "enable append-only ledger mode")
	fs.StringVar(&opts.encryptionKey, "encrypt-key", "", "encryption key (16/24/32 bytes)")
	fs.IntVar(&opts.numBackups, "backups", 0, "number of rotating backups to keep (snapshot mode only)")
	fs.StringVar(&opts.compression, "compression", "none", "compression: none|gzip|zstd|snappy")
	fs.IntVar(&opts.compressionLevel, "compression-level", 0, "compression level (gzip/zstd only; 0 = default)")

	showVersion := fs.Bool("version", false, "print version and exit")
	fs.BoolVar(showVersion, "v", false, "print version and exit (shorthand)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Printf("codexdb %s (commit %s, built %s)\n", version, commit, date)
		return 0
	}

	rest := fs.Args()
	if len(rest) == 0 {
		usage(os.Stderr)
		return 2
	}

	cmd := rest[0]
	cmdArgs := rest[1:]

	store, closeFn, err := openStore(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer closeFn()

	switch cmd {
	case "set":
		if len(cmdArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db set <key> <json-or-string>")
			return 2
		}
		key := cmdArgs[0]
		valStr := strings.Join(cmdArgs[1:], " ")
		val := parseValue(valStr)
		if err := store.Set(key, val); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("OK")
		return 0

	case "get":
		if len(cmdArgs) != 1 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db get <key>")
			return 2
		}
		var v interface{}
		if err := store.Get(cmdArgs[0], &v); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		out, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(out))
		return 0

	case "has":
		if len(cmdArgs) != 1 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db has <key>")
			return 2
		}
		fmt.Println(store.Has(cmdArgs[0]))
		return 0

	case "delete", "del", "rm":
		if len(cmdArgs) != 1 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db delete <key>")
			return 2
		}
		if err := store.Delete(cmdArgs[0]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("OK")
		return 0

	case "keys":
		if len(cmdArgs) != 0 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db keys")
			return 2
		}
		for _, k := range store.Keys() {
			fmt.Println(k)
		}
		return 0

	case "clear":
		if len(cmdArgs) != 0 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db clear")
			return 2
		}
		if err := store.Clear(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("OK")
		return 0

	case "interactive", "repl":
		if len(cmdArgs) != 0 {
			fmt.Fprintln(os.Stderr, "usage: codexdb --file=db interactive")
			return 2
		}
		return repl(store)

	default:
		fmt.Fprintln(os.Stderr, "unknown command:", cmd)
		usage(os.Stderr)
		return 2
	}
}

func usage(w *os.File) {
	fmt.Fprintln(w, "CodexDB CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  codexdb [global flags] <command> [args]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Global flags:")
	fmt.Fprintln(w, "  --file PATH                 database file path")
	fmt.Fprintln(w, "  --home NAME                 use ~/.codex/ (auto-generated filename)")
	fmt.Fprintln(w, "  --ledger                    enable ledger mode")
	fmt.Fprintln(w, "  --encrypt-key KEY           encryption key (16/24/32 bytes)")
	fmt.Fprintln(w, "  --backups N                 rotating backups (snapshot mode only)")
	fmt.Fprintln(w, "  --compression ALGO          none|gzip|zstd|snappy")
	fmt.Fprintln(w, "  --compression-level LEVEL   gzip/zstd level (0 = default)")
	fmt.Fprintln(w, "  --version, -v               print version and exit")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  set <key> <json-or-string>")
	fmt.Fprintln(w, "  get <key>")
	fmt.Fprintln(w, "  has <key>")
	fmt.Fprintln(w, "  delete <key>")
	fmt.Fprintln(w, "  keys")
	fmt.Fprintln(w, "  clear")
	fmt.Fprintln(w, "  interactive")
}

func openStore(opts options) (*codex.Store, func(), error) {
	co, err := toCodexOptions(opts)
	if err != nil {
		return nil, nil, err
	}

	if opts.homeName != "" {
		s, err := codex.NewHomeWithOptions(opts.homeName, co)
		if err != nil {
			return nil, nil, err
		}
		return s, func() { _ = s.Close() }, nil
	}

	if opts.file == "" {
		return nil, nil, fmt.Errorf("missing --file (or use --home)")
	}

	s, err := codex.NewWithOptions(opts.file, co)
	if err != nil {
		return nil, nil, err
	}
	return s, func() { _ = s.Close() }, nil
}

func toCodexOptions(opts options) (codex.Options, error) {
	var co codex.Options
	co.LedgerMode = opts.ledger
	co.NumBackups = opts.numBackups

	if opts.encryptionKey != "" {
		co.EncryptionKey = []byte(opts.encryptionKey)
	}

	switch strings.ToLower(strings.TrimSpace(opts.compression)) {
	case "", "none", "no", "off":
		co.Compression = codex.NoCompression
	case "gzip":
		co.Compression = codex.GzipCompression
	case "zstd":
		co.Compression = codex.ZstdCompression
	case "snappy":
		co.Compression = codex.SnappyCompression
	default:
		return codex.Options{}, fmt.Errorf("invalid --compression %q", opts.compression)
	}

	co.CompressionLevel = opts.compressionLevel
	return co, nil
}

func parseValue(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Accept common literals without needing JSON quoting.
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && strings.ContainsAny(s, ".eE") {
		return f
	}

	// Try JSON (object/array/quoted-string/number/bool/null).
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err == nil {
		return v
	}

	// Fallback: store as plain string.
	return s
}

func repl(store *codex.Store) int {
	in := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	fmt.Fprintln(out, "CodexDB interactive mode. Type 'help' for commands. 'exit' to quit.")
	_ = out.Flush()

	for {
		fmt.Fprint(out, "codexdb> ")
		_ = out.Flush()

		if !in.Scan() {
			return 0
		}
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}

		fields := splitArgs(line)
		if len(fields) == 0 {
			continue
		}
		cmd := fields[0]
		args := fields[1:]

		switch cmd {
		case "exit", "quit":
			return 0
		case "help":
			fmt.Fprintln(out, "Commands: set/get/has/delete/keys/clear/version/help/exit")
		case "version":
			fmt.Fprintf(out, "codexdb %s (commit %s, built %s)\n", version, commit, date)
		case "set":
			if len(args) < 2 {
				fmt.Fprintln(out, "usage: set <key> <json-or-string>")
				continue
			}
			key := args[0]
			valStr := strings.Join(args[1:], " ")
			if err := store.Set(key, parseValue(valStr)); err != nil {
				fmt.Fprintln(out, "error:", err)
				continue
			}
			fmt.Fprintln(out, "OK")
		case "get":
			if len(args) != 1 {
				fmt.Fprintln(out, "usage: get <key>")
				continue
			}
			var v interface{}
			if err := store.Get(args[0], &v); err != nil {
				fmt.Fprintln(out, "error:", err)
				continue
			}
			b, _ := json.MarshalIndent(v, "", "  ")
			fmt.Fprintln(out, string(b))
		case "has":
			if len(args) != 1 {
				fmt.Fprintln(out, "usage: has <key>")
				continue
			}
			fmt.Fprintln(out, store.Has(args[0]))
		case "delete", "del", "rm":
			if len(args) != 1 {
				fmt.Fprintln(out, "usage: delete <key>")
				continue
			}
			if err := store.Delete(args[0]); err != nil {
				fmt.Fprintln(out, "error:", err)
				continue
			}
			fmt.Fprintln(out, "OK")
		case "keys":
			if len(args) != 0 {
				fmt.Fprintln(out, "usage: keys")
				continue
			}
			for _, k := range store.Keys() {
				fmt.Fprintln(out, k)
			}
		case "clear":
			if len(args) != 0 {
				fmt.Fprintln(out, "usage: clear")
				continue
			}
			if err := store.Clear(); err != nil {
				fmt.Fprintln(out, "error:", err)
				continue
			}
			fmt.Fprintln(out, "OK")
		default:
			fmt.Fprintln(out, "unknown command:", cmd)
		}

		_ = out.Flush()
	}
}

func splitArgs(line string) []string {
	// Minimal splitter: supports double-quoted strings.
	// Example: set k "{\"a\":1}"  => args: ["set","k","{\"a\":1}"]
	var out []string
	var cur strings.Builder
	inQuotes := false
	escaped := false

	for _, r := range line {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			if inQuotes {
				escaped = true
			} else {
				cur.WriteRune(r)
			}
		case r == '"':
			inQuotes = !inQuotes
		case !inQuotes && (r == ' ' || r == '\t'):
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func init() {
	// Provide a reasonable default date for `go run`.
	if date == "unknown" {
		date = time.Now().UTC().Format(time.RFC3339)
	}
}
