package pagebuffer

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

func BufferedByUserId(dao *daos.Dao, userId string) ([]PageBuffer, error) {
	var pages []PageBuffer
	err := dao.ModelQuery(&PageBuffer{}).
		AndWhere(dbx.HashExp{"owner": userId}).
		All(&pages)

	return pages, err
}
