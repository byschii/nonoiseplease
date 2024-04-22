package controllers

import (
	u "be/utils"
	"be/webscraping"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"be/model/page"
	rest "be/model/rest"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/models"
)

type WebController struct {
	PageController   PageControllerInterface
	UserController   UserControllerInterface
	ConfigController ConfigControllerInterface
}

type NoNoiseInterface interface {
	GetSearch(c echo.Context) error
	GetSearchExtentionHtml(c echo.Context) error
	GetSearchInfo(c echo.Context) error
	DeleteAccount(c echo.Context) error
	PostUrlScrape(c echo.Context) error
	PostPagemanageCategory(c echo.Context) error
	GetPagemanage(c echo.Context) error
	DeletePagemanageCategory(c echo.Context) error
	PostPagemanageLoad(c echo.Context) error
	DeletePagemanagePage(c echo.Context) error
}

func NewNoNoiseInterface(pageController PageControllerInterface, userController UserControllerInterface, configController ConfigControllerInterface) NoNoiseInterface {
	return &WebController{
		PageController:   pageController,
		UserController:   userController,
		ConfigController: configController,
	}
}

func (controller WebController) DeleteAccount(c echo.Context) error {
	// retrive user id from get params
	record, _ := controller.UserController.UserRecordFromRequest(c, true)

	controller.UserController.DropAccount(record)
	return c.NoContent(http.StatusOK)
}

/*
GetSearchInfo
used to get all info needed for search page before actually searching

right now it only returns all categories from user pages (cause a user can filter by category)
*/
func (controller WebController) GetSearchInfo(c echo.Context) error {

	user, err := controller.UserController.UserFromRequest(c, false)
	if err != nil {
		log.Debug().Msgf("failed to get user from request, %v\n", err)

		return c.String(http.StatusBadRequest, u.WrapError("failed to get user from request ", err).Error())
	}

	// get all categories from user pages
	categories, err := controller.PageController.FindCategoriesFromUser(user)
	if err != nil {
		log.Debug().Msgf("failed to get categories, %v\n", err)

		return c.String(http.StatusBadRequest, u.WrapError("failed to get categories ", err).Error())
	}

	preSearchInfo := rest.PreSearchInfoResponse{
		Categories: categories,
	}
	c.JSON(http.StatusOK, preSearchInfo)

	return nil
}

func (controller WebController) SearchPages(c echo.Context) (rest.SearchResponse, error) {
	// print request
	log.Debug().Msgf(c.Request().URL.String())
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		return rest.SearchResponse{}, errors.New("unauthorized, user not verified")
	}
	userID := record.Id

	// read query and list of categories from url params
	query := c.QueryParam("query")
	categoriesParam := c.QueryParam("categories")

	// split categories over comma
	categories := strings.Split(categoriesParam, ",")

	pageResp, err := controller.PageController.PageSearch(query, []string{userID}, categories)
	if err != nil {
		log.Error().Err(err).Msg("failed to search ")
		return rest.SearchResponse{}, u.WrapError("failed to search ", err)
	}

	resp := rest.SearchResponse{
		Pages: *pageResp,
	}
	return resp, nil
}

// used by extention to get html for search page
// this allows to have a search page in the extention
// without having to create a new page in the extention and without having to update the extention to deliver updates
func (controller WebController) GetSearchExtentionHtml(c echo.Context) error {
	data, err := controller.SearchPages(c)
	if err != nil {
		log.Error().Err(err).Msg("failed to search")
		return c.String(http.StatusBadRequest, u.WrapError("failed to search ", err).Error())
	}

	// open fs file
	templateName := "./extention_template/search.html"
	parsedSearchTemplate := template.Must(template.ParseFiles(templateName))

	// execute template
	err = parsedSearchTemplate.Execute(c.Response().Writer, data)
	if err != nil {
		log.Error().Err(err).Msg("error executing template")
		return c.String(http.StatusInternalServerError, u.WrapError("error executing template ", err).Error())
	}

	return nil
}

func (controller WebController) GetSearch(c echo.Context) error {
	data, err := controller.SearchPages(c)
	if err != nil {
		log.Error().Err(err).Msg("failed to search")
		return c.String(http.StatusBadRequest, u.WrapError("failed to search ", err).Error())
	}

	return c.JSON(http.StatusOK, data)
}

func (controller WebController) PostUrlScrape(c echo.Context) error {
	// retrive user id from get params
	userRecord, _ := c.Get("authRecord").(*models.Record)
	if userRecord == nil || !userRecord.GetBool("verified") {
		return c.String(http.StatusUnauthorized, "unauthorized, user not verified")
	}

	pagesAlreadyScraped, err := controller.PageController.CountUserPagesByOriginThisMonth(userRecord.Id, page.AvailableOriginScrape)
	if err != nil {
		log.Debug().Msgf("failed to count user pages, %v\n", err)
		return c.String(http.StatusInternalServerError, "failed to count user pages")
	}
	// if user has reached the limit, return error
	if pagesAlreadyScraped >= controller.ConfigController.MaxScrapePerMonth() {

		return c.String(http.StatusForbidden, "you have reached the limit of pages you can scrape")
	}

	// get url from json body
	var urlData rest.Url
	if err := c.Bind(&urlData); err != nil {
		log.Debug().Msgf("failed to parse json body, %v\n", err)

		return c.String(http.StatusBadRequest, "failed to parse json body")
	}
	// scrape url and get info
	article, withProxy, err := webscraping.GetArticle(urlData.Url, false, controller.PageController.AppDao())
	if err != nil {
		log.Debug().Msgf("failed to parse %s, %v\n", urlData.Url, err)
		return c.String(http.StatusBadRequest, "failed to parse url")
	}

	meili_ref, err := controller.PageController.SaveNewPage(
		userRecord.Id, urlData.Url, article.Title, []string{}, article.TextContent, page.AvailableOriginScrape, withProxy,
	)
	if err != nil {
		log.Debug().Msgf("failed to save page, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
	}

	return c.String(http.StatusOK, meili_ref)
}

func (controller WebController) PostPagemanageCategory(c echo.Context) error {
	// retrive user id from get params
	userRecord, _ := c.Get("authRecord").(*models.Record)
	if userRecord == nil || !userRecord.GetBool("verified") {
		return c.String(http.StatusUnauthorized, "unauthorized, user not verified")
	}

	// read page and category id from body
	var data rest.PostPagemanageCategoryRequest
	if err := c.Bind(&data); err != nil {
		log.Debug().Msgf("failed to parse json body, %v\n", err)
		return c.String(http.StatusBadRequest, "failed to parse json body")
	}

	log.Debug().Msgf("data %+v", data)
	err := controller.PageController.AddCategoryToPage(userRecord.Id, data.PageID, data.CategoryName)

	if err != nil {
		log.Debug().Msgf("failed to add category, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to add category ", err).Error())
	}

	return c.NoContent(http.StatusOK)
}

