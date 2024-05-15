package jobs

import (
	"be/pkg/config"
	pagebuffer "be/pkg/page/buffer"
	pagepage "be/pkg/page/page"
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
		bufferedPages, err := pagebuffer.BufferedByUserId(dao, user.Id)
		if err != nil {
			log.Error().Msgf("failed to get buffered pages for user %s error: %v", user.Id, err)
			continue
		}

		// cout already scraped pages
		scraped, err := pagepage.CountUserPagesScrapedThisMonth(dao, user.Id)
		if err != nil {
			log.Error().Msgf("failed to count scraped pages for user %s error: %v", user.Id, err)
			continue
		}
		// get max scrape per month
		maxScraperPerMonth := config.CountMaxScrapePerMonth(dao)
		// log ser info on auto scraping
		log.Debug().Msgf(
			"user %s scraped %d pages this month, %d pages yet to be scraped from buffer, but up to %d pages per month, goint to scrape %d pages",
			user.Id,
			scraped,
			len(bufferedPages),
			maxScraperPerMonth,
			maxScraperPerMonth-scraped,
		)

		for i, _ := range bufferedPages {
			if i >= maxScraperPerMonth-scraped {
				break
			}
			// scrape url and get info

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
