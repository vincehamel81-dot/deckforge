package user

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Role      Role
	CreatedAt time.Time
}

func New(username string) *User {
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Role:      RoleUser,
		CreatedAt: time.Now().UTC(),
	}
}
