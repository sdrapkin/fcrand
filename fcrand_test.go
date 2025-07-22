package fcrand

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"strings"
	"testing"
)

// Test that Reader.Read calls Read internally
func TestReaderRead(t *testing.T) {
	buf := make([]byte, 32)
	n, err := Reader.Read(buf)
	if err != nil {
		t.Fatalf("Reader.Read returned error: %v", err)
	}
	if n != len(buf) {
		t.Fatalf("Reader.Read returned n=%d, want %d", n, len(buf))
	}
	if bytes.Equal(buf, make([]byte, 32)) {
		t.Fatal("Reader.Read returned all zero bytes")
	}
}

// Test Read for small buffers (uses sb)
func TestRead_SmallBuffer(t *testing.T) {
	buf := make([]byte, 16)
	n, err := Read(buf)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if n != 16 {
		t.Fatalf("Read returned n=%d, want 16", n)
	}
	if bytes.Equal(buf, make([]byte, 16)) {
		t.Fatal("Read returned all zero bytes")
	}
}

// Test Read for medium buffers (uses lb)
func TestRead_MediumBuffer(t *testing.T) {
	buf := make([]byte, 128)
	n, err := Read(buf)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if n != 128 {
		t.Fatalf("Read returned n=%d, want 128", n)
	}
	if bytes.Equal(buf, make([]byte, 128)) {
		t.Fatal("Read returned all zero bytes")
	}
}

// Test Read for large buffers (>512, falls back to crypto/rand)
func TestRead_LargeBuffer(t *testing.T) {
	buf := make([]byte, 1024)
	n, err := Read(buf)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if n != 1024 {
		t.Fatalf("Read returned n=%d, want 1024", n)
	}
	if bytes.Equal(buf, make([]byte, 1024)) {
		t.Fatal("Read returned all zero bytes")
	}
}

// Test Read with zero-length buffer
func TestRead_ZeroLength(t *testing.T) {
	n, err := Read([]byte{})
	if err != nil {
		t.Fatalf("Read with zero-length buffer returned error: %v", err)
	}
	if n != 0 {
		t.Fatalf("Read returned n=%d, want 0", n)
	}
}

// Test Text returns string of correct length and chars
func TestText(t *testing.T) {
	s := Text()
	if len(s) != 26 {
		t.Fatalf("Text returned string of length %d, want 26", len(s))
	}
	if !isBase32(s) {
		t.Fatalf("Text returned string with invalid characters: %s", s)
	}
}

func isBase32(s string) bool {
	for _, r := range s {
		if !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567", r) {
			return false
		}
	}
	return true
}

// Test Prime returns a prime of the correct bit length
func TestPrime(t *testing.T) {
	prime, err := Prime(rand.Reader, 128)
	if err != nil {
		t.Fatalf("Prime returned error: %v", err)
	}
	if prime.BitLen() < 128 {
		t.Fatalf("Prime returned value with bit length %d, want >=128", prime.BitLen())
	}
}

// Test Int returns a valid random int < max
func TestInt(t *testing.T) {
	max := big.NewInt(1 << 62)
	n, err := Int(rand.Reader, max)
	if err != nil {
		t.Fatalf("Int returned error: %v", err)
	}
	if n.Cmp(max) >= 0 {
		t.Fatalf("Int returned %v, expected < %v", n, max)
	}
}

// Coverage test for cachePool.New
func TestCachePool_New(t *testing.T) {
	c := cachePool.New().(*cache)
	if len(c.lb) != lbByteSize {
		t.Fatalf("Expected lb size %d, got %d", lbByteSize, len(c.lb))
	}
	if len(c.sb) != sbByteSize {
		t.Fatalf("Expected sb size %d, got %d", sbByteSize, len(c.sb))
	}
}
