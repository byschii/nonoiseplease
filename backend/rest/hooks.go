package rest

import (
	"be/controllers"
	categories "be/model/categories"
	fts_page_doc "be/model/fts_page_doc"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

// OnRecordAfterCreateRequest - users
func CreateNewFTSIndex(meiliClient *meilisearch.Client, indexName string, waitTimeRange float32) error {
	// create index for his searches
	taskInfo, err := meiliClient.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexName,
		PrimaryKey: "id",
	})
	if err != nil {
		return err
	}

	// wait til creation
	creationSuccess := false
	for !creationSuccess {
		taskData, err := meiliClient.GetTask(taskInfo.TaskUID)
		fmt.Print(".")
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(rand.Float32()*waitTimeRange) * time.Second)
		creationSuccess = taskData.Status == "succeeded"
	}

	// make it searchable
	_, err = meiliClient.Index(indexName).UpdateFilterableAttributes(&fts_page_doc.FTSDOCATTRIBUTES)
	return err
}

// OnRecordAfterDeleteRequest - users
func DeleteIndexFromFTS(meiliClient *meilisearch.Client, indexName string) error {
	_, err := meiliClient.DeleteIndex(indexName)
	return err
}

// OnRecordBeforeDeleteRequest - pages
// this function doesn't explicitly delete the page
// it s supposed to be called OnRecordBeforeDeleteRequest
func BeforeRemovePage(pageId string, fulltextsearchController controllers.FTSController) error {

	// elimina da FTD
	err := fulltextsearchController.DeleteDocFTSIndex(pageId)
	if err != nil {
		log.Printf("failed to delete page from FTS, %v\n", err)
		return err
	}

	return nil

}

// OnRecordAfterDeleteRequest - pages
func AfterRemovePage(categoryController controllers.CategoryController) error {
	// get all categories
	categories, err := categories.GetAllCategories(categoryController.PBDao)
	if err != nil {
		log.Printf("failed to get all categories, %v\n", err)
		return err
	}

	// delete orphan categories
	for _, category := range categories {
		err = categoryController.DeleteOrphanCategory(&category)
		if err != nil {
			log.Printf("failed to delete orphan category, %v\n", err)
			return err
		}
	}
	return nil
}
