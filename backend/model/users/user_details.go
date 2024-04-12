package users

import (
	"github.com/pocketbase/pocketbase/models"
)

/*
USER DETAILS
this struct is used to store user's details
these data are ment to be low importance and editable by the user
*/
type UserDetails struct {
	models.BaseModel

	Nickname    string `db:"nickname" json:"nickname"`
	RelatedUser string `db:"related_user" json:"related_user"`
}

func (m *UserDetails) TableName() string {
	return "user_details" // the name of your collection
}

var _ models.Model = (*UserDetails)(nil)
