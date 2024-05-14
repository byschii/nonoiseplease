package page

import (
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

// get all pages from 'userId'
func ByUserId(dao *daos.Dao, userId string) ([]Page, error) {
	var pages []Page
	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"owner": userId}).
		All(&pages)

	return pages, err
}

// get all pages from 'userId'
func ByUserIdAndOrigin(dao *daos.Dao, userId string, originType AvailableOrigin) ([]Page, error) {

	var pages []Page
	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"owner": userId}).
		AndWhere(dbx.HashExp{"origin": originType}).
		All(&pages)

	return pages, err
}

func CountThisMonth(pages *[]Page) int {
	counter := 0
	nowMonth := time.Now().Month()
	nowYear := time.Now().Year()

	for _, page := range *pages {
		if page.Created.Time().Month() == nowMonth && page.Created.Time().Year() == nowYear {
			counter++
		}
	}

	return counter
}

// convert page id to fts id
func FromId(dao *daos.Dao, pageId string) (Page, error) {
	var page Page
	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"id": pageId}).
		One(&page)

	return page, err
}
