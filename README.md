# resetgen

Generate allocation-free `Reset()` methods for your structs. Perfect for [`sync.Pool`](https://pkg.go.dev/sync) usage.

## Tools

This project provides two tools:

| Tool | Description |
|------|-------------|
| `resetgen` | Code generator — creates `Reset()` methods from struct tags |
| `resetgen-analyzer` | Static analyzer — detects missing `Reset()` calls before `sync.Pool.Put()` |

---

## resetgen (Code Generator)

### Installation

```bash
go install github.com/flaticols/resetgen@latest
```

Or add as a tool dependency (Go 1.24+):

```bash
go get -tool github.com/flaticols/resetgen@latest
```

### Go 1.24+ Tool Mechanism

Go 1.24 introduced the ability to manage CLI tools as dependencies. You can declare tool requirements in `go.mod`:

```go
tool (
    github.com/flaticols/resetgen
)
```

Run with `go tool`:

```bash
# Generate from current package
go tool resetgen

# Generate from specific packages
go tool resetgen ./...
go tool resetgen ./cmd ./internal

# With flags
go tool resetgen -structs User,Order ./...
go tool resetgen -version
```

This approach keeps your tool versions synchronized with your project, just like regular dependencies.

### Usage

Add `reset` tags to your struct fields and run the generator:

```go
//go:generate resetgen

package main

type Request struct {
    ID      string            `reset:""`
    Method  string            `reset:"GET"`
    Headers map[string]string `reset:""`
    Body    []byte            `reset:""`
}
```

Run:

```bash
go generate ./...
```

Generated `request.gen.go`:

```go
func (s *Request) Reset() {
    s.ID = ""
    s.Method = "GET"
    clear(s.Headers)      // preserves capacity
    s.Body = s.Body[:0]   // preserves capacity
}
```

## Tag Syntax

| Tag | Behavior |
|-----|----------|
| `reset:""` | Zero value |
| `reset:"value"` | Default value |
| `reset:"-"` | Skip field |

## CLI Flag Syntax

### `-structs` Flag

Specify which structs to generate using the `-structs` flag:

```bash
//go:generate resetgen -structs User,Order,Config

# Or with multiple files
resetgen -structs User,Order,Config ./...
```

When `-structs` is specified:
- **ONLY** the listed structs are processed (tags and directives are ignored for struct selection)
- All exported fields are reset to zero values
- Field-level `reset` tags still work for custom values or to skip specific fields

**Example:**
```go
//go:generate resetgen -structs User,Order

type User struct {
    ID      int64
    Name    string
    Secret  string `reset:"-"` // Still respected - field will not be reset
}

type Order struct {
    ID    int64
    Items []string
    Total float64 `reset:"0.0"` // Custom value still works
}

type Logger struct {
    Level string  // Will NOT be generated (not in -structs list)
}
```

### Package-Qualified Names

When you have structs with the same name in different packages, use package-qualified names:

```bash
# Process User in models package only
resetgen -structs models.User ./...

# Process User in both models and api packages
resetgen -structs models.User,api.User ./...

# Mix simple and qualified names
resetgen -structs Order,models.User ./...
```

**Rules:**
- Simple name (`User`) → processes ALL User structs in all packages
- Qualified name (`models.User`) → processes only User in models package
- Package path uses Go import path format (lowercase with dots/slashes)

**Example with multiple packages:**
```go
// models/user.go
//go:generate resetgen -structs models.User,api.User

package models

type User struct {
    ID    int64  `reset:""`
    Name  string `reset:""`
    Email string `reset:""`
}

// api/user.go
//go:generate resetgen -structs models.User,api.User

package api

type User struct {
    ID       string `reset:""`
    Status   string `reset:"active"`
}
```

Both packages can use the same go:generate directive with package-qualified names, and each will generate only its own Reset() method.

## Directive Syntax

Use the `+resetgen` comment directive to mark structs for automatic `Reset()` generation without tagging every field:

```go
//go:generate resetgen

package main

// +resetgen
type Request struct {
    ID      string            // defaults to zero value
    Method  string            // defaults to zero value
    Headers map[string]string // defaults to zero value
    Secret  string `reset:"-"` // skipped from reset
}
```

Generated `request.gen.go`:

```go
func (s *Request) Reset() {
    s.ID = ""
    s.Method = ""
    clear(s.Headers)  // preserves capacity
    // Secret is not reset (reset:"-")
}
```

### How Directive Works

- **Struct Selection**: Structs are processed if they have a `+resetgen` comment OR contain `reset` tags
- **Field Processing**: All exported fields are reset to zero values
- **Custom Values**: Fields with explicit `reset` tags use their specified values
- **Skip Fields**: Use `reset:"-"` to exclude specific fields from reset
- **Unexported Fields**: Private fields (lowercase) are automatically skipped for safety

### Directive Formats

All of these are recognized:
- `//+resetgen`
- `// +resetgen`
- `//  +resetgen`
- `/* +resetgen */`

## Features

- **Allocation-free** — slices truncate (`s[:0]`), maps clear (`clear(m)`)
- **Embedded structs** — calls `Reset()` recursively
- **Selective** — only structs with `reset` tags are processed
- **Fast** — single-pass AST, minimal allocations

> [!TIP]
> Structs without any `reset` tags are automatically ignored. You can have pooled and regular structs in the same file.

## Resetter Interface

Define a common interface for pooled objects:

```go
type Resetter interface {
    Reset()
}
```

All generated `Reset()` methods satisfy this interface, enabling generic pool helpers.

## sync.Pool Examples

### Example 1: HTTP Request Pool

```go
//go:generate resetgen

package server

type Request struct {
    Path    string            `reset:""`
    Method  string            `reset:"GET"`
    Headers map[string]string `reset:""`
    Body    []byte            `reset:""`
    UserID  int               `reset:""`
}

var requestPool = sync.Pool{
    New: func() any { return new(Request) },
}

func HandleRequest(path, method string, body []byte) {
    req := requestPool.Get().(*Request)

    req.Path = path
    req.Method = method
    req.Body = append(req.Body, body...)

    process(req)

    req.Reset()
    requestPool.Put(req)
}
```

### Example 2: Generic Pool with Resetter

```go
//go:generate resetgen

package pool

type Resetter interface {
    Reset()
}

// Pool is a generic, allocation-free pool for any Resetter
type Pool[T Resetter] struct {
    p sync.Pool
}

func NewPool[T Resetter](newFn func() T) *Pool[T] {
    return &Pool[T]{
        p: sync.Pool{New: func() any { return newFn() }},
    }
}

func (p *Pool[T]) Get() T      { return p.p.Get().(T) }
func (p *Pool[T]) Put(v T)     { v.Reset(); p.p.Put(v) }
```

Usage with a buffer:

```go
//go:generate resetgen

package encoding

type Buffer struct {
    data []byte `reset:""`
}

var bufPool = pool.NewPool(func() *Buffer { return new(Buffer) })

// MarshalTo writes encoded data directly to dst — zero allocations
func MarshalTo(dst io.Writer, v any) error {
    buf := bufPool.Get()

    buf.data = encodeJSON(buf.data, v)
    _, err := dst.Write(buf.data)

    bufPool.Put(buf)
    return err
}
```

> [!NOTE]
> Both examples avoid defer closures and return values that reference pooled memory.

---

## resetgen-analyzer (Static Analyzer)

Detects when `sync.Pool.Put()` is called without a preceding `Reset()` call.

### Installation

```bash
go install github.com/flaticols/resetgen/cmd/resetgen-analyzer@latest
```

Or add as a tool dependency (Go 1.24+):

```bash
go get -tool github.com/flaticols/resetgen/cmd/resetgen-analyzer
```

### Usage

Run standalone:

```bash
resetgen-analyzer ./...
```

Run with `go vet`:

```bash
go vet -vettool=$(which resetgen-analyzer) ./...
```

Add to your CI pipeline or Makefile:

```makefile
.PHONY: lint
lint:
	go vet -vettool=$(which resetgen-analyzer) ./...
```

### What it detects

```go
func BadUsage() {
    buf := bufferPool.Get().(*Buffer)
    buf.data = append(buf.data, "hello"...)
    bufferPool.Put(buf) // ERROR: sync.Pool.Put() called without Reset() on buf
}

func GoodUsage() {
    buf := bufferPool.Get().(*Buffer)
    buf.data = append(buf.data, "hello"...)
    buf.Reset()
    bufferPool.Put(buf) // OK
}
```

### Detected patterns

| Pattern | Example |
|---------|---------|
| Global pool | `bufferPool.Put(buf)` without `buf.Reset()` |
| Struct field pool | `s.pool.Put(buf)` without `buf.Reset()` |
| Wrapped variable | `pool.Put(w.buf)` without `w.buf.Reset()` |
| Pool wrapper | `p.p.Put(v)` without `v.Reset()` inside wrapper method |

## Benchmarks

```
BenchmarkReset-8    1000000000    0.32 ns/op    0 B/op    0 allocs/op
```

## License

[MIT](LICENSE)

---

<p align="center">Made with ❤️ by Denis</p>
