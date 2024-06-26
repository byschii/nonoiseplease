package controllers

import (
	pagebuffer "be/pkg/page/buffer"
	pagecommons "be/pkg/page/commons"
	pagefts "be/pkg/page/fts"
	page "be/pkg/page/page"
	pageservice "be/pkg/page/service"
	users "be/pkg/users"
	web "be/pkg/web"
	u "be/utils"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"

	cats "be/pkg/categories"
)

type PageController struct {
	PBDao              *daos.Dao
	MeiliClient        *meilisearch.Client
	CategoryController CategoryControllerInterface
	FTSController      FTSControllerInterface
}

type PageControllerInterface interface {
	CommonController
	SetPBDAO(dao *daos.Dao)
	RemoveCategoryFromPage(pageId string, categoryName string) error
	RemoveCategoryFromPageWithOwner(owner string, pageId string, categoryName string) error
	PageSearch(query string, users []string, categories []string) (*[]web.PageResponse, error)
	PageID2FullPageData(owner string, pageId string) (*page.Page, []cats.Category, *pagefts.FTSPageDoc, error)
	SaveNewPage(owner string,
		url string,
		pageTitle string,
		categories []string,
		content string,
		originType pagecommons.AvailableOrigin,
		withProxy bool) (string, error)
	RemoveDocFTSIndex(pageId string) error
	FindCategoriesFromUser(user *users.User) ([]cats.Category, error)
	AddCategoryToPage(owner string, pageId string, categoryName string) error
	SetDBCategoriesOnFTSDoc(owner string, FTSRef string, categories []cats.Category) error
	AddToBuffer(owner string, url string, priority int, origin pagecommons.AvailableOrigin) error
}

func NewPageController(dao *daos.Dao, meiliClient *meilisearch.Client, categoryController CategoryControllerInterface, fulltextsearchController FTSControllerInterface) PageControllerInterface {
	return &PageController{
		PBDao:              dao,
		MeiliClient:        meiliClient,
		CategoryController: categoryController,
		FTSController:      fulltextsearchController,
	}
}

func (controller *PageController) SetDBCategoriesOnFTSDoc(owner string, FTSRef string, categories []cats.Category) error {
	return controller.FTSController.SetDBCategoriesOnFTSDoc(owner, FTSRef, categories)
}

func (controller *PageController) AddCategoryToPage(owner string, pageId string, categoryName string) error {
	return controller.CategoryController.AddCategoryToPage(controller.FTSController, owner, pageId, categoryName)
}

func (controller *PageController) FindCategoriesFromUser(user *users.User) ([]cats.Category, error) {
	return controller.CategoryController.CategoryByUser(user)
}

func (controller *PageController) RemoveDocFTSIndex(pageId string) error {
	return controller.FTSController.RemoveDocFTSIndex(pageId)
}

func (controller *PageController) AppDao() *daos.Dao {
	return controller.PBDao
}

func (controller *PageController) SetPBDAO(dao *daos.Dao) {
	controller.PBDao = dao
}

func (controller PageController) RemoveCategoryFromPage(pageId string, categoryName string) error {
	page, err := page.FromId(controller.PBDao, pageId)
	if err != nil {
		return err
	}

	return controller.RemoveCategoryFromPageWithOwner(page.Owner, pageId, categoryName)
}

// delete category just for this page
// (it s not destroing all categories with this name on other pages)
func (controller PageController) RemoveCategoryFromPageWithOwner(owner string, pageId string, categoryName string) error {
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

	log.Debug().Msgf("check if last category %s %s", pageId, category.Name)
	err = controller.CategoryController.RemoveOrphanCategory(category)
	if err != nil {
		return err
	}

	// remove category from fts doc
	go controller.FTSController.AlignCategoriesBetweenFTSAndDB(
		owner, page.FTSRef, page.Id,
	)

	return nil
}

