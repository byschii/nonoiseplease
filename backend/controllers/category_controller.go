package controllers

import (
	"errors"

	cats "be/model/categories"
	"be/model/page"
	users "be/model/users"

	u "be/utils"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

type CategoryController struct {
	pbDao *daos.Dao
}

type CategoryControllerInterface interface {
	DAO() *daos.Dao
	SetPBDAO(dao *daos.Dao)
	RemoveOrphanCategory(category *cats.Category) error
	RemoveOrphanCategoryWithException(category *cats.Category, expeptPageId []string) error
	AddCategoryToPage(fulltextsearchController FTSControllerInterface, owner string, pageId string, categoryName string) error
	CategoryByUser(user *users.User) ([]cats.Category, error)
}

func NewCategoryController(dao *daos.Dao) CategoryControllerInterface {
	return &CategoryController{
		pbDao: dao,
	}
}

func (controller CategoryController) DAO() *daos.Dao {
	return controller.pbDao
}

func (controller *CategoryController) SetPBDAO(dao *daos.Dao) {
	controller.pbDao = dao
}

// checks if category is linked to page
// if not it will be deleted
func (controller CategoryController) RemoveOrphanCategory(category *cats.Category) error {
	return controller.RemoveOrphanCategoryWithException(category, []string{})
}

// add a category to a page
//
//	also:
//	- check if user own page
//	- create category if not exists
//	- modify fts doc
func (controller CategoryController) AddCategoryToPage(fulltextsearchController FTSControllerInterface, owner string, pageId string, categoryName string) error {
	var page page.Page
	err := page.FillWithUserAndId(controller.DAO(), owner, pageId)
	if err != nil { // if user not own page -> page not found -> err
		return err
	}

	categoryId := ""
	categoryName = u.SanitizeString(categoryName)

	// check if category already exists
	err = controller.DAO().RunInTransaction(func(txDao *daos.Dao) error {
		dbCategory, err := cats.CategoryExistsByName(txDao, categoryName)
		if err != nil {
			return err
		}

		if dbCategory == nil {
			// create category (only way)
			categoryId, err = cats.NewCategory(txDao, categoryName)
			if err != nil {
				return err
			}
		} else {
			categoryId = dbCategory.Id
		}
		return nil
	})
	if err != nil {
		return err
	}

	// link category and page
	newCat := cats.PageToCategories{
		PageId:     pageId,
		CategoryId: categoryId,
	}
	err = newCat.Save(controller.DAO())
	if err != nil {
		return err
	}

	// modify doc in fts
	go fulltextsearchController.AlignCategoriesBetweenFTSAndDB(owner, page.FTSRef, page.Id)

	return nil
}

func (controller CategoryController) RemoveOrphanCategoryWithException(category *cats.Category, expeptPageId []string) error {
	if category == nil {
		return errors.New("category not found")
	}

	err := controller.DAO().RunInTransaction(func(txDao *daos.Dao) error {
		// if last one category (whit given name)
		wasLast, err := category.NoMoreLinksWithException(txDao, expeptPageId)
		if err != nil {
			return err
		}
		if wasLast {
			// remove category from db
			err = category.Delete(txDao)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// gets all categories associated with a page scraped by a user
func (controller CategoryController) CategoryByUser(user *users.User) ([]cats.Category, error) {
	// get every page owned by user
	var pages []page.Page
	err := controller.DAO().ModelQuery(&page.Page{}).
		AndWhere(dbx.HashExp{"owner": user.Id}).All(&pages)

	if err != nil {
		return nil, err
	}

	// get all categories
	var allCategories []cats.Category
	for _, page := range pages {
		categories, err := cats.GetCategoriesByPageId(controller.DAO(), page.Id)
		if err != nil {
			return nil, err
		}

		allCategories = append(allCategories, categories...)
	}

	// remove duplicates
	return cats.RemoveDuplicateCategory(allCategories), nil
}
