package native

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func invoke(t *testing.T, e *Engine, request Request, success bool) Response {
	t.Helper()
	input, _ := json.Marshal(request)
	var response Response
	if err := json.Unmarshal([]byte(e.Call(string(input))), &response); err != nil {
		t.Fatal(err)
	}
	if response.OK != success {
		t.Fatalf("%+v: %+v", request, response)
	}
	return response
}
func TestEngine(t *testing.T) {
	for _, ledger := range []bool{false, true} {
		t.Run(map[bool]string{false: "snapshot", true: "ledger"}[ledger], func(t *testing.T) {
			var e Engine
			file := filepath.Join(t.TempDir(), "data.json")
			h := invoke(t, &e, Request{Op: "open", File: file, Ledger: ledger}, true).Result.(string)
			invoke(t, &e, Request{Op: "open", File: file, Ledger: ledger}, false)
			invoke(t, &e, Request{Op: "set", Handle: h, Key: "child", Value: json.RawMessage(`{"name":"João","count":9007199254740993}`)}, true)
			got := invoke(t, &e, Request{Op: "get", Handle: h, Key: "child"}, true)
			if got.Result.(map[string]interface{})["name"] != "João" {
				t.Fatal(got)
			}
			invoke(t, &e, Request{Op: "has", Handle: h, Key: "child"}, true)
			invoke(t, &e, Request{Op: "keys", Handle: h}, true)
			invoke(t, &e, Request{Op: "delete", Handle: h, Key: "child"}, true)
			if r := invoke(t, &e, Request{Op: "get", Handle: h, Key: "child"}, false); r.Code != "NOT_FOUND" {
				t.Fatal(r)
			}
			invoke(t, &e, Request{Op: "set", Handle: h}, false)
			invoke(t, &e, Request{Op: "set", Handle: h, Key: "null", Value: json.RawMessage(`null`)}, true)
			invoke(t, &e, Request{Op: "unknown", Handle: h}, false)
			invoke(t, &e, Request{Op: "clear", Handle: h}, true)
			invoke(t, &e, Request{Op: "close", Handle: h}, true)
			invoke(t, &e, Request{Op: "get", Handle: h}, false)
		})
	}
}
func TestOpenErrors(t *testing.T) {
	var e Engine
	invoke(t, &e, Request{Op: "open"}, false)
	if r := invoke(t, &e, Request{Op: "open", File: filepath.Join(t.TempDir(), "x"), EncryptionKey: "short"}, false); r.Code != "INVALID_KEY" {
		t.Fatal(r)
	}
	file := filepath.Join(t.TempDir(), "encrypted.json")
	h := invoke(t, &e, Request{Op: "open", File: file, EncryptionKey: "01234567890123456789012345678901"}, true).Result.(string)
	invoke(t, &e, Request{Op: "set", Handle: h, Value: json.RawMessage(`"secret"`)}, true)
	invoke(t, &e, Request{Op: "close", Handle: h}, true)
	invoke(t, &e, Request{Op: "open", File: file}, false)
	invoke(t, &e, Request{Op: "open", File: file, EncryptionKey: "11234567890123456789012345678901"}, false)
	h = invoke(t, &e, Request{Op: "open", File: file, EncryptionKey: "01234567890123456789012345678901"}, true).Result.(string)
	invoke(t, &e, Request{Op: "close", Handle: h}, true)
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	invoke(t, &e, Request{Op: "open", File: filepath.Join(blocker, "db")}, false)
	var r Response
	json.Unmarshal([]byte(e.Call("{bad")), &r)
	if r.OK {
		t.Fatal(r)
	}
}

func TestEncryptedLedgerReopen(t *testing.T) {
	var e Engine
	file := filepath.Join(t.TempDir(), "ledger.db")
	key := "01234567890123456789012345678901"
	h := invoke(t, &e, Request{Op: "open", File: file, Ledger: true, EncryptionKey: key}, true).Result.(string)
	invoke(t, &e, Request{Op: "set", Handle: h, Key: "a", Value: json.RawMessage(`"secret"`)}, true)
	invoke(t, &e, Request{Op: "close", Handle: h}, true)
	before, _ := os.ReadFile(file)
	invoke(t, &e, Request{Op: "open", File: file, Ledger: true}, false)
	invoke(t, &e, Request{Op: "open", File: file, Ledger: true, EncryptionKey: "11234567890123456789012345678901"}, false)
	after, _ := os.ReadFile(file)
	if string(before) != string(after) {
		t.Fatal("failed open changed the encrypted file")
	}
	h = invoke(t, &e, Request{Op: "open", File: file, Ledger: true, EncryptionKey: key}, true).Result.(string)
	if r := invoke(t, &e, Request{Op: "get", Handle: h, Key: "a"}, true); r.Result != "secret" {
		t.Fatal(r)
	}
	invoke(t, &e, Request{Op: "close", Handle: h}, true)
}
