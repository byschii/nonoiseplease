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
	"github.com/pocketbase/pocketbase/models"
)

type WebController struct {
	PageController   PageControllerInterface
	UserController   UserControllerInterface
	ConfigController ConfigControllerInterface
}

type NoNoiseInterface interface {
	GetSearch(c echo.Context) error
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

func (controller WebController) GetSearchInfo(c echo.Context) error {

	user, err := controller.UserController.UserFromRequest(c, false)
	if err != nil {
		log.Printf("failed to get user from request, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to get user from request ", err).Error())
		return nil
	}

	// get all categories from user pages
	categories, err := controller.PageController.FindCategoriesFromUser(user)
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
		log.Println("failed to search ", err)
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
	if pagesAlreadyScraped >= controller.ConfigController.MaxScrapePerMonth() {
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
	article, withProxy, err := webscraping.GetArticle(urlData.Url, false, controller.PageController.AppDao())
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

func (controller WebController) PostPagemanageCategory(c echo.Context) error {
	// retrive user id from get params
	userRecord, _ := c.Get("authRecord").(*models.Record)
	if userRecord == nil || !userRecord.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}

	// read page and category id from body
	var data rest.PostCategoryRequest
	if err := c.Bind(&data); err != nil {
		log.Printf("failed to parse json body, %v\n", err)
		c.String(http.StatusBadRequest, "failed to parse json body")
		return nil
	}

	log.Println("data", data)
	err := controller.PageController.AddCategoryToPage(userRecord.Id, data.PageID, data.CategoryName)

	if err != nil {
		log.Printf("failed to add category, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to add category ", err).Error())
		return nil
	}

	return c.NoContent(http.StatusOK)
}

func (controller WebController) GetPagemanage(c echo.Context) error {
	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}
	userID := record.Id

	// read page id from params
	pageID := c.QueryParam("id")

	page, categories, ref, err := controller.PageController.PageID2FullPageData(userID, pageID)
	if err != nil || ref == nil {
		log.Printf("failed to get page, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to get page ", err).Error())
		return nil
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
		/*log.Println("categories are not the same")
		log.Println("db categories", categories, len(categories))
		log.Println("meili categories", ref.Category, len(ref.Category))
		c.String(http.StatusBadRequest, "categories are not the same")
		return nil*/
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
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}
	userID := record.Id

	// read page and category id from body
	var data rest.DeleteCategoryRequest
	if err := c.Bind(&data); err != nil {
		log.Printf("failed to parse json body, %v\n", err)
		c.String(http.StatusBadRequest, "failed to parse json body")
		return nil
	}

	err := controller.PageController.RemoveCategoryFromPageWithOwner(userID, data.PageID, data.CategoryName)
	if err != nil {
		log.Printf("failed to delete category, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
		return nil
	}

	return c.NoContent(http.StatusOK)
}

func (controller WebController) PostPagemanageLoad(c echo.Context) error {
	// get url from json body
	var postData rest.UrlWithHTML
	if err := c.Bind(&postData); err != nil {
		log.Printf("failed to parse json body, %v\n", err)
		c.String(http.StatusBadRequest, "failed to parse json body")
		return nil
	}

	userRecord, err := controller.UserController.AuthorizationController().FindUserForExtention(postData.UserId, postData.ExtentionToken, postData.AuthCode)
	if err != nil {
		log.Printf("failed to get user from request, %v\n", err)
		c.String(http.StatusUnauthorized, "unauthorized, user not found")
		return nil
	}

	article, err := webscraping.GetArticleFromHtml(postData.HTML, postData.Url)
	if err != nil {
		log.Printf("failed to parse %s, %v\n", postData.Url, err)
		c.String(http.StatusBadRequest, "failed to parse url or html")
		return nil
	}

	meili_ref, err := controller.PageController.SaveNewPage(
		userRecord.Id, postData.Url, postData.Title, []string{}, article.TextContent, page.AvailableOriginExtention, false,
	)
	if err != nil {
		log.Printf("failed to save page, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
		return nil
	}

	return c.String(http.StatusOK, meili_ref)
}

func (controller WebController) DeletePagemanagePage(c echo.Context) error {
	// retrive user id from get params
	userRecord, _ := c.Get("authRecord").(*models.Record)
	if userRecord == nil || !userRecord.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}

	var data rest.DeletePageRequest
	if err := c.Bind(&data); err != nil {
		log.Printf("failed to parse json body, %v\n", err)
		c.String(http.StatusBadRequest, "failed to parse json body")
		return nil
	}

	err := controller.PageController.RemoveDocFTSIndex(data.PageID)
	if err != nil {
		log.Println("error deleting page from index ", err)
		c.String(http.StatusBadRequest, "error deleting page from index")
		return nil
	}

	page, categories, _, err := controller.PageController.PageID2FullPageData(userRecord.Id, data.PageID)
	if err != nil {
		log.Printf("failed to get page, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to get page ", err).Error())
		return nil
	}
	for _, cat := range categories {
		err := controller.PageController.RemoveCategoryFromPageWithOwner(userRecord.Id, page.Id, cat.Name)
		if err != nil {
			log.Printf("failed to delete category, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
			return nil
		}
	}

	err = controller.PageController.AppDao().Delete(page)
	if err != nil {
		log.Printf("failed to delete page, %v\n", err)
		c.String(http.StatusBadRequest, u.WrapError("failed to delete page ", err).Error())
		return nil
	}

	return c.NoContent(http.StatusOK)
}
