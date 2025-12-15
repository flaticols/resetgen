//go:generate go run ../../.. -structs models.User,api.User

package models

type User struct {
	ID    int64  `reset:""`
	Name  string `reset:""`
	Email string `reset:""`
}
