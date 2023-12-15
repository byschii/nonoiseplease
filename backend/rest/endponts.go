package rest

import (
	"be/controllers"
	page "be/model/page"
	"be/model/rest"
	u "be/utils"
	webscraping "be/webscraping"

	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/models"
)

func GerVersion(version string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, version)
	}
}

func PostUrlScrape(pageController controllers.PageController, maxScrapePerMonth int) echo.HandlerFunc {

	return func(c echo.Context) error {
		// retrive user id from get params
		userRecord, _ := c.Get("authRecord").(*models.Record)
		if userRecord == nil || !userRecord.GetBool("verified") {
			c.String(http.StatusUnauthorized, "unauthorized, user not verified")
			return nil
		}

		pagesAlreadyScraped, err := pageController.CountUserPagesByOriginThisMonth(userRecord.Id, page.AvailableOriginScrape)
		if err != nil {
			log.Printf("failed to count user pages, %v\n", err)
			c.String(http.StatusInternalServerError, "failed to count user pages")
			return nil
		}

		log.Println("pagesAlreadyScraped", pagesAlreadyScraped)

		// if user has reached the limit, return error
		if pagesAlreadyScraped >= maxScrapePerMonth {
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
		article, withProxy, err := webscraping.GetArticle(urlData.Url, false, pageController.PBDao)
		if err != nil {
			log.Printf("failed to parse %s, %v\n", urlData.Url, err)
			c.String(http.StatusBadRequest, "failed to parse url")
			return nil
		}

		meili_ref, err := pageController.SaveNewPage(
			userRecord.Id, urlData.Url, article.Title, []string{}, article.TextContent, page.AvailableOriginScrape, withProxy,
		)
		if err != nil {
			log.Printf("failed to save page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
			return nil
		}

		return c.String(http.StatusOK, meili_ref)
	}
}

func PostPagemanageLoad(pageController controllers.PageController, authController controllers.AuthController, tokenSecret string) echo.HandlerFunc {
	return func(c echo.Context) error {

		// get url from json body
		var postData rest.UrlWithHTML
		if err := c.Bind(&postData); err != nil {
			log.Printf("failed to parse json body, %v\n", err)
			c.String(http.StatusBadRequest, "failed to parse json body")
			return nil
		}

		userRecord, err := authController.GetUserFromJWT(postData.AuthCode)
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

		meili_ref, err := pageController.SaveNewPage(
			userRecord.Id, postData.Url, postData.Title, []string{}, article.TextContent, page.AvailableOriginExtention, false,
		)
		if err != nil {
			log.Printf("failed to save page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to save page ", err).Error())
			return nil
		}

		return c.String(http.StatusOK, meili_ref)
	}
}

func GetPagemanage(pageController controllers.PageController) echo.HandlerFunc {
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

		page, categories, ref, err := pageController.GetFullPageDataByID(userID, pageID)
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
			pageController.FTSController.SetDBCategoriesForFTSDoc(userID, page.FTSRef, categories)
		}

		result := rest.PageResponse{
			Page:       *page,
			Categories: categories,
			FTSDoc:     *ref,
		}

		return c.JSON(http.StatusOK, result)
	}
}

func PostPagemanageCategory(pageController controllers.PageController) echo.HandlerFunc {
	return func(c echo.Context) error {
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
		err := pageController.CategoryController.AddCategoryToPage(*pageController.FTSController, userRecord.Id, data.PageID, data.CategoryName)

		if err != nil {
			log.Printf("failed to add category, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to add category ", err).Error())
			return nil
		}

		return c.NoContent(http.StatusOK)
	}
}

func DeletePagemanagePage(pageController controllers.PageController) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		err := pageController.FTSController.DeleteDocFTSIndex(data.PageID)
		if err != nil {
			log.Print("error deleting page from index ", err)
			c.String(http.StatusBadRequest, "error deleting page from index")
			return nil
		}

		page, categories, _, err := pageController.GetFullPageDataByID(userRecord.Id, data.PageID)
		if err != nil {
			log.Printf("failed to get page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to get page ", err).Error())
			return nil
		}
		for _, cat := range categories {
			err := pageController.DeleteCategoryFromPageWithOwner(userRecord.Id, page.Id, cat.Name)
			if err != nil {
				log.Printf("failed to delete category, %v\n", err)
				c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
				return nil
			}
		}

		err = pageController.PBDao.Delete(page)
		if err != nil {
			log.Printf("failed to delete page, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to delete page ", err).Error())
			return nil
		}

		return c.NoContent(http.StatusOK)
	}

}

func DeletePagemanageCategory(pageController controllers.PageController) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		err := pageController.DeleteCategoryFromPageWithOwner(userID, data.PageID, data.CategoryName)
		if err != nil {
			log.Printf("failed to delete category, %v\n", err)
			c.String(http.StatusBadRequest, u.WrapError("failed to delete category ", err).Error())
			return nil
		}

		return c.NoContent(http.StatusOK)
	}
}
