# ⚡`fcrand` (fast crypto rand)
[![name](https://goreportcard.com/badge/github.com/sdrapkin/fcrand)](https://goreportcard.com/report/github.com/sdrapkin/fcrand)
## `fcrand` is high-performance drop-in replacement for Go's `crypto/rand`.<br>By [Stan Drapkin](https://github.com/sdrapkin/).

## Usage
1. Find `"crypto/rand"` imports in your Go code
2. Add a blank identifier `_` in front: `_ "crypto/rand"`
3. Add `rand "github.com/sdrapkin/fcrand"`

## Features
- Up to 10x faster for random data requests ≤512 bytes
- Maintains all cryptographic security guarantees of `crypto/rand`
- Complete 100% API compatibility with `crypto/rand`
- Thread-safe with zero configuration

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