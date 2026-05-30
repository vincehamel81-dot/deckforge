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
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func New(username string) *User {
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Role:      RoleUser,
		CreatedAt: time.Now().UTC(),
	}
}
