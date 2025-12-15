//go:generate go run ../..

package directive

// +resetgen
type Request struct {
	ID      string
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

// +resetgen
type Response struct {
	Status  int       `reset:"200"`
	Body    []byte
	Headers map[string]string
}

// Struct with directive and mixed tags
// +resetgen
type Config struct {
	Host    string
	Port    int       `reset:"8080"`
	Timeout int
	secret  string // unexported, will be skipped
}

// Regular struct with tags (no directive) - should still work
type User struct {
	ID    int64  `reset:""`
	Name  string `reset:"unknown"`
	Email string
}

// Struct without directive and no tags - should be ignored
type NoTags struct {
	Field1 string
	Field2 int
}
