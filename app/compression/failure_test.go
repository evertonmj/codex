package compression

import "testing"

func TestCompressionFailurePaths(t *testing.T) {
	for _, level := range []int{-5, 20} {
		for _, algo := range []Algorithm{Gzip, Zstd} {
			data, err := Compress([]byte("payload"), algo, level)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Decompress(data); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, algo := range []Algorithm{Gzip, Zstd, Snappy} {
		if _, err := Decompress([]byte{byte(algo), 0, 42}); err == nil {
			t.Fatal("bad data accepted")
		}
	}
	data, _ := Compress([]byte("payload"), Gzip, 3)
	data[len(data)-1] ^= 255
	if _, err := Decompress(data); err == nil {
		t.Fatal("gzip corruption accepted")
	}
}
func TestUnknownDecompressionHeader(t *testing.T) {
	if _, err := Decompress([]byte{255, 0}); err == nil {
		t.Fatal("unsupported header accepted")
	}
}
