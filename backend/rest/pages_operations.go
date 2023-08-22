package rest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	cats "be/model/categories"
	tfs_page_doc "be/model/fts_page_doc"
	page "be/model/page"

	u "be/utils"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

func SaveNewPage(
	owner string,
	url string,
	pageTitle string,
	categories []string,
	content string,
	originType page.AvailableOrigin,
	withProxy bool,
	meiliClient *meilisearch.Client,
	dao *daos.Dao) (string, error) {
	// create doc id with sha
	hash := sha256.New()
	hash.Write([]byte(url + pageTitle + owner + fmt.Sprint(rand.Intn(1000000)) + time.Now().String()))
	reference := hex.EncodeToString(hash.Sum(nil))

	// checking for errors
	errs := make(chan error, 2)

	go func() {
		savedArticle := page.Page{
			Link:      url,
			PageTitle: pageTitle,
			Owner:     owner,
			FTSRef:    reference,
			Votes:     0,
			WithProxy: withProxy,
			Origin:    originType,
		}
		errs <- dao.Save(&savedArticle)
	}()

	go func() {
		doc := tfs_page_doc.FTSPageDoc{
			ID:       reference,
			Category: []string{},
			Content:  content,
		}
		errs <- doc.Save(meiliClient, owner)
	}()

	// maybe errored
	for i := 0; i < 2; i++ {
		err := <-errs
		if err != nil {
			return "", err
		}
	}

	return reference, nil
}

func GetFullPageDataByID(owner string, pageId string, meiliClient *meilisearch.Client, dao *daos.Dao) (*page.Page, []cats.Category, *tfs_page_doc.FTSPageDoc, error) {

	// get page
	var page page.Page
	err := page.FillWithUserAndId(dao, owner, pageId)
	if err != nil {
		return nil, nil, nil, err
	}

	// get all categories for the page
	categories, err := cats.GetCategoriesByPageId(dao, pageId)
	if err != nil {
		return nil, nil, nil, u.WrapError("on categories", err)
	}

	// get fts doc
	ftsDoc, err := tfs_page_doc.FromIndexAndRef(meiliClient, owner, page.FTSRef)
	if err != nil {
		return nil, nil, nil, err
	}

	return &page, categories, &ftsDoc, nil
}

func DeleteCategoryFromPageToCategory(ptcId string, meiliClient *meilisearch.Client, dao *daos.Dao) error {
	ptc, err := cats.PageToCategoryFromId(ptcId, dao)
	if err != nil {
		return err
	}

	category, err := cats.CategoryFromId(dao, ptc.CategoryId)
	if err != nil {
		return err
	}
	return DeleteCategoryFromPage(ptc.PageId, category.Name, meiliClient, dao)
}

func DeleteCategoryFromPage(pageId string, categoryName string, meiliClient *meilisearch.Client, dao *daos.Dao) error {
	page, err := page.GetPageFromPageId(dao, pageId)
	if err != nil {
		return err
	}

	return DeleteCategoryFromPageWithOwner(page.Owner, pageId, categoryName, meiliClient, dao)

}

func DeleteCategoryFromPageWithOwner(owner string, pageId string, categoryName string, meiliClient *meilisearch.Client, dao *daos.Dao) error {
	var page page.Page
	err := page.FillWithUserAndId(dao, owner, pageId)
	if err != nil {
		return err
	}

	// get category
	category, err := cats.CategoryExistsByName(dao, categoryName)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("category not found")
	}

	// remove link between page and category (for sure)
	var link cats.PageToCategories
	err = link.FillWithPageIdAndCategoryId(dao, pageId, category.Id)
	if err != nil {
		return err
	}
	err = link.Delete(dao)
	if err != nil {
		return err
	}

	log.Print("check if last category", pageId, category.Name)
	err = DeleteOrphanCategory(category, dao)
	if err != nil {
		return err
	}

	// remove category from fts doc
	go AlignCategoriesBetweenFTSAndDB(meiliClient, owner, page.FTSRef, page.Id, dao)

	return nil
}

func DeleteOrphanCategoryFromPTCId(ptcId string, dao *daos.Dao) error {
	ptc, err := cats.PageToCategoryFromId(ptcId, dao)
	if err != nil {
		return err
	}
	return DeleteOrphanCategoryFromCategoryId(ptc.CategoryId, dao)
}

func DeleteOrphanCategoryFromCategoryId(categoryId string, dao *daos.Dao) error {
	category, err := cats.CategoryFromId(dao, categoryId)
	if err != nil {
		return err
	}
	return DeleteOrphanCategory(category, dao)
}

// checks if category is linked to page
// if not it will be deleted
func DeleteOrphanCategory(category *cats.Category, dao *daos.Dao) error {
	return DeleteOrphanCategoryWithException(category, dao, []string{})
}

