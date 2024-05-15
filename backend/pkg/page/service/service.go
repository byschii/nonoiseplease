package services

import (
	pagecommons "be/pkg/page/commons"
	pagefts "be/pkg/page/fts"
	page "be/pkg/page/page"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
)

func SaveNewPage(
	dao *daos.Dao,
	meiliClient *meilisearch.Client,
	owner string,
	url string,
	pageTitle string,
	categories []string,
	content string,
	originType pagecommons.AvailableOrigin,
	withProxy bool) (string, error) {

	reference := pagefts.NewFtsDocRef(owner, url, pageTitle)

	// checking for errors
	errs := make(chan error, 2)

	go func() {
		savedArticle := page.Page{
			Link:      url,
			PageTitle: pageTitle,
			Owner:     owner,
			FTSRef:    reference,
			Votes:     0,
			WithProxy: withProxy,
			Origin:    originType,
		}
		errs <- dao.Save(&savedArticle)
	}()

	go func() {
		doc := pagefts.FTSPageDoc{
			ID:       reference,
			Category: []string{},
			Content:  content,
		}
		errs <- pagefts.Save(&doc, meiliClient, owner)
	}()

	// maybe errored
	for i := 0; i < 2; i++ {
		err := <-errs
		if err != nil {
			return "", err
		}
	}

	return reference, nil

}
