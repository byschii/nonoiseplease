package page

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

// get all pages from 'userId'
func GetPagesByUserId(dao *daos.Dao, userId string) ([]Page, error) {
	var pages []Page
	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"owner": userId}).
		All(&pages)

	return pages, err
}

// get all pages from 'userId'
func GetPagesByUserIdAndOrigin(dao *daos.Dao, userId string, originType AvailableOrigin) ([]Page, error) {
	var pages []Page
	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"owner": userId}).
		AndWhere(dbx.HashExp{"origin": originType}).
		All(&pages)

	return pages, err
}

// convert page id to fts id
func GetPageFromPageId(dao *daos.Dao, pageId string) (Page, error) {
	var page Page
	err := dao.ModelQuery(&Page{}).
		AndWhere(dbx.HashExp{"id": pageId}).
		One(&page)

	return page, err
}
