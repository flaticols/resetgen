package complex

import (
	"net/netip"
	"time"
)

// Metadata holds common entity metadata
type Metadata struct {
	ID        int64     `reset:""`
	CreatedAt time.Time `reset:""`
	UpdatedAt time.Time `reset:""`
	Version   int       `reset:"1"`
}

// Address represents a physical address
type Address struct {
	Street     string `reset:""`
	City       string `reset:""`
	State      string `reset:""`
	PostalCode string `reset:""`
	Country    string `reset:"US"`
}

// Contact holds contact information
type Contact struct {
	Email       string   `reset:""`
	Phone       string   `reset:""`
	PhoneAlt    *string  `reset:""`
	Preferences []string `reset:""`
}

// Credentials stores authentication data
type Credentials struct {
	PasswordHash []byte            `reset:""`
	Salt         []byte            `reset:""`
	TokenHash    *[32]byte         `reset:""`
	Sessions     map[string]int64  `reset:""`
	Permissions  map[string]bool   `reset:""`
}

// Profile contains user profile data
type Profile struct {
	DisplayName string            `reset:"Anonymous"`
	AvatarURL   string            `reset:""`
	Bio         string            `reset:""`
	Location    *Address          `reset:""`
	Social      map[string]string `reset:""`
	Tags        []string          `reset:""`
	Verified    bool              `reset:""`
}

// Subscription represents a user subscription
type Subscription struct {
	PlanID      string     `reset:"free"`
	StartedAt   time.Time  `reset:""`
	ExpiresAt   *time.Time `reset:""`
	AutoRenew   bool       `reset:"true"`
	PaymentInfo *string    `reset:""`
}

// UsageStats tracks resource usage
type UsageStats struct {
	RequestCount   int64   `reset:""`
	BytesIn        int64   `reset:""`
	BytesOut       int64   `reset:""`
	ErrorCount     int64   `reset:""`
	LastActivityAt int64   `reset:""`
	Endpoints      []string `reset:""`
	StatusCodes    map[int]int64 `reset:""`
}

// RateLimit holds rate limiting state
type RateLimit struct {
	Limit     int   `reset:"100"`
	Remaining int   `reset:"100"`
	ResetAt   int64 `reset:""`
	Tokens    []int64 `reset:""`
}

// User is a comprehensive user entity for testing
type User struct {
	// Embedded types
	Metadata `reset:""`

	// Basic fields
	Username    string `reset:""`
	Email       string `reset:""`
	DisplayName string `reset:"Guest"`
	Locale      string `reset:"en-US"`
	Timezone    string `reset:"UTC"`

	// Status fields
	Active      bool `reset:""`
	Verified    bool `reset:""`
	Suspended   bool `reset:""`
	DeletedAt   *time.Time `reset:""`

	// Pointer to struct
	Profile     *Profile `reset:""`
	Contact     *Contact `reset:""`

	// Non-pointer struct
	Address     Address `reset:""`
	Credentials Credentials `reset:""`

	// Subscription info
	Subscription    Subscription   `reset:""`
	Subscriptions   []Subscription `reset:""`

	// Usage tracking
	Usage      UsageStats `reset:""`
	RateLimit  RateLimit  `reset:""`

	// Collections
	Roles           []string          `reset:""`
	Permissions     []string          `reset:""`
	Groups          []int64           `reset:""`
	Followers       []int64           `reset:""`
	Following       []int64           `reset:""`
	BlockedUsers    []int64           `reset:""`

	// Maps
	Settings        map[string]any    `reset:""`
	Preferences     map[string]string `reset:""`
	FeatureFlags    map[string]bool   `reset:""`
	CustomFields    map[string]any    `reset:""`

	// Nested collections
	Addresses       []Address         `reset:""`
	PaymentMethods  []*string         `reset:""`

	// Internal - ignored
	internalCache   map[string]any
	internalLock    bool
}

