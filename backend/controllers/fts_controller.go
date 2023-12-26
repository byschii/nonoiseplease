package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	cats "be/model/categories"
	fts_page_doc "be/model/fts_page_doc"
	page "be/model/page"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
)

type FTSController struct {
	PBDao       *daos.Dao
	MeiliClient *meilisearch.Client
}

type FTSControllerInterface interface {
	CommonController
	RemoveDocFTSIndex(pageId string) error
	SetDBCategoriesOnFTSDoc(owner string, FTSRef string, categories []cats.Category) error
	AlignCategoriesBetweenFTSAndDB(owner string, FTSRef string, pageId string) error
	CreateNewFTSIndex(indexName string, waitTimeRange float32) error
}

func (controller FTSController) GetDao() *daos.Dao {
	return controller.PBDao
}

func (controller FTSController) RemoveDocFTSIndex(pageId string) error {

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

func (controller FTSController) SetDBCategoriesOnFTSDoc(owner string, FTSRef string, categories []cats.Category) error {
	// convert categories to string slice
	var categoryNames []string
	for _, category := range categories {
		categoryNames = append(categoryNames, category.Name)
	}
	err := fts_page_doc.SetCategoriesForFTSDoc(controller.MeiliClient, owner, FTSRef, categoryNames)
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
	return controller.SetDBCategoriesOnFTSDoc(owner, FTSRef, cateories)
}

func (controller FTSController) CreateNewFTSIndex(indexName string, waitTimeRange float32) error {
	// create index for his searches
	taskInfo, err := controller.MeiliClient.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexName,
		PrimaryKey: "id",
	})
	if err != nil {
		return err
	}

	// wait til creation
	creationSuccess := false
	for !creationSuccess {
		taskData, err := controller.MeiliClient.GetTask(taskInfo.TaskUID)
		fmt.Print(".")
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(rand.Float32()*waitTimeRange) * time.Second)
		creationSuccess = taskData.Status == "succeeded"
	}

	// make it searchable
	_, err = controller.MeiliClient.Index(indexName).UpdateFilterableAttributes(&fts_page_doc.FTSDOCATTRIBUTES)
	return err
}
