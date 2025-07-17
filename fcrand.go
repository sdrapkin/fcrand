// Package fcrand (fast crypto rand) implements a fast wrapper around the
// standard crypto/rand package.
//
// fcrand is designed to improve performance for applications that frequently request
// small amounts of cryptographically secure random data (under 512 bytes).
// It uses a sync.Pool of pre-filled 4kb byte caches to reduce the overhead of direct
// crypto/rand system calls for random numbers.
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
	blockByteSize  = 8                              // Size of a block in bytes
	blocksPerCache = 1 << 9                         // 512 blocks per cache (power of 2)
	cacheByteSize  = blockByteSize * blocksPerCache // 4096 bytes per cache (512*8)
)

// Ensure that the constants are not changed without thought.
var _ = map[bool]int{false: 0, blocksPerCache == 512: 1}
var _ = map[bool]int{false: 0, cacheByteSize == 4096: 1}

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
	const MaxBytesToFillViaCache = 512
	n = len(b)

	if n == 0 {
		return 0, nil
	}

	if n > MaxBytesToFillViaCache {
		return cryptoRand.Read(b)
	}

	cachePtr := cachePool.Get().(*cache)

	if n > cachePtr.bytesAvailable {
		// Not enough bytes remaining: refill cache completely.
		// Go 1.24+ guarantees crypto/rand.Read succeeds.
		cryptoRand.Read(cachePtr.buffer)
		cachePtr.bytesAvailable = cacheByteSize
	}

	copy(b, cachePtr.buffer[cacheByteSize-cachePtr.bytesAvailable:])

	// Update bytesAvailable based on the number of blocks consumed.
	// The ceiling division accounts for partial block consumption.
	bytesConsumed := (n + blockByteSize - 1) &^ (blockByteSize - 1)
	cachePtr.bytesAvailable -= bytesConsumed

	cachePool.Put(cachePtr)
	return n, nil
}

// Read fills b with cryptographically secure random bytes.
// It never returns an error, and always fills b entirely.
func Read(b []byte) (n int, err error) {
	return _reader.Read(b)
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
	buffer         []byte
	bytesAvailable int
	_              [32]byte // pad ensures each bytesAvailable is on its own cache line
}

// cachePool is a sync.Pool that holds cache instances.
var cachePool = sync.Pool{
	New: func() any {
		return &cache{buffer: make([]byte, cacheByteSize)}
	},
}
