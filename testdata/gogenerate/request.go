//go:generate go run ../..

package gogenerate

// Request demonstrates using go:generate directive.
//
// Usage options:
//   //go:generate resetgen                         (if installed)
//   //go:generate go tool resetgen                 (Go 1.24+ with tool in go.mod)
//   //go:generate go run github.com/flaticols/resetgen
type Request struct {
	ID      string            `reset:""`
	Method  string            `reset:"GET"`
	Path    string            `reset:""`
	Headers map[string]string `reset:""`
	Body    []byte            `reset:""`
	Query   map[string]string `reset:""`
}

// Response is another struct in the same file.
type Response struct {
	Status  int               `reset:"200"`
	Headers map[string]string `reset:""`
	Body    []byte            `reset:""`
}
