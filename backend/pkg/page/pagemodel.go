package page

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// activity to category mapping
type AvailableOrigin string

const (
	AvailableOriginScrape    AvailableOrigin = "scrape"
	AvailableOriginExtention AvailableOrigin = "extention"
)

/*
SAVED Page
this struct is used to store user's details
these data are ment to be low importance and editable by the user
*/
type Page struct {
	models.BaseModel

	Owner     string          `db:"owner" json:"owner"`
	Link      string          `db:"link" json:"link"`
	PageTitle string          `db:"page_title" json:"page_title"`
	FTSRef    string          `db:"fts_ref" json:"fts_ref"`
	Votes     int             `db:"votes" json:"votes"`
	WithProxy bool            `db:"with_proxy" json:"with_proxy"`
	Origin    AvailableOrigin `db:"origin" json:"origin"`
}

func (m *Page) TableName() string {
	return "pages"
}

var _ models.Model = (*Page)(nil)

// returns page 'pageId' only if it belongs to user 'userId'
func (page *Page) FillWithUserAndId(dao *daos.Dao, userId string, pageId string) error {

	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"owner": userId}).
		AndWhere(dbx.HashExp{"id": pageId}).
		One(&page)

	return err
}

func (page *Page) FillWithId(dao *daos.Dao, pageId string) error {

	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"id": pageId}).
		One(&page)

	return err
}

func (page *Page) FillWithRef(dao *daos.Dao, ref string) error {

	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"fts_ref": ref}).
		One(&page)

	return err
}
