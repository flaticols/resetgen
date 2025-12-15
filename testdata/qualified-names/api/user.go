//go:generate go run ../../.. -structs models.User,api.User

package api

type User struct {
	ID       string `reset:""`
	Status   string `reset:"active"`
	Metadata map[string]string `reset:""`
}
