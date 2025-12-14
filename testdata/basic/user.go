package basic

type User struct {
	ID       int64    `reset:""`
	Name     string   `reset:"guest"`
	Email    string   `reset:""`
	Tags     []string `reset:""`
	Active   bool     `reset:"-"`
	internal string
}

type Config struct {
	Debug    bool           `reset:""`
	Timeout  int            `reset:"30"`
	Settings map[string]any `reset:""`
}
