package controllers

import (
	page "be/model/page"
	"be/model/rest"
	u "be/utils"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"

	cats "be/model/categories"
	tfs_page_doc "be/model/fts_page_doc"
)

type PageController struct {
	PBDao              *daos.Dao
	MeiliClient        *meilisearch.Client
	CategoryController *CategoryController
	FTSController      *FTSController
}

func (controller PageController) CountUserPagesByOriginThisMonth(userid string, originType page.AvailableOrigin) (int, error) {
	pages, err := page.GetPagesByUserIdAndOrigin(controller.PBDao, userid, originType)
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

func (controller PageController) DeleteCategoryFromPage(pageId string, categoryName string) error {
	page, err := page.GetPageFromPageId(controller.PBDao, pageId)
	if err != nil {
		return err
	}

	return controller.DeleteCategoryFromPageWithOwner(page.Owner, pageId, categoryName)
}

// delete category just for this page
// (it s not destroing all categories with this name on other pages)
func (controller PageController) DeleteCategoryFromPageWithOwner(owner string, pageId string, categoryName string) error {
	var page page.Page
	err := page.FillWithUserAndId(controller.PBDao, owner, pageId)
	if err != nil {
		return err
	}

	// get category
	category, err := cats.CategoryExistsByName(controller.PBDao, categoryName)
	if err != nil {
		return err
	}
	if category == nil {
		return errors.New("category not found")
	}

	// remove link between page and category (for sure)
	var link cats.PageToCategories
	err = link.FillWithPageIdAndCategoryId(controller.PBDao, pageId, category.Id)
	if err != nil {
		return err
	}
	err = link.Delete(controller.PBDao)
	if err != nil {
		return err
	}

	log.Print("check if last category", pageId, category.Name)
	err = controller.CategoryController.DeleteOrphanCategory(category)
	if err != nil {
		return err
	}

	// remove category from fts doc
	go controller.FTSController.AlignCategoriesBetweenFTSAndDB(owner, page.FTSRef, page.Id)

	return nil
}

// categories comes ','-separated
func (controller PageController) PageSearch(query string, users []string, categories []string) (*[]rest.PageResponse, error) {

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

	var pages []rest.PageResponse
	for _, userIndex := range users {
		resp, err := controller.MeiliClient.Index(userIndex).Search(query, &meiliSearchRequest)
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
			err = page.FillWithRef(controller.PBDao, doc.ID)
			if err != nil {
				return nil, u.WrapError("couldnt get page with ref "+doc.ID, err)
			}

			// get categories
			categories, err := cats.GetCategoriesByPageId(controller.PBDao, page.Id)
			if err != nil {
				return nil, u.WrapError("couldnt get categories for page "+page.Id, err)
			}

			result := rest.PageResponse{
				Page:       page,
				Categories: categories,
				FTSDoc:     doc,
			}
			pages = append(pages, result)
		}
	}

	return &pages, nil
}

func (controller PageController) GetFullPageDataByID(owner string, pageId string) (*page.Page, []cats.Category, *tfs_page_doc.FTSPageDoc, error) {

	// get page
	var page page.Page
	err := page.FillWithUserAndId(controller.PBDao, owner, pageId)
	if err != nil {
		return nil, nil, nil, err
	}

	// get all categories for the page
	categories, err := cats.GetCategoriesByPageId(controller.PBDao, pageId)
	if err != nil {
		return nil, nil, nil, u.WrapError("on categories", err)
	}

	// get fts doc
	ftsDoc, err := tfs_page_doc.FromIndexAndRef(controller.MeiliClient, owner, page.FTSRef)
	if err != nil {
		return nil, nil, nil, err
	}

	return &page, categories, &ftsDoc, nil
}

func (controller PageController) SaveNewPage(owner string,
	url string,
	pageTitle string,
	categories []string,
	content string,
	originType page.AvailableOrigin,
	withProxy bool) (string, error) {

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
		errs <- controller.PBDao.Save(&savedArticle)
	}()

	go func() {
		doc := tfs_page_doc.FTSPageDoc{
			ID:       reference,
			Category: []string{},
			Content:  content,
		}
		errs <- doc.Save(controller.MeiliClient, owner)
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
