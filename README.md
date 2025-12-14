# resetgen

Generate allocation-free `Reset()` methods for your structs. Perfect for [`sync.Pool`](https://pkg.go.dev/sync) usage.

## Installation

```bash
go install github.com/flaticols/resetgen@latest
```

Or add as a tool dependency (Go 1.24+):

```bash
go get -tool github.com/flaticols/resetgen
```

## Usage

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

## Features

- **Allocation-free** — slices truncate (`s[:0]`), maps clear (`clear(m)`)
- **Embedded structs** — calls `Reset()` recursively
- **Selective** — only structs with `reset` tags are processed
- **Fast** — single-pass AST, minimal allocations

> [!TIP]
> Structs without any `reset` tags are automatically ignored. You can have pooled and regular structs in the same file.

## Example

```go
//go:generate resetgen

package pool

type Buffer struct {
    Data    []byte            `reset:""`
    Headers map[string]string `reset:""`
    Status  int               `reset:"200"`
    err     error             // no tag = unchanged
}

// Config has no reset tags — ignored
type Config struct {
    Timeout int
    Debug   bool
}
```

> [!NOTE]
> Only `Buffer` gets a `Reset()` method. `Config` is ignored.

## Benchmarks

```
BenchmarkReset-8    1000000000    0.32 ns/op    0 B/op    0 allocs/op
```

## License

[MIT](LICENSE)

---

<p align="center">Made with ❤️ by Denis</p>
