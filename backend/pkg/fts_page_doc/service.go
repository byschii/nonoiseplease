package fts_page

import (
	"encoding/json"

	"github.com/meilisearch/meilisearch-go"
)

func FromMeiliResultInterface(result interface{}) (FTSPageDoc, error) {
	var d FTSPageDoc
	jsonDoc, _ := json.Marshal(result)
	err := json.Unmarshal(jsonDoc, &d)
	return d, err
}

// returns a doc from meilisearch
// the doc is taken from the index 'owner', with the reference 'FTSRef'
func FromIndexAndRef(meiliClient *meilisearch.Client, indexName string, FTSRef string) (FTSPageDoc, error) {
	var parsedDoc FTSPageDoc
	doc, err := meiliClient.Index(indexName).Search(FTSRef, &meilisearch.SearchRequest{
		Limit:  2,
		Filter: "id = " + FTSRef,
	})
	if err != nil {
		return parsedDoc, err
	}

	if len(doc.Hits) == 0 {
		return parsedDoc, nil
	}
	return FromMeiliResultInterface(doc.Hits[0])
}
func SetCategoriesForFTSDoc(meiliClient *meilisearch.Client, owner string, FTSRef string, categories []string) error {

	ftsDoc, err := FromIndexAndRef(meiliClient, owner, FTSRef)
	if err != nil {
		return err
	}

	ftsDoc.UpdateCategories(meiliClient, owner, categories)
	return nil
}
