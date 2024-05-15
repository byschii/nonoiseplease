package pagebuffer

import (
	"github.com/pocketbase/pocketbase/models"

	pagecommons "be/pkg/page/commons"
)

/*
SAVED Page
this struct is used to store user's details
these data are ment to be low importance and editable by the user
*/
type PageBuffer struct {
	models.BaseModel

	Owner    string                      `db:"owner" json:"owner"`
	PageUrl  string                      `db:"page_url" json:"page_url"`
	Priority int                         `db:"priority" json:"priority"`
	Origin   pagecommons.AvailableOrigin `db:"origin" json:"origin"`
}

func (m *PageBuffer) TableName() string {
	return "page_buffer"
}

var _ models.Model = (*PageBuffer)(nil)
