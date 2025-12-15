//go:generate go run ../.. -structs User,Order

package structsflag

type User struct {
	ID     int64
	Name   string
	Email  string
	Secret string `reset:"-"` // Should respect the ignore tag even with -structs
}

type Order struct {
	ID    int64
	Total float64 `reset:"0.0"` // Should respect custom value
	Items []string
}

type Logger struct {
	Level string
} // Should NOT be generated (not in -structs list)

type Config struct {
	Host string `reset:""`
	Port int    `reset:"8080"`
} // Has tags but NOT in -structs list - should NOT be generated
