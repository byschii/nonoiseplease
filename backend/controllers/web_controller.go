package controllers

import (
	u "be/utils"
	"log"
	"net/http"
	"strings"

	rest "be/model/rest"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type WebController struct {
	PageController PageController
}

type NoNoiseController interface {
	AppDao() *daos.Dao
	AppMeiliClient() *meilisearch.Client
	GetSearch(c echo.Context) error
	GetSearchInfo(c echo.Context) error
}

func (controller WebController) AppDao() *daos.Dao {
	return controller.PageController.PBDao
}

func (controller WebController) AppMeiliClient() *meilisearch.Client {
	return controller.PageController.MeiliClient
}

func (controller WebController) GetSearchInfo(c echo.Context) error {

	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}
	userID := record.Id

	// get all categories from user pages
	categories, err := controller.PageController.CategoryController.GetCategoriesByUserId(userID)
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
