// Package fcrand (fast crypto rand) implements a fast wrapper around the
// standard crypto/rand package.
//
// fcrand is designed to improve performance for applications that frequently request
// small amounts of cryptographically secure random data (under 512 bytes).
// For all other cases (>512 bytes) fcrand invokes crypto/rand.
// fcrand uses a sync.Pool of 4kb byte caches to reduce the overhead of direct
// crypto/rand system calls.
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
	lbBlockByteSize = 8                              // Size of a large buffer block in bytes
	lbBlockCount    = 1 << 9                         // 512 blocks in large buffer (power of 2)
	lbByteSize      = lbBlockByteSize * lbBlockCount // 4096 bytes per large buffer (512*8)

	// small buffer does not use blocks (ie. small buffer block size is 1 byte)
	sbByteSize = 1 << 10 // 1024 bytes per small buffer

	maxBytesToFillViaCache = 512
	sbCutoff               = 32

	/*
		Design Logic:
		Large Buffer (4KB, 8-byte blocks):
			- Efficient for requests ≥32 bytes, but wastes bytes when request size isn't a block multiple
		Small Buffer (1KB, 1-byte granularity):
			- Prevents waste for small requests <32 bytes by using every single byte

		For example:
			A 35-byte request uses the large buffer and consumes 40 bytes (5 blocks × 8), wasting 5 bytes
			A 7-byte request uses the small buffer and consumes exactly 7 bytes, no waste

		Why 32-byte small buffer cutoff:
			- Below 32 bytes: Potential waste is significant (up to 7/8 = 87.5% for small requests)
			- At 32+ bytes: Maximum waste is 7/40 = 17.5%, which becomes acceptable
	*/
)

// Ensure that the constants are not changed without thought.
var _ = map[bool]int{false: 0, lbBlockByteSize == 8: 1}
var _ = map[bool]int{false: 0, lbBlockCount == 512: 1}
var _ = map[bool]int{false: 0, lbByteSize == 4096: 1}
var _ = map[bool]int{false: 0, sbByteSize == 1024: 1}
var _ = map[bool]int{false: 0, sbCutoff == 32: 1}
var _ = map[bool]int{false: 0, maxBytesToFillViaCache == 512: 1}

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
// Deterministric Random Bit Generator (DRBG).
var Reader io.Reader
var _reader = reader{}

func init() {
	Reader = &_reader
}

type reader struct{}

func (r *reader) Read(b []byte) (n int, err error) {
	return Read(b)
}

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

	cachePtr := cachePool.Get().(*cache)

	if n < sbCutoff {
		if n > cachePtr.sbCount {
			cryptoRand.Read(cachePtr.sb)
			cachePtr.sbCount = sbByteSize
		}
		copy(b, cachePtr.sb[sbByteSize-cachePtr.sbCount:])
		cachePtr.sbCount -= n
	} else {
		if n > cachePtr.lbCount {
			cryptoRand.Read(cachePtr.lb)
			cachePtr.lbCount = lbByteSize
		}
		copy(b, cachePtr.lb[lbByteSize-cachePtr.lbCount:])

		// Update lbBytesConsumed based on the number of blocks consumed.
		// The ceiling division accounts for partial block consumption.
		lbBytesConsumed := (n + lbBlockByteSize - 1) &^ (lbBlockByteSize - 1)
		cachePtr.lbCount -= lbBytesConsumed
	}

	cachePool.Put(cachePtr)
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
	_reader.Read(src) // guaranteed not to fail since Go 1.24
	for i := range src {
		src[i] = base32_256[src[i]]
	}
	return unsafe.String(&src[0], textLength)
}

type cache struct {
	lb      []byte // large buffer
	sb      []byte // small buffer
	lbCount int    // count of bytes available in lb
	sbCount int    // count of bytes available in sb
}

// cachePool is a sync.Pool that holds cache instances.
var cachePool = sync.Pool{
	New: func() any {
		return &cache{
			lb: make([]byte, lbByteSize),
			sb: make([]byte, sbByteSize),
		}
	},
}
