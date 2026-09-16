package batch

import "testing"

func TestSerializeUnsupportedValue(t *testing.T) {
	b := New()
	b.Set("key", func() {})
	if _, err := b.Serialize(); err == nil {
		t.Fatal("unsupported value accepted")
	}
}
