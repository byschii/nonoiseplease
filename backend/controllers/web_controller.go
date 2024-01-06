package controllers

import (
	u "be/utils"
	"be/webscraping"
	"log"
	"net/http"
	"strings"

	"be/model/page"
	rest "be/model/rest"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type WebController struct {
	PageController   PageController
	UserController   UserController
	ConfigController ConfigController
}

type NoNoiseController interface {
	AppDao() *daos.Dao
	AppMeiliClient() *meilisearch.Client
	GetSearch(c echo.Context) error
	GetSearchInfo(c echo.Context) error
	DeleteAccount(c echo.Context) error
}

func (controller WebController) AppDao() *daos.Dao {
	return controller.PageController.PBDao
}

func (controller WebController) AppMeiliClient() *meilisearch.Client {
	return controller.PageController.MeiliClient
}

func (controller WebController) DeleteAccount(c echo.Context) error {
	// retrive user id from get params
	record, _ := controller.UserController.UserRecordFromRequest(c, true)

	controller.UserController.DropAccount(record)
	return c.NoContent(http.StatusOK)
}

func (controller WebController) GetSearchInfo(c echo.Context) error {

	user, err := controller.UserController.UserFromRequest(c, false)
	if err != nil {
		log.Printf("failed to get user from request, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to get user from request ", err).Error())
		return nil
	}

	// get all categories from user pages
	categories, err := controller.PageController.CategoryController.FindCategoriesFromUser(user)
	if err != nil {
		log.Printf("failed to get categories, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to get categories ", err).Error())
		return nil
	}

	preSearchInfo := rest.PreSearchInfoResponse{
		Categories: categories,
	}
	c.JSON(http.StatusOK, preSearchInfo)

	return nil
}

func (controller WebController) GetSearch(c echo.Context) error {
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}
	userID := record.Id

	// read query and list of categories from url params
	query := c.QueryParam("query")
	categoriesParam := c.QueryParam("categories")

	// split categories over comma
	categories := strings.Split(categoriesParam, ",")

	pageResp, err := controller.PageController.PageSearch(query, []string{userID}, categories)
	if err != nil {
		log.Print("failed to search ", err)
		return c.String(http.StatusBadRequest, u.WrapError("failed to search ", err).Error())
	}

	resp := rest.SearchResponse{
		Pages: *pageResp,
	}
	return c.JSON(http.StatusOK, resp)
}

func (controller WebController) PostUrlScrape(c echo.Context) error {
	// retrive user id from get params
	userRecord, _ := c.Get("authRecord").(*models.Record)
	if userRecord == nil || !userRecord.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}

	pagesAlreadyScraped, err := controller.PageController.CountUserPagesByOriginThisMonth(userRecord.Id, page.AvailableOriginScrape)
	if err != nil {
		log.Printf("failed to count user pages, %v\n", err)
		c.String(http.StatusInternalServerError, "failed to count user pages")
		return nil
	}
	// if user has reached the limit, return error
	if pagesAlreadyScraped >= controller.ConfigController.maxScrapePerMonth {
		c.String(http.StatusForbidden, "you have reached the limit of pages you can scrape")
		return nil
	}

	// get url from json body
	var urlData rest.Url
	if err := c.Bind(&urlData); err != nil {
		log.Printf("failed to parse json body, %v\n", err)
		c.String(http.StatusBadRequest, "failed to parse json body")
		return nil
	}
	// scrape url and get info
	article, withProxy, err := webscraping.GetArticle(urlData.Url, false, controller.PageController.PBDao)
	if err != nil {
		log.Printf("failed to parse %s, %v\n", urlData.Url, err)
		c.String(http.StatusBadRequest, "failed to parse url")
		return nil
	}

	meili_ref, err := controller.PageController.SaveNewPage(
		userRecord.Id, urlData.Url, article.Title, []string{}, article.TextContent, page.AvailableOriginScrape, withProxy,
	)
	if err != nil {
		log.Printf("failed to save page, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
		return nil
	}

	return c.String(http.StatusOK, meili_ref)
}
