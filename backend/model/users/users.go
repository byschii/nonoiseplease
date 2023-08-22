package users

import (
	"github.com/pocketbase/pocketbase/models"
)

type Users struct {
	models.BaseModel

	Email    string `db:"email" json:"email"`
	Username string `db:"username" json:"username"`
	Name     string `db:"name" json:"name"`
	Avatar   string `db:"avatar" json:"avatar"`
}

func (m *Users) TableName() string {
	return "users" // the name of your collection
}

var _ models.Model = (*Users)(nil)
