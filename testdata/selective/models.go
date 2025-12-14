//go:generate go run ../..

package selective

// PooledRequest has reset tags - Reset() will be generated.
// Use this struct with sync.Pool.
type PooledRequest struct {
	ID      string            `reset:""`
	Method  string            `reset:"GET"`
	Path    string            `reset:""`
	Headers map[string]string `reset:""`
	Body    []byte            `reset:""`
}

// PooledResponse has reset tags - Reset() will be generated.
type PooledResponse struct {
	Status int    `reset:"200"`
	Body   []byte `reset:""`
}

// Config has NO reset tags - will be IGNORED.
// This struct is not used with pools, just regular config.
type Config struct {
	Host    string
	Port    int
	Timeout int
	Debug   bool
}

// Logger has NO reset tags - will be IGNORED.
type Logger struct {
	Level  string
	Output string
}

// MixedStruct shows partial tagging - only tagged fields are reset.
// Fields without reset tag are left unchanged.
type MixedStruct struct {
	// These fields will be reset:
	RequestID string   `reset:""`
	Buffer    []byte   `reset:""`
	Counter   int      `reset:"0"`

	// These fields are NOT tagged - they keep their values:
	CreatedAt int64  // keeps value
	Owner     string // keeps value
}
