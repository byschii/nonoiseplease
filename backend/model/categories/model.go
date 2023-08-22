package categories

import (
	"database/sql"
	"errors"

	u "be/utils"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type PageToCategories struct {
	models.BaseModel

	PageId     string `db:"page_id" json:"page_id"`
	CategoryId string `db:"category_id" json:"category_id"`
}

func (m *PageToCategories) TableName() string {
	return "page_to_categories"
}

var _ models.Model = (*PageToCategories)(nil)

func (m *PageToCategories) Delete(dao *daos.Dao) error {
	return dao.Delete(m)
}

func (m *PageToCategories) Save(dao *daos.Dao) error {
	return dao.Save(m)
}

func (l *PageToCategories) FillWithPageIdAndCategoryId(dao *daos.Dao, pageId string, categoryId string) error {

	err := dao.ModelQuery(&PageToCategories{}).
		AndWhere(dbx.HashExp{"page_id": pageId}).
		AndWhere(dbx.HashExp{"category_id": categoryId}).
		One(&l)

	if err != nil {
		return err
	}

	return nil
}

type Category struct {
	models.BaseModel

	Name  string `db:"name" json:"name"`
	Color string `db:"color" json:"color"`
}

func (m *Category) TableName() string {
	return "categories"
}

var _ models.Model = (*Category)(nil)

func (c Category) GetLinks(dao *daos.Dao) ([]PageToCategories, error) {
	var links []PageToCategories
	err := dao.ModelQuery(&PageToCategories{}).
		AndWhere(dbx.HashExp{"category_id": c.Id}).
		All(&links)

	return links, err
}

// return 'true' if a category is "dead" (not linked to any page)
func (c Category) NoMoreLinks(dao *daos.Dao) (bool, error) {
	link, err := c.GetLinks(dao)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		} else {
			return false, err
		}
	}

	return len(link) == 0, nil

}

func (c Category) NoMoreLinksWithException(dao *daos.Dao, exceptPageId []string) (bool, error) {
	links, err := c.GetLinks(dao)

	if err != nil {
		// if no links, return true
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		} else {
			return false, err
		}
	}

	// if no links
	// clearly no more links
	// return true
	if len(links) == 0 {
		return true, nil
	}

	// check if all link.PageId are also in exceptPageId
	everyLinkInException := true
	for _, l := range links {
		if !u.InList(l.PageId, exceptPageId) {
			everyLinkInException = false
			break
		}
	}
	return everyLinkInException, nil
}

func (c Category) Delete(dao *daos.Dao) error {
	// check if category is linked to any page
	dead, err := c.NoMoreLinks(dao)
	if err != nil {
		return err
	}
	if !dead {
		return errors.New("category still linked to page(s)")
	}

	return dao.Delete(&c)
}
