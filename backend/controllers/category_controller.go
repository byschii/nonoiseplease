package controllers

import (
	"errors"

	cats "be/model/categories"
	"be/model/page"

	u "be/utils"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

type CategoryController struct {
	PBDao *daos.Dao
}

func (controller CategoryController) DeleteOrphanCategoryFromCategoryId(categoryId string) error {
	category, err := cats.CategoryFromId(controller.PBDao, categoryId)
	if err != nil {
		return err
	}
	return controller.DeleteOrphanCategory(category)
}

// checks if category is linked to page
// if not it will be deleted
func (controller CategoryController) DeleteOrphanCategory(category *cats.Category) error {
	return controller.DeleteOrphanCategoryWithException(category, []string{})
}

// add a category to a page
//
//	also:
//	- check if user own page
//	- create category if not exists
//	- modify fts doc
func (controller CategoryController) AddCategoryToPage(fulltextsearchController FTSController, owner string, pageId string, categoryName string) error {
	var page page.Page
	err := page.FillWithUserAndId(controller.PBDao, owner, pageId)
	if err != nil { // if user not own page -> page not found -> err
		return err
	}

	categoryId := ""
	categoryName = u.SanitizeString(categoryName)

	// check if category already exists
	err = controller.PBDao.RunInTransaction(func(txDao *daos.Dao) error {
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
	err = newCat.Save(controller.PBDao)
	if err != nil {
		return err
	}

	// modify doc in fts
	go fulltextsearchController.AlignCategoriesBetweenFTSAndDB(owner, page.FTSRef, page.Id)

	return nil
}

func (controller CategoryController) DeleteOrphanCategoryWithException(category *cats.Category, expeptPageId []string) error {
	if category == nil {
		return errors.New("category not found")
	}

	err := controller.PBDao.RunInTransaction(func(txDao *daos.Dao) error {
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

func (controller CategoryController) GetCategoriesByUserId(pageId string) ([]cats.Category, error) {
	// get every page owned by user
	var pages []page.Page
	err := controller.PBDao.ModelQuery(&page.Page{}).
		AndWhere(dbx.HashExp{"owner": pageId}).All(&pages)

	if err != nil {
		return nil, err
	}

	// get all categories
	var allCategories []cats.Category
	for _, page := range pages {
		categories, err := cats.GetCategoriesByPageId(controller.PBDao, page.Id)
		if err != nil {
			return nil, err
		}

		allCategories = append(allCategories, categories...)
	}

	// remove duplicates
	return cats.RemoveDuplicateCategory(allCategories), nil
}
