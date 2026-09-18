// Package fcrand (fast crypto rand) implements a fast wrapper around the
// standard crypto/rand package.
//
// fcrand is designed to improve performance for applications that frequently request
// small amounts of cryptographically secure random data (up to 512 bytes).
// For all other cases (>512 bytes) fcrand invokes crypto/rand.
// fcrand uses a 4KB sync.Pool cache to reduce the overhead of direct crypto/rand system calls.
//
// The fcrand API provides a drop-in replacement for crypto/rand,
// including the Reader, Int(), Prime(), and Read() functions.
// All random data is sourced from crypto/rand, preserving all its
// cryptographic security guarantees.
package fcrand

import (
	cryptoRand "crypto/rand"
	"io"
	"math/big"
	"sync"
	"unsafe"
)

const (
	bufferByteSize         = 1024*4 - 8 // 4088 bytes
	maxBytesToFillViaCache = 512
)

// Ensure that the constants are not changed without thought.
var _ [0]struct{} = [bufferByteSize - 4088]struct{}{}
var _ [0]struct{} = [maxBytesToFillViaCache - 512]struct{}{}

type readerStruct struct{}

func (rs readerStruct) Read(b []byte) (int, error) { return Read(b) }

// Reader is a global, shared instance of a cryptographically
// secure random number generator. It is safe for concurrent use.
//
//   - On Linux, FreeBSD, Dragonfly, and Solaris, Reader uses getrandom(2).
//   - On legacy Linux (< 3.17), Reader opens /dev/urandom on first use.
//   - On macOS, iOS, and OpenBSD Reader, uses arc4random_buf(3).
//   - On NetBSD, Reader uses the kern.arandom sysctl.
//   - On Windows, Reader uses the ProcessPrng API.
//   - On js/wasm, Reader uses the Web Crypto API.
//   - On wasip1/wasm, Reader uses random_get.
//
// In FIPS 140-3 mode, the output passes through an SP 800-90A Rev. 1
// Deterministic Random Bit Generator (DRBG).
var Reader io.Reader = readerStruct{}

// Read fills b with cryptographically secure random bytes.
// It never returns an error, and always fills b entirely.
func Read(b []byte) (n int, err error) {
	n = len(b)

	if n == 0 {
		return 0, nil
	}

	if n > maxBytesToFillViaCache {
		return cryptoRand.Read(b)
	}

	cachei := cachePool.Get()
	pcache := cachei.(*cacheStruct)
	offset := pcache.offset

	if offset+n > bufferByteSize {
		cryptoRand.Read(pcache.buffer[:])
		offset = 0
	}

	copy(b, pcache.buffer[offset:])
	pcache.offset = offset + n

	cachePool.Put(cachei)
	return n, nil
}

// Prime returns a number of the given bit length that is prime with high probability.
// Prime will return error for any error returned by rand.Read or if bits < 2.
func Prime(rand io.Reader, bits int) (*big.Int, error) {
	return cryptoRand.Prime(rand, bits)
}

// Int returns a uniform random value in [0, max). It panics if max <= 0, and
// returns an error if rand.Read returns one.
func Int(rand io.Reader, max *big.Int) (n *big.Int, err error) {
	return cryptoRand.Int(rand, max)
}

// Text returns a cryptographically random string using the standard RFC 4648 base32 alphabet
// for use when a secret string, token, password, or other text is needed.
// The result contains at least 128 bits of randomness, enough to prevent brute force
// guessing attacks and to make the likelihood of collisions vanishingly small.
// A future version may return longer texts as needed to maintain those properties.
func Text() string {
	const (
		base32 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567" // Standard Base32 encoding alphabet from RFC 4648.
		// base32_256 is the base32 repeated 8 times to cover all byte values (0-255).
		base32_256 = base32 + base32 + base32 + base32 + base32 + base32 + base32 + base32
		textLength = 26 // ⌈log₃₂ 2¹²⁸⌉ = 26 chars
	)

	src := make([]byte, textLength)
	Read(src) // guaranteed not to fail since Go 1.24
	for i, ch := range src {
		src[i] = base32_256[ch]
	}
	return unsafe.String(&src[0], textLength)
}

// cacheStruct holds a pre-filled random buffer reused across calls via cachePool.
type cacheStruct struct {
	buffer [bufferByteSize]byte // large buffer
	offset int
}

// cachePool is a sync.Pool that holds cacheStruct instances.
var cachePool = sync.Pool{
	New: func() any { return &cacheStruct{offset: bufferByteSize} },
}
