package jobs

import (
	"be/pkg/page"
	"be/pkg/users"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/rs/zerolog/log"
)

func ScrapeBufferedPages(dao *daos.Dao) error {

	// get all users
	userList, err := users.List(dao)
	if err != nil {
		log.Error().Msgf("authenticating user")
		return err
	}
	for _, user := range userList {
		pages, err := page.ByUserId(dao, user.Id)
		if err != nil {
			log.Error().Msgf("failed to get pages for user %s error: %v", user.Id, err)
			continue
		}

		for _, page := range pages {

		}
	}

	/*

		// scrape url and get info
		article, withProxy, err := GetArticle(urlData.Url, false, controller.ConfigController)
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

	*/

	return nil
}
