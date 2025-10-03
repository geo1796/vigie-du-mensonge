package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the users table

type User struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	Roles []*Role `gorm:"many2many:user_roles;"`

	Tag      string `gorm:"column:tag;unique;not null"`
	Email    string `gorm:"column:email;unique;not null"`
	Password string `gorm:"column:password;not null"`

	CreatedAt time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:now()"`
}

func (User) TableName() string { return "users" }

func (u User) RoleNames() []RoleName {
	if u.Roles == nil {
		return nil
	}

	var roles []RoleName
	for _, r := range u.Roles {
		if r == nil {
			continue
		}
		roles = append(roles, r.Name)
	}
	return roles
}
