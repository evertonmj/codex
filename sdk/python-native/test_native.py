"""Executable C ABI example and integration checks, using only Python stdlib."""
import ctypes
import json
import pathlib
import sys
import tempfile
import unittest

library = ctypes.CDLL(str(pathlib.Path(sys.argv.pop(1)).resolve()))
library.CodexCall.argtypes = [ctypes.c_char_p]
# Keep the pointer: c_char_p restype would lose the address needed by CodexFree.
library.CodexCall.restype = ctypes.c_void_p
library.CodexFree.argtypes = [ctypes.c_void_p]
library.CodexFree.restype = None


def call(request):
    pointer = library.CodexCall(json.dumps(request, ensure_ascii=False).encode())
    try:
        return json.loads(ctypes.string_at(pointer).decode())
    finally:
        library.CodexFree(pointer)


class NativeTest(unittest.TestCase):
    def test_c_abi(self):
        with tempfile.TemporaryDirectory() as directory:
            file = str(pathlib.Path(directory) / 'db.json')
            opened = call({'op': 'open', 'file': file})
            self.assertTrue(opened['ok'])
            handle = opened['result']
            try:
                self.assertFalse(call({'op': 'open', 'file': file})['ok'])
                value = {'name': 'João', 'number': 9007199254740993}
                self.assertTrue(call({'op': 'set', 'handle': handle, 'key': 'x', 'value': value})['ok'])
                self.assertEqual(call({'op': 'get', 'handle': handle, 'key': 'x'})['result'], value)
                self.assertEqual(call({'op': 'keys', 'handle': handle})['result'], ['x'])
                self.assertTrue(call({'op': 'has', 'handle': handle, 'key': 'x'})['result'])
            finally:
                self.assertTrue(call({'op': 'close', 'handle': handle})['ok'])
            handle = call({'op': 'open', 'file': file})['result']
            try:
                self.assertEqual(call({'op': 'get', 'handle': handle, 'key': 'x'})['result'], value)
                self.assertTrue(call({'op': 'delete', 'handle': handle, 'key': 'x'})['ok'])
                self.assertEqual(call({'op': 'get', 'handle': handle, 'key': 'x'})['code'], 'NOT_FOUND')
                self.assertTrue(call({'op': 'clear', 'handle': handle})['ok'])
            finally:
                call({'op': 'close', 'handle': handle})
            self.assertFalse(call({'op': 'get', 'handle': handle})['ok'])

    def test_null_pointer(self):
        pointer = library.CodexCall(None)
        try:
            self.assertFalse(json.loads(ctypes.string_at(pointer))['ok'])
        finally:
            library.CodexFree(pointer)
        library.CodexFree(None)


if __name__ == '__main__':
    unittest.main()
