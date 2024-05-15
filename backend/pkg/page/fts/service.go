package page

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

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

	updateCategories(&ftsDoc, meiliClient, owner, categories)
	return nil
}

func updateCategories(d *FTSPageDoc, meiliClient *meilisearch.Client, indexName string, categories []string) error {
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

func NewFtsDocRef(owner string,
	url string,
	pageTitle string) string {
	// create doc id with sha
	hash := sha256.New()
	hash.Write([]byte(url + pageTitle + owner + fmt.Sprint(rand.Intn(1000000)) + time.Now().String()))
	reference := hex.EncodeToString(hash.Sum(nil))
	return reference
}

func Save(d *FTSPageDoc, meiliClient *meilisearch.Client, indexName string) error {
	_, err := meiliClient.Index(indexName).AddDocuments(d, "id")
	return err
}
