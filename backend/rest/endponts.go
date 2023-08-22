package rest

import (
	page "be/model/page"
	u "be/utils"
	webscraping "be/webscraping"
	"strings"

	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"

	users "be/model/users"
)

func GerVersion(version string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, version)
	}
}

func PostUrlScrape(meiliClient *meilisearch.Client, dao *daos.Dao, maxScrapePerMonth int) echo.HandlerFunc {
	return func(c echo.Context) error {
		// retrive user id from get params
		userRecord, _ := c.Get("authRecord").(*models.Record)
		if userRecord == nil || !userRecord.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}

		pagesAlreadyScraped, err := CountUserPagesByOriginThisMonth(userRecord.Id, dao, page.AvailableOriginScrape)
		if err != nil {
			log.Printf("failed to count user pages, %v\n", err)
			c.String(http.StatusInternalServerError, "failed to count user pages")
			return nil
		}

		// if user has reached the limit, return error
		if pagesAlreadyScraped >= maxScrapePerMonth {
			c.String(http.StatusForbidden, "you have reached the limit of pages you can scrape")
			return nil
		}

		// get url from json body
		var urlData Url
		if err := c.Bind(&urlData); err != nil {
			log.Printf("failed to parse json body, %v\n", err)
			c.String(http.StatusBadRequest, "failed to parse json body")
			return nil
		}
		// scrape url and get info
		article, withProxy, err := webscraping.GetArticle(urlData.Url, false, dao)
		if err != nil {
			log.Printf("failed to parse %s, %v\n", urlData.Url, err)
			c.String(http.StatusBadRequest, "failed to parse url")
			return nil
		}

		meili_ref, err := SaveNewPage(
			userRecord.Id, urlData.Url, article.Title, []string{}, article.TextContent, page.AvailableOriginScrape, withProxy, meiliClient, dao,
		)
		if err != nil {
			log.Printf("failed to save page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
			return nil
		}

		return c.String(http.StatusOK, meili_ref)
	}
}

func PostPagemanageLoad(meiliClient *meilisearch.Client, dao *daos.Dao, tokenSecret string) echo.HandlerFunc {
	return func(c echo.Context) error {

		// get url from json body
		var postData UrlWithHTML
		if err := c.Bind(&postData); err != nil {
			log.Printf("failed to parse json body, %v\n", err)
			c.String(http.StatusBadRequest, "failed to parse json body")
			return nil
		}

		userRecord, err := users.GetUserFromJWT(postData.AuthCode, dao, tokenSecret)
		if err != nil {
			log.Printf("failed to get user from request, %v\n", err)
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}

		// scrape url and get info
		article, err := webscraping.GetArticleFromHtml(postData.HTML, postData.Url)
		if err != nil {
			log.Printf("failed to parse %s, %v\n", postData.Url, err)
			c.String(http.StatusBadRequest, "failed to parse url or html")
			return nil
		}

		meili_ref, err := SaveNewPage(
			userRecord.Id, postData.Url, postData.Title, []string{}, article.TextContent, page.AvailableOriginExtention, false, meiliClient, dao,
		)
		if err != nil {
			log.Printf("failed to save page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
			return nil
		}

		return c.String(http.StatusOK, meili_ref)
	}
}

func GetPagemanage(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {
		// retrive user id from get params
		record, _ := c.Get("authRecord").(*models.Record)
		if record == nil || !record.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}
		userID := record.Id

		// read page id from params
		pageID := c.QueryParam("id")

		page, categories, ref, err := GetFullPageDataByID(userID, pageID, meiliClient, dao)
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
			SetDBCategoriesForFTSDoc(meiliClient, userID, page.FTSRef, categories)
		}

		result := PageResponse{
			Page:       *page,
			Categories: categories,
			FTSDoc:     *ref,
		}

		return c.JSON(http.StatusOK, result)
	}
}