// HTTPRequest represents a pooled HTTP request context
type HTTPRequest struct {
	// Metadata
	ID          string `reset:""`
	TraceID     string `reset:""`
	SpanID      string `reset:""`
	ParentID    string `reset:""`

	// Request info
	Method      string `reset:"GET"`
	Path        string `reset:""`
	RawQuery    string `reset:""`
	Proto       string `reset:"HTTP/1.1"`
	Host        string `reset:""`
	RemoteAddr  string `reset:""`

	// Headers (reusable slices)
	HeaderKeys   []string `reset:""`
	HeaderValues []string `reset:""`

	// Query params (reusable)
	QueryKeys    []string `reset:""`
	QueryValues  []string `reset:""`

	// Body
	Body         []byte `reset:""`
	ContentType  string `reset:""`
	ContentLen   int64  `reset:""`

	// Parsed data
	FormData     map[string][]string `reset:""`
	Cookies      map[string]string   `reset:""`

	// Auth
	AuthType     string  `reset:""`
	AuthToken    string  `reset:""`
	UserID       *int64  `reset:""`
	SessionID    *string `reset:""`

	// Context values
	ContextKeys   []string `reset:""`
	ContextValues []any    `reset:""`

	// Timing
	StartTime    int64 `reset:""`
	ParseTime    int64 `reset:""`
	AuthTime     int64 `reset:""`

	// Response tracking
	StatusCode   int   `reset:"200"`
	BytesWritten int64 `reset:""`

	// Error handling
	Errors       []error `reset:""`
	Warnings     []string `reset:""`

	// Internal
	processed bool
}

// HTTPResponse represents a pooled HTTP response context
type HTTPResponse struct {
	// Status
	StatusCode   int    `reset:"200"`
	StatusText   string `reset:"OK"`

	// Headers
	HeaderKeys   []string `reset:""`
	HeaderValues []string `reset:""`

	// Body buffer
	Body         []byte `reset:""`
	BodyLen      int    `reset:""`

	// Compression
	Compressed   bool   `reset:""`
	CompressType string `reset:""`

	// Streaming
	Chunked      bool     `reset:""`
	Chunks       [][]byte `reset:""`

	// Cookies to set
	SetCookies   []string `reset:""`

	// Cache control
	CacheControl string `reset:""`
	ETag         string `reset:""`
	LastModified int64  `reset:""`

	// Timing
	StartTime    int64 `reset:""`
	FirstByte    int64 `reset:""`
	EndTime      int64 `reset:""`
}

// DatabaseQuery represents a pooled database query
type DatabaseQuery struct {
	// Query
	SQL          string   `reset:""`
	Args         []any    `reset:""`
	ArgTypes     []string `reset:""`

	// Prepared statement
	StmtName     string `reset:""`
	StmtCached   bool   `reset:""`

	// Transaction
	TxID         *string `reset:""`
	TxIsolation  string  `reset:""`

	// Results
	Columns      []string  `reset:""`
	ColumnTypes  []string  `reset:""`
	Rows         [][]any   `reset:""`
	RowsAffected int64     `reset:""`
	LastInsertID int64     `reset:""`

	// Pagination
	Offset       int64 `reset:""`
	Limit        int64 `reset:"100"`
	TotalCount   int64 `reset:""`

	// Timing
	PrepareTime  int64 `reset:""`
	ExecuteTime  int64 `reset:""`
	FetchTime    int64 `reset:""`

	// Errors
	Error        error    `reset:""`
	Warnings     []string `reset:""`

	// Debug
	ExplainPlan  string `reset:""`
	SlowQuery    bool   `reset:""`
}

// CacheEntry represents a pooled cache entry
type CacheEntry struct {
	Key          string `reset:""`
	Value        []byte `reset:""`
	ValueType    string `reset:""`

	// TTL
	ExpiresAt    int64 `reset:""`
	CreatedAt    int64 `reset:""`
	AccessedAt   int64 `reset:""`
	AccessCount  int64 `reset:""`

	// Metadata
	Tags         []string          `reset:""`
	Dependencies []string          `reset:""`
	Metadata     map[string]string `reset:""`

	// Compression
	Compressed   bool   `reset:""`
	OriginalSize int    `reset:""`

	// Versioning
	Version      int64  `reset:""`
	ETag         string `reset:""`

	// Flags
	NoExpire     bool `reset:""`
	NoEvict      bool `reset:""`
	Dirty        bool `reset:""`
}

// LogEntry represents a pooled log entry
type LogEntry struct {
	// Core fields
	Timestamp   int64  `reset:""`
	Level       string `reset:"INFO"`
	Message     string `reset:""`
	Logger      string `reset:""`

	// Context
	TraceID     string `reset:""`
	SpanID      string `reset:""`
	RequestID   string `reset:""`
	UserID      string `reset:""`
	SessionID   string `reset:""`

	// Source
	File        string `reset:""`
	Line        int    `reset:""`
	Function    string `reset:""`
	Package     string `reset:""`

	// Error info
	Error       error  `reset:""`
	ErrorType   string `reset:""`
	StackTrace  string `reset:""`

	// Structured data
	Fields      map[string]any `reset:""`
	Tags        []string       `reset:""`

	// HTTP context
	Method      string `reset:""`
	Path        string `reset:""`
	StatusCode  int    `reset:""`
	Duration    int64  `reset:""`

	// Database context
	Query       string `reset:""`
	QueryTime   int64  `reset:""`
	RowCount    int64  `reset:""`

	// Sampling
	Sampled     bool    `reset:""`
	SampleRate  float64 `reset:"1.0"`
}

