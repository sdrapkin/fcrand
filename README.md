# ⚡`fcrand` (fast crypto rand)
[![name](https://goreportcard.com/badge/github.com/sdrapkin/fcrand)](https://goreportcard.com/report/github.com/sdrapkin/fcrand)
## `fcrand` is high-performance drop-in replacement for Go's `crypto/rand`.<br>By [Stan Drapkin](https://github.com/sdrapkin/).

## Usage
1. Find `"crypto/rand"` imports in your Go code
2. Add a blank identifier `_` in front: `_ "crypto/rand"`
3. Add `rand "github.com/sdrapkin/fcrand"`

## Features
- Up to **10x faster** for random data requests ≤512 bytes
- Maintains all cryptographic security guarantees of `crypto/rand`
- Complete **100% API compatibility** with `crypto/rand`: a drop-in replacement
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