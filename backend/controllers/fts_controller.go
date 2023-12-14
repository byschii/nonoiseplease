package controllers

import (
	"log"

	cats "be/model/categories"
	tfs_page_doc "be/model/fts_page_doc"
	page "be/model/page"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
)

type FTSController struct {
	PBDao       *daos.Dao
	MeiliClient *meilisearch.Client
}

func (controller FTSController) DeleteDocFTSIndex(pageId string) error {

	log.Print("deleting " + pageId)
	// convert docId to ftsRef
	page, err := page.GetPageFromPageId(controller.PBDao, pageId)
	if err != nil {
		return err
	}
	_, err = controller.MeiliClient.Index(page.Owner).DeleteDocument(page.FTSRef)
	if err != nil {
		return err
	}

	return nil
}

func (controller FTSController) SetDBCategoriesForFTSDoc(owner string, FTSRef string, categories []cats.Category) error {
	// convert categories to string slice
	var categoryNames []string
	for _, category := range categories {
		categoryNames = append(categoryNames, category.Name)
	}
	err := tfs_page_doc.SetCategoriesForFTSDoc(controller.MeiliClient, owner, FTSRef, categoryNames)
	if err != nil {
		log.Printf("error while setting categories for doc %s: %s , cannot align db e fts", FTSRef, err.Error())
	}
	return err
}

func (controller FTSController) AlignCategoriesBetweenFTSAndDB(owner string, FTSRef string, pageId string) error {
	cateories, err := cats.GetCategoriesByPageId(controller.PBDao, pageId)
	if err != nil {
		log.Printf("error while getting categories for page %s: %s , cannot align db e fts", pageId, err.Error())
		return err
	}
	return controller.SetDBCategoriesForFTSDoc(owner, FTSRef, cateories)
}
