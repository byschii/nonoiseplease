package jobs

func ScrapeBufferedPages(a int, b int) int {

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

	return a + b
}