func DeleteOrphanCategoryWithException(category *cats.Category, dao *daos.Dao, expeptPageId []string) error {
	if category == nil {
		return errors.New("category not found")
	}

	err := dao.RunInTransaction(func(txDao *daos.Dao) error {
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

func DeleteDocFTSIndex(meiliClient *meilisearch.Client, dao *daos.Dao, pageId string) error {

	log.Print("deleting " + pageId)
	// convert docId to ftsRef
	page, err := page.GetPageFromPageId(dao, pageId)
	if err != nil {
		return err
	}
	_, err = meiliClient.Index(page.Owner).DeleteDocument(page.FTSRef)
	if err != nil {
		return err
	}

	return nil
}

// add a category to a page
//
//	also:
//	- check if user own page
//	- create category if not exists
//	- modify fts doc
func AddCategoryToPage(owner string, pageId string, categoryName string, meiliClient *meilisearch.Client, dao *daos.Dao) error {
	var page page.Page
	err := page.FillWithUserAndId(dao, owner, pageId)
	if err != nil { // if user not own page -> page not found -> err
		return err
	}

	categoryId := ""
	categoryName = u.SanitizeString(categoryName)

	// check if category already exists
	err = dao.RunInTransaction(func(txDao *daos.Dao) error {
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
	err = newCat.Save(dao)
	if err != nil {
		return err
	}

	// modify doc in fts
	go AlignCategoriesBetweenFTSAndDB(meiliClient, owner, page.FTSRef, page.Id, dao)

	return nil
}

func GetCategoriesByUserId(dao *daos.Dao, pageId string) ([]cats.Category, error) {
	// get every page owner dy user
	var pages []page.Page
	err := dao.ModelQuery(&page.Page{}).
		AndWhere(dbx.HashExp{"owner": pageId}).All(&pages)

	if err != nil {
		return nil, err
	}

	// get all categories
	var allCategories []cats.Category
	for _, page := range pages {
		categories, err := cats.GetCategoriesByPageId(dao, page.Id)
		if err != nil {
			return nil, err
		}

		allCategories = append(allCategories, categories...)
	}

	// remove duplicates
	return cats.RemoveDuplicateCategory(allCategories), nil
}

func AlignCategoriesBetweenFTSAndDB(meiliClient *meilisearch.Client, owner string, FTSRef string, pageId string, dao *daos.Dao) error {
	cateories, err := cats.GetCategoriesByPageId(dao, pageId)
	if err != nil {
		log.Printf("error while getting categories for page %s: %s , cannot align db e fts", pageId, err.Error())
		return err
	}
	return SetDBCategoriesForFTSDoc(meiliClient, owner, FTSRef, cateories)
}

func SetDBCategoriesForFTSDoc(meiliClient *meilisearch.Client, owner string, FTSRef string, categories []cats.Category) error {
	// convert categories to string slice
	var categoryNames []string
	for _, category := range categories {
		categoryNames = append(categoryNames, category.Name)
	}
	err := tfs_page_doc.SetCategoriesForFTSDoc(meiliClient, owner, FTSRef, categoryNames)
	if err != nil {
		log.Printf("error while setting categories for doc %s: %s , cannot align db e fts", FTSRef, err.Error())
	}
	return err
}

func CountUserPagesByOrigin(userid string, dao *daos.Dao, originType page.AvailableOrigin) (int, error) {
	pages, err := page.GetPagesByUserIdAndOrigin(dao, userid, originType)
	if err != nil {
		return 0, err
	}

	return len(pages), nil
}

func CountUserPagesByOriginThisMonth(userid string, dao *daos.Dao, originType page.AvailableOrigin) (int, error) {
	pages, err := page.GetPagesByUserIdAndOrigin(dao, userid, originType)
	if err != nil {
		return 0, err
	}

	counter := 0
	for _, page := range pages {
		if page.Created.Time().Month() == time.Now().Month() && page.Created.Time().Year() == time.Now().Year() {
			counter++
		}
	}

	return counter, nil
}

// categories comes ','-separated
func PageSearch(query string, users []string, categories []string, meiliClient *meilisearch.Client, dao *daos.Dao) (*[]PageResponse, error) {

	// parse categories array (if not empty)
	// into meili search filter query
	var categoryFilterQuery string
	if len(categories) > 0 && categories[0] != "" {
		for i, category := range categories {
			categories[i] = "category = '" + category + "' "
		}
		categoryFilterQuery = strings.Join(categories, " AND ")
	}
	log.Println(categoryFilterQuery)

	var meiliSearchRequest meilisearch.SearchRequest
	if categoryFilterQuery != "" {
		meiliSearchRequest = meilisearch.SearchRequest{
			Filter: categoryFilterQuery,
		}
	} else {
		meiliSearchRequest = meilisearch.SearchRequest{}
	}

	var pages []PageResponse
	for _, userIndex := range users {
		resp, err := meiliClient.Index(userIndex).Search(query, &meiliSearchRequest)
		if err != nil {
			return nil, u.WrapError("couldnt search "+userIndex+" index", err)
		}
		for _, hit := range resp.Hits {
			// get doc
			doc, err := tfs_page_doc.FromMeiliResultInterface(hit)
			if err != nil {
				return nil, u.WrapError("couldnt convert meili result to fts doc", err)
			}

			// get page
			var page page.Page
			err = page.FillWithRef(dao, doc.ID)
			if err != nil {
				return nil, u.WrapError("couldnt get page with ref "+doc.ID, err)
			}

			// get categories
			categories, err := cats.GetCategoriesByPageId(dao, page.Id)
			if err != nil {
				return nil, u.WrapError("couldnt get categories for page "+page.Id, err)
			}

			result := PageResponse{
				Page:       page,
				Categories: categories,
				FTSDoc:     doc,
			}
			pages = append(pages, result)
		}
	}

	return &pages, nil
}
