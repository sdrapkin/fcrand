# ⚡`fcrand` (fast crypto rand)
[![name](https://goreportcard.com/badge/github.com/sdrapkin/fcrand)](https://goreportcard.com/report/github.com/sdrapkin/fcrand)
## `fcrand` is a high-performance drop-in replacement for Go's `crypto/rand`.<br>By [Stan Drapkin](https://github.com/sdrapkin/).

## Usage
1. Find `"crypto/rand"` imports in your Go code
2. Add a blank identifier `_` in front: `_ "crypto/rand"`
3. Add `rand "github.com/sdrapkin/fcrand"`

## Features
- Up to **10x faster** for random data requests ≤512 bytes
- Maintains all cryptographic security guarantees of `crypto/rand`
- **100% API compatible** with `crypto/rand` – true drop-in replacement
- Thread-safe with zero configuration (same as `crypto/rand`)

### ⚙️Before:
```go
import (
    "crypto/rand"
)
```
### ✅ After:
```go
import (
    _ "crypto/rand"
    rand "github.com/sdrapkin/fcrand"
)
```
## Example
[go playground](https://go.dev/play/p/5SDsQH5RMbC)
```go
package main

import (
	"fmt"

	rand "github.com/sdrapkin/fcrand"
)

func main() {
	// Using rand.Read()
	buf := make([]byte, 16)
	rand.Read(buf)
	fmt.Printf("rand.Read buf:   [%x]\n", buf)

	// Using rand.Reader
	rand.Reader.Read(buf)
	fmt.Printf("rand.Reader buf: [%x]\n", buf)

	// Using .Text()
	token := rand.Text()
	fmt.Printf("token: [%s]\n", token)

	// Use .Prime() with rand.Reader
	key, err := rand.Prime(rand.Reader, 1024)
	fmt.Printf("prime: [%q]\n%v\n", key, err)
}
```

## Requirements
- Go 1.24+

## Installation
### Using `go get`

To install the `fcrand` package, run the following command:

```sh
go get -u github.com/sdrapkin/fcrand
```

To use the `fcrand` package in your Go project, import it as follows:

```go
import rand "github.com/sdrapkin/fcrand"
```

## FIPS Ready
* **FIPS-140 ready** (https://go.dev/doc/security/fips140)
	* set `GODEBUG=fips140=on` environment variable
	* https://go.dev/blog/fips140

## Benchmarks
**Key observations:** `fcrand` delivers 5-10x speedups for typical small requests, with even better concurrent performance. Performance falls back to `crypto/rand` levels at 512+ bytes as designed.

### Serial bench (1 goroutine):
| buf size | #g | fcrand ns/op | crypto/rand ns/op | speedup % | speedup x |
|------|---|---------|---------|------|-------|
| 0    | 1 | 3.08    | 134.90  | 98%  | 43.8x |
| 1    | 1 | 23.21   | 209.30  | 89%  | 9x    |
| 2    | 1 | 20.74   | 186.80  | 89%  | 9x    |
| 3    | 1 | 19.80   | 181.50  | 89%  | 9.2x  |
| 4    | 1 | 18.80   | 178.40  | 89%  | 9.5x  |
| 5    | 1 | 20.04   | 182.00  | 89%  | 9.1x  |
| 6    | 1 | 20.35   | 180.20  | 89%  | 8.9x  |
| 7    | 1 | 27.09   | 186.10  | 85%  | 6.9x  |
| 8    | 1 | 29.66   | 184.30  | 84%  | 6.2x  |
| 9    | 1 | 29.31   | 189.40  | 85%  | 6.5x  |
| 12   | 1 | 31.36   | 230.30  | 86%  | 7.3x  |
| 16   | 1 | 34.62   | 208.20  | 83%  | 6x    |
| 17   | 1 | 39.44   | 200.90  | 80%  | 5.1x  |
| 25   | 1 | 41.59   | 219.90  | 81%  | 5.3x  |
| 31   | 1 | 44.47   | 226.20  | 80%  | 5.1x  |
| 32   | 1 | 37.36   | 223.00  | 83%  | 6x    |
| 33   | 1 | 41.45   | 231.10  | 82%  | 5.6x  |
| 48   | 1 | 45.92   | 252.10  | 82%  | 5.5x  |
| 56   | 1 | 49.69   | 270.70  | 82%  | 5.4x  |
| 57   | 1 | 52.92   | 329.70  | 84%  | 6.2x  |
| 64   | 1 | 53.59   | 300.50  | 82%  | 5.6x  |
| 65   | 1 | 72.28   | 290.10  | 75%  | 4x    |
| 128  | 1 | 83.68   | 373.60  | 78%  | 4.5x  |
| 256  | 1 | 139.40  | 423.40  | 67%  | 3x    |
| 512  | 1 | 253.30  | 502.50  | 50%  | 2x    |
| 513  | 1 | 554.30  | 545.20  | -2%  | 1x    |
| 1024 | 1 | 702.60  | 688.80  | -2%  | 1x    |
| 2048 | 1 | 1049.00 | 1120.00 | 6%   | 1.1x  |
| 4096 | 1 | 1784.00 | 1775.00 | -1%  | 1x    |

### Concurrent bench (64 goroutines on a 16 vCPU machine):
| buf size | #g | fcrand ns/op | crypto/rand ns/op | speedup % | speedup x |
|------|----|--------|--------|-----|-------|
| 0    | 64 | 0.72   | 58.45  | 99% | 81.2x |
| 1    | 64 | 6.48   | 72.86  | 91% | 11.2x |
| 2    | 64 | 6.37   | 73.87  | 91% | 11.6x |
| 3    | 64 | 11.37  | 76.99  | 85% | 6.8x  |
| 4    | 64 | 12.24  | 85.54  | 86% | 7x    |
| 5    | 64 | 11.70  | 82.37  | 86% | 7x    |
| 6    | 64 | 11.88  | 89.45  | 87% | 7.5x  |
| 7    | 64 | 12.51  | 84.41  | 85% | 6.7x  |
| 8    | 64 | 13.38  | 80.34  | 83% | 6x    |
| 9    | 64 | 14.95  | 81.98  | 82% | 5.5x  |
| 12   | 64 | 14.10  | 80.43  | 82% | 5.7x  |
| 16   | 64 | 15.42  | 84.61  | 82% | 5.5x  |
| 17   | 64 | 14.85  | 93.45  | 84% | 6.3x  |
| 25   | 64 | 17.71  | 97.25  | 82% | 5.5x  |
| 31   | 64 | 18.72  | 97.43  | 81% | 5.2x  |
| 32   | 64 | 14.61  | 98.35  | 85% | 6.7x  |
| 33   | 64 | 15.81  | 102.00 | 85% | 6.5x  |
| 48   | 64 | 17.11  | 111.80 | 85% | 6.5x  |
| 56   | 64 | 18.06  | 122.10 | 85% | 6.8x  |
| 57   | 64 | 18.51  | 122.20 | 85% | 6.6x  |
| 64   | 64 | 18.68  | 127.60 | 85% | 6.8x  |
| 65   | 64 | 20.27  | 126.30 | 84% | 6.2x  |
| 128  | 64 | 28.68  | 142.20 | 80% | 5x    |
| 256  | 64 | 55.05  | 156.60 | 65% | 2.8x  |
| 512  | 64 | 100.50 | 182.10 | 45% | 1.8x  |
| 513  | 64 | 200.30 | 194.10 | -3% | 1x    |
| 1024 | 64 | 241.50 | 232.30 | -4% | 1x    |
| 2048 | 64 | 349.20 | 361.60 | 3%  | 1x    |
| 4096 | 64 | 675.50 | 684.20 | 1%  | 1x    |
