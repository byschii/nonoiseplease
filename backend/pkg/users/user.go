package users

import (
	"github.com/pocketbase/pocketbase/models"
)

type User struct {
	models.BaseModel

	Email    string `db:"email" json:"email"`
	Username string `db:"username" json:"username"`
	Name     string `db:"name" json:"name"`
	Avatar   string `db:"avatar" json:"avatar"`
}

func (m *User) TableName() string {
	return "users" // the name of your collection
}

var _ models.Model = (*User)(nil)
