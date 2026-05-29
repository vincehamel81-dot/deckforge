package user

import "github.com/google/uuid"

type Repository interface {
	Create(u *User) error
	FindByID(id uuid.UUID) (*User, error)
	FindByUsername(username string) (*User, error)
	ExistsByUsername(username string) (bool, error)
	Update(u *User) error
}