func (controller WebController) GetPagemanage(c echo.Context) error {
	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		return c.String(http.StatusUnauthorized, "unauthorized, user not verified")
	}
	userID := record.Id

	// read page id from params
	pageID := c.QueryParam("id")

	page, categories, ref, err := controller.PageController.PageID2FullPageData(userID, pageID)
	if err != nil || ref == nil {
		log.Debug().Msgf("failed to get page, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to get page ", err).Error())
	}

	// check same categories in meili and db
	// if not (send notification) ... update meili
	sameCats := true
	for _, dbCat := range categories {
		if ref.Category == nil || len(ref.Category) == 0 {
			sameCats = false
			break
		}
		for _, meiliCat := range ref.Category {
			if dbCat.Name == meiliCat {
				continue
			} else {
				sameCats = false
			}
		}
	}
	if !sameCats {
		controller.PageController.SetDBCategoriesOnFTSDoc(userID, page.FTSRef, categories)
	}

	result := rest.PageResponse{
		Page:       *page,
		Categories: categories,
		FTSDoc:     *ref,
	}

	return c.JSON(http.StatusOK, result)
}

func (controller WebController) DeletePagemanageCategory(c echo.Context) error {
	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		return c.String(http.StatusUnauthorized, "unauthorized, user not verified")
	}
	userID := record.Id

	// read page and category id from body
	var data rest.DeleteCategoryRequest
	if err := c.Bind(&data); err != nil {
		log.Debug().Msgf("failed to parse json body, %v\n", err)
		return c.String(http.StatusBadRequest, "failed to parse json body")
	}

	err := controller.PageController.RemoveCategoryFromPageWithOwner(userID, data.PageID, data.CategoryName)
	if err != nil {
		log.Debug().Msgf("failed to delete category, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
	}

	return c.NoContent(http.StatusOK)
}

func (controller WebController) PostPagemanageLoad(c echo.Context) error {

	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		return c.String(http.StatusUnauthorized, "unauthorized, user not verified")
	}

	// get url from json body
	var postData rest.UrlWithHTML
	if err := c.Bind(&postData); err != nil {
		log.Debug().Msgf("failed to parse json body, %v\n", err)
		return c.String(http.StatusBadRequest, "failed to parse json body")
	}
	log.Debug().Msgf("postData %+v", postData)

	article, err := webscraping.GetArticleFromHtml(postData.HTML, postData.Url)
	log.Debug().Msgf("article %+v", article)
	if err != nil {
		log.Debug().Msgf("failed to parse %s, %v\n", postData.Url, err)
		return c.String(http.StatusBadRequest, "failed to parse url or html")
	}

	meili_ref, err := controller.PageController.SaveNewPage(
		record.Id, postData.Url, postData.Title, []string{}, article.TextContent, page.AvailableOriginExtention, false,
	)
	if err != nil {
		log.Debug().Msgf("failed to save page, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
	}

	return c.String(http.StatusOK, meili_ref)
}

func (controller WebController) DeletePagemanagePage(c echo.Context) error {
	// retrive user id from get params
	userRecord, _ := c.Get("authRecord").(*models.Record)
	if userRecord == nil || !userRecord.GetBool("verified") {
		return c.String(http.StatusUnauthorized, "unauthorized, user not verified")
	}

	var data rest.DeletePageRequest
	if err := c.Bind(&data); err != nil {
		log.Debug().Msgf("failed to parse json body, %v\n", err)
		return c.String(http.StatusBadRequest, "failed to parse json body")
	}

	err := controller.PageController.RemoveDocFTSIndex(data.PageID)
	if err != nil {
		log.Error().Err(err).Msg("error deleting page from index ")
		return c.String(http.StatusBadRequest, "error deleting page from index")
	}

	page, categories, _, err := controller.PageController.PageID2FullPageData(userRecord.Id, data.PageID)
	if err != nil {
		log.Debug().Msgf("failed to get page, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to get page ", err).Error())
	}
	for _, cat := range categories {
		err := controller.PageController.RemoveCategoryFromPageWithOwner(userRecord.Id, page.Id, cat.Name)
		if err != nil {
			log.Debug().Msgf("failed to delete category, %v\n", err)
			return c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
		}
	}

	err = controller.PageController.AppDao().Delete(page)
	if err != nil {
		log.Debug().Msgf("failed to delete page, %v\n", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to delete page ", err).Error())
	}

	return c.NoContent(http.StatusOK)
}
