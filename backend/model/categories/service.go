package categories

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"

	u "be/utils"
)

func PageToCategoriesQuery(dao *daos.Dao) *dbx.SelectQuery {
	return dao.ModelQuery(&PageToCategories{})
}

func PageToCategoryFromId(pageToCategoriesID string, dao *daos.Dao) (PageToCategories, error) {
	var pageToCategories PageToCategories
	err := PageToCategoriesQuery(dao).
		AndWhere(dbx.HashExp{"id": pageToCategoriesID}).
		One(&pageToCategories)

	return pageToCategories, err
}

func NewCategory(dao *daos.Dao, name string) (string, error) {
	cat := Category{
		Name:  name,
		Color: u.GenerateRandomHexColor(),
	}

	err := dao.Save(&cat)
	if err != nil {
		log.Println("error creating category: ", err)
		return "", err
	}

	log.Println("created category: ", cat, cat.Id)
	return cat.Id, nil
}

func CategoryQuery(dao *daos.Dao) *dbx.SelectQuery {
	return dao.ModelQuery(&Category{})
}

// kinda transfrorms a list into a set
func RemoveDuplicateCategory(sliceList []Category) []Category {
	allKeys := make(map[Category]bool)
	list := []Category{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func GetAllCategories(dao *daos.Dao) ([]Category, error) {
	var categories []Category
	err := CategoryQuery(dao).All(&categories)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func GetCategoriesByPageId(dao *daos.Dao, pageId string) ([]Category, error) {
	var categories []PageToCategories
	err := dao.ModelQuery(&PageToCategories{}).
		AndWhere(dbx.HashExp{"page_id": pageId}).
		All(&categories)

	if err != nil {
		return nil, err
	}

	// get all categories as string
	var categoriesString []string
	for _, category := range categories {
		categoriesString = append(categoriesString, category.CategoryId)
	}

	records, err := dao.FindRecordsByIds("categories", categoriesString)
	if err != nil {
		return nil, err
	}

	var returnCategories []Category
	for _, record := range records {
		// record to json
		recordJson, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}

		// json to category
		var category Category
		err = json.Unmarshal(recordJson, &category)
		if err != nil {
			return nil, err
		}
		returnCategories = append(returnCategories, category)
	}

	return returnCategories, nil
}

// Checks if a category exists
// returns 'nil' if not found by name, or found multiple times
func CategoryExistsByName(dao *daos.Dao, categoryName string) (*Category, error) {
	var category []Category
	err := CategoryQuery(dao).
		AndWhere(dbx.HashExp{"name": categoryName}).
		All(&category)

	if err != nil {
		return nil, err
	}
	if len(category) == 0 {
		return nil, nil
	}
	if len(category) > 1 {
		log.Println("multiple categories with same name")
		return nil, errors.New("multiple categories with same name")
	}

	return &category[0], nil
}

func CategoryFromId(dao *daos.Dao, categoryId string) (*Category, error) {
	var category Category
	err := CategoryQuery(dao).
		AndWhere(dbx.HashExp{"id": categoryId}).
		One(&category)

	return &category, err
}
