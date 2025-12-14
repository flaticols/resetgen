package embedded

type Metadata struct {
	CreatedAt int64 `reset:""`
	UpdatedAt int64 `reset:""`
}

type Entity struct {
	Metadata `reset:""`
	ID       string `reset:""`
	Version  int    `reset:"1"`
}

type Document struct {
	*Metadata `reset:""`
	Title     string   `reset:""`
	Tags      []string `reset:""`
}
