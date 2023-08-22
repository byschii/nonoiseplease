package fts_page

import (
	"github.com/meilisearch/meilisearch-go"
)

var FTSDOCATTRIBUTES = []string{"id", "category", "content"}

// currently Full text Search is provided by MeiliSearch
type FTSPageDoc struct {
	// ID is the primary key of the document.
	// It is used to identify the document and to perform search queries.
	// It is also used to update or delete a document.
	// It is also used to retrieve a document.
	ID       string   `json:"id"`
	Category []string `json:"category"`
	Content  string   `json:"content"`
}

func (d *FTSPageDoc) Save(meiliClient *meilisearch.Client, indexName string) error {
	_, err := meiliClient.Index(indexName).AddDocuments(d, "id")
	return err
}

func (d *FTSPageDoc) UpdateCategories(meiliClient *meilisearch.Client, indexName string, categories []string) error {
	d.Category = categories

	// delete category from fts via update
	idx, err := meiliClient.GetIndex(indexName)
	if err != nil {
		return err
	}
	// update = add with same id
	idx.UpdateDocuments(d, "id")
	return err
}