// NetworkConnection represents a pooled network connection state
type NetworkConnection struct {
	// Connection info
	ID           string     `reset:""`
	Protocol     string     `reset:"tcp"`
	LocalAddr    netip.Addr `reset:""`
	LocalPort    uint16     `reset:""`
	RemoteAddr   netip.Addr `reset:""`
	RemotePort   uint16     `reset:""`

	// State
	State        string `reset:""`
	Connected    bool   `reset:""`
	Encrypted    bool   `reset:""`
	Compressed   bool   `reset:""`

	// TLS
	TLSVersion   string   `reset:""`
	CipherSuite  string   `reset:""`
	ServerName   string   `reset:""`
	Certificates [][]byte `reset:""`

	// Buffers
	ReadBuffer   []byte `reset:""`
	WriteBuffer  []byte `reset:""`

	// Stats
	BytesRead    int64 `reset:""`
	BytesWritten int64 `reset:""`
	PacketsIn    int64 `reset:""`
	PacketsOut   int64 `reset:""`

	// Timing
	ConnectedAt  int64 `reset:""`
	LastReadAt   int64 `reset:""`
	LastWriteAt  int64 `reset:""`

	// Errors
	ReadErrors   int   `reset:""`
	WriteErrors  int   `reset:""`
	LastError    error `reset:""`

	// Metadata
	Tags         []string          `reset:""`
	Metadata     map[string]string `reset:""`
}

// WorkerTask represents a pooled worker task
type WorkerTask struct {
	// Task info
	ID           string `reset:""`
	Type         string `reset:""`
	Queue        string `reset:"default"`
	Priority     int    `reset:"0"`

	// Payload
	Payload      []byte            `reset:""`
	PayloadType  string            `reset:""`
	Args         map[string]any    `reset:""`

	// Scheduling
	ScheduledAt  int64  `reset:""`
	StartedAt    int64  `reset:""`
	CompletedAt  int64  `reset:""`
	Deadline     int64  `reset:""`

	// Retry
	Attempts     int      `reset:""`
	MaxAttempts  int      `reset:"3"`
	RetryDelay   int64    `reset:""`
	RetryErrors  []string `reset:""`

	// Progress
	Progress     float64 `reset:""`
	ProgressMsg  string  `reset:""`
	Checkpoints  []int64 `reset:""`

	// Result
	Result       []byte `reset:""`
	ResultType   string `reset:""`
	Error        error  `reset:""`

	// Dependencies
	DependsOn    []string `reset:""`
	Blocks       []string `reset:""`

	// Metadata
	Tags         []string          `reset:""`
	Labels       map[string]string `reset:""`

	// Tracing
	TraceID      string `reset:""`
	SpanID       string `reset:""`
	ParentID     string `reset:""`
}

// EventMessage represents a pooled event/message
type EventMessage struct {
	// Message info
	ID           string `reset:""`
	Type         string `reset:""`
	Source       string `reset:""`
	Subject      string `reset:""`

	// Content
	ContentType  string `reset:"application/json"`
	Data         []byte `reset:""`
	DataSchema   string `reset:""`

	// Routing
	Topic        string   `reset:""`
	Partition    int      `reset:""`
	Key          []byte   `reset:""`
	Headers      map[string]string `reset:""`

	// Timing
	Timestamp    int64 `reset:""`
	PublishedAt  int64 `reset:""`
	ReceivedAt   int64 `reset:""`
	ProcessedAt  int64 `reset:""`

	// Delivery
	DeliveryMode int    `reset:""`
	Redelivered  bool   `reset:""`
	ReplyTo      string `reset:""`
	CorrelationID string `reset:""`

	// Acknowledgment
	Acknowledged bool   `reset:""`
	AckTime      int64  `reset:""`
	NackReason   string `reset:""`

	// Tracing
	TraceID      string            `reset:""`
	SpanID       string            `reset:""`
	TraceContext map[string]string `reset:""`

	// Error handling
	Error        error    `reset:""`
	RetryCount   int      `reset:""`
	DeadLettered bool     `reset:""`
}