// categories comes ','-separated
func (controller PageController) PageSearch(query string, users []string, categories []string) (*[]web.PageResponse, error) {

	// parse categories array (if not empty)
	// into meili search filter query
	var categoryFilterQuery string
	if len(categories) > 0 && categories[0] != "" {
		for i, category := range categories {
			categories[i] = "category = '" + category + "' "
		}
		categoryFilterQuery = strings.Join(categories, " AND ")
	}
	log.Debug().Msgf(categoryFilterQuery)

	// build meili search request
	var meiliSearchRequest meilisearch.SearchRequest
	if categoryFilterQuery != "" {
		meiliSearchRequest = meilisearch.SearchRequest{
			Filter: categoryFilterQuery,
		}
	} else {
		meiliSearchRequest = meilisearch.SearchRequest{}
	}

	var pages []web.PageResponse
	for _, userIndex := range users {
		resp, err := controller.MeiliClient.Index(userIndex).Search(query, &meiliSearchRequest)
		if err != nil {
			return nil, u.WrapError("couldnt search "+userIndex+" index", err)
		}
		for _, hit := range resp.Hits {
			result, err := controller.parseMeiliResponse(hit)
			if err != nil {
				return nil, u.WrapError("couldnt parse meili response", err)
			}
			pages = append(pages, result)

		}
	}

	return &pages, nil
}

func (controller PageController) parseMeiliResponse(hit interface{}) (web.PageResponse, error) {
	pageResp := web.PageResponse{}

	// get doc
	doc, err := pagefts.FromMeiliResultInterface(hit)
	if err != nil {
		return pageResp, u.WrapError("couldnt convert meili result to fts doc", err)
	}

	// get page
	var page page.Page
	err = page.FillWithRef(controller.PBDao, doc.ID)
	if err != nil {
		return pageResp, u.WrapError("couldnt get page with ref "+doc.ID, err)
	}

	// get categories
	categories, err := cats.GetCategoriesByPageId(controller.PBDao, page.Id)
	if err != nil {
		return pageResp, u.WrapError("couldnt get categories for page "+page.Id, err)
	}

	// build response
	pageResp.Page = page
	pageResp.Categories = categories
	pageResp.FTSDoc = doc

	return pageResp, nil
}

func (controller PageController) PageID2FullPageData(owner string, pageId string) (*page.Page, []cats.Category, *pagefts.FTSPageDoc, error) {

	// get page
	var simplePage page.Page
	err := simplePage.FillWithUserAndId(controller.PBDao, owner, pageId)
	if err != nil {
		return nil, nil, nil, err
	}

	// get all categories for the page
	categories, err := cats.GetCategoriesByPageId(controller.PBDao, pageId)
	if err != nil {
		return nil, nil, nil, u.WrapError("on categories", err)
	}

	// get fts doc
	ftsDoc, err := pagefts.FromIndexAndRef(controller.MeiliClient, owner, simplePage.FTSRef)
	if err != nil {
		return nil, nil, nil, err
	}

	return &simplePage, categories, &ftsDoc, nil
}

func (controller PageController) SaveNewPage(
	owner string,
	url string,
	pageTitle string,
	categories []string,
	content string,
	originType pagecommons.AvailableOrigin,
	withProxy bool) (string, error) {

	//page.service
	reference, err := pageservice.SaveNewPage(
		controller.PBDao,
		controller.MeiliClient,
		owner,
		url,
		pageTitle,
		categories,
		content,
		originType,
		withProxy,
	)

	return reference, err
}

func (controller PageController) AddToBuffer(owner string, url string, priority int, origin pagecommons.AvailableOrigin) error {
	buffer := pagebuffer.PageBuffer{
		Owner:    owner,
		PageUrl:  url,
		Priority: priority,
		Origin:   origin,
	}
	err := controller.PBDao.Save(&buffer)

	if err == nil {
		return nil
	}
	return fmt.Errorf("couldnt save buffer url %s (%s)", url, err)
}
