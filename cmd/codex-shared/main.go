// Build with go build -buildmode=c-shared. No Go pointers cross the C ABI.
package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"github.com/evertonmj/codex/internal/native"
	"unsafe"
)

var engine native.Engine

// CodexCall accepts a NUL-terminated UTF-8 JSON request and returns an owned
// NUL-terminated UTF-8 JSON response. Release it exactly once with CodexFree.
//
//export CodexCall
func CodexCall(request *C.char) *C.char {
	return C.CString(engine.Call(C.GoString(request)))
}

//export CodexFree
func CodexFree(response *C.char) { C.free(unsafe.Pointer(response)) }
func main()                      {}