func PostPagemanageCategory(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {
		// retrive user id from get params
		userRecord, _ := c.Get("authRecord").(*models.Record)
		if userRecord == nil || !userRecord.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}

		// read page and category id from body
		var data PostCategoryRequest
		if err := c.Bind(&data); err != nil {
			log.Printf("failed to parse json body, %v\n", err)
			c.String(http.StatusBadRequest, "failed to parse json body")
			return nil
		}

		log.Println("data", data)
		err := AddCategoryToPage(userRecord.Id, data.PageID, data.CategoryName, meiliClient, dao)

		if err != nil {
			log.Printf("failed to add category, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to add category ", err).Error())
			return nil
		}

		return c.NoContent(http.StatusOK)
	}
}

func DeletePagemanagePage(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {
		// retrive user id from get params
		userRecord, _ := c.Get("authRecord").(*models.Record)
		if userRecord == nil || !userRecord.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}

		var data DeletePageRequest
		if err := c.Bind(&data); err != nil {
			log.Printf("failed to parse json body, %v\n", err)
			c.String(http.StatusBadRequest, "failed to parse json body")
			return nil
		}

		err := DeleteDocFTSIndex(meiliClient, dao, data.PageID)
		if err != nil {
			log.Print("error deleting page from index ", err)
			c.String(http.StatusBadRequest, "error deleting page from index")
			return nil
		}

		page, categories, _, err := GetFullPageDataByID(userRecord.Id, data.PageID, meiliClient, dao)
		if err != nil {
			log.Printf("failed to get page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to get page ", err).Error())
			return nil
		}
		for _, cat := range categories {
			err := DeleteCategoryFromPageWithOwner(userRecord.Id, page.Id, cat.Name, meiliClient, dao)
			if err != nil {
				log.Printf("failed to delete category, %v\n", err)
				c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
				return nil
			}
		}

		err = dao.Delete(page)
		if err != nil {
			log.Printf("failed to delete page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to delete page ", err).Error())
			return nil
		}

		return c.NoContent(http.StatusOK)
	}

}

func DeletePagemanageCategory(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {
		// retrive user id from get params
		record, _ := c.Get("authRecord").(*models.Record)
		if record == nil || !record.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}
		userID := record.Id

		// read page and category id from body
		var data DeleteCategoryRequest
		if err := c.Bind(&data); err != nil {
			log.Printf("failed to parse json body, %v\n", err)
			c.String(http.StatusBadRequest, "failed to parse json body")
			return nil
		}

		err := DeleteCategoryFromPageWithOwner(userID, data.PageID, data.CategoryName, meiliClient, dao)
		if err != nil {
			log.Printf("failed to delete category, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
			return nil
		}

		return c.NoContent(http.StatusOK)
	}
}

// delete user in auth table
// shoud trigger
//   - delete user important data in db
//   - delete user details in db
//
// then
//   - delete user meili index
func DeleteDropaccount(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {
		// retrive user id from get params
		record, _ := c.Get("authRecord").(*models.Record)
		if record == nil || !record.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}
		userID := record.Id

		go func() {
			meiliClient.Index(userID).DeleteAllDocuments()
			meiliClient.DeleteIndex(userID)
		}()

		go func() {
			// delete user data
			details, _ := GetUserPartFromId(dao, userID)
			dao.Delete(details)
			dao.Delete(record)
		}()

		return c.NoContent(http.StatusOK)
	}
}

func GetSearchInfo(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {

		record, _ := c.Get("authRecord").(*models.Record)
		if record == nil || !record.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}
		userID := record.Id

		// get all categories from user pages
		categories, err := GetCategoriesByUserId(dao, userID)
		if err != nil {
			log.Printf("failed to get categories, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to get categories ", err).Error())
			return nil
		}

		preSearchInfo := PreSearchInfoResponse{
			Categories: categories,
		}
		c.JSON(http.StatusOK, preSearchInfo)

		return nil
	}
}

func GetSearch(meiliClient *meilisearch.Client, dao *daos.Dao) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		pageResp, err := PageSearch(query, []string{userID}, categories, meiliClient, dao)
		if err != nil {
			log.Print("failed to search ", err)
			return c.String(http.StatusBadRequest, u.WrapError("failed to search ", err).Error())
		}

		resp := SearchResponse{
			Pages: *pageResp,
		}
		return c.JSON(http.StatusOK, resp)

	}
}
