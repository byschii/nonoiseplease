package jobs

import (
	config "be/pkg/config"
	pagebuffer "be/pkg/page/buffer"
	pagecommons "be/pkg/page/commons"
	pagepage "be/pkg/page/page"
	pageservice "be/pkg/page/service"

	"be/pkg/scraping"
	"be/pkg/users"
	"math/rand"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/rs/zerolog/log"
)

func ScrapeBufferedPages(dao *daos.Dao, meiliClient *meilisearch.Client) error {

	// get all users
	userList, err := users.List(dao)
	if err != nil {
		log.Error().Msgf("authenticating user")
		return err
	}
	proxyProb := config.ProxyProb(dao)
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

		for i, bufferedPage := range bufferedPages {
			if i >= maxScraperPerMonth-scraped {
				break
			}
			useProxy := rand.Float32() < proxyProb
			article, withProxy, err := scraping.GetArticle(dao, bufferedPage.PageUrl, useProxy)
			if err != nil {
				log.Debug().Msgf("failed to parse %s, %v\n", bufferedPage.PageUrl, err)
				continue
			}

			// save page
			_, err = pageservice.SaveNewPage(
				dao,
				meiliClient,
				bufferedPage.Owner, bufferedPage.PageUrl, article.Title, []string{}, article.TextContent, pagecommons.AvailableOriginScrape,
				withProxy,
			)
			if err != nil {
				log.Debug().Msgf("failed to save page, %v\n", err)
				continue
			}
			// remove from buffer
			err = pagebuffer.Remove(dao, bufferedPage.Id)
		}
	}

	/*

		// scrape url and get info


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
