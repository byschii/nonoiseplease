package rest

import (
	categories "be/model/categories"
	fts_page_doc "be/model/fts_page_doc"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
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

// OnRecordAfterCreateRequest - users
func SaveNewUser(dao *daos.Dao, relatedUser string, nickname string) error {

	err := StoreNewUserDetails(dao, nickname, relatedUser)
	if err != nil {
		return err
	}

	return nil
}

// OnRecordAfterDeleteRequest - users
func DeleteIndexFromFTS(meiliClient *meilisearch.Client, indexName string) error {
	_, err := meiliClient.DeleteIndex(indexName)
	return err
}

// OnRecordBeforeDeleteRequest - pages
// this function doesn't explicitly delete the page
// it s supposed to be called OnRecordBeforeDeleteRequest
func BeforeRemovePage(pageId string, meiliClient *meilisearch.Client, dao *daos.Dao) error {

	// elimina da FTD
	err := DeleteDocFTSIndex(meiliClient, dao, pageId)
	if err != nil {
		log.Printf("failed to delete page from FTS, %v\n", err)
		return err
	}

	return nil

}

// OnRecordAfterDeleteRequest - pages
func AfterRemovePage(dao *daos.Dao) error {
	// get all categories
	categories, err := categories.GetAllCategories(dao)
	if err != nil {
		log.Printf("failed to get all categories, %v\n", err)
		return err
	}

	// delete orphan categories
	for _, category := range categories {
		err = DeleteOrphanCategory(&category, dao)
		if err != nil {
			log.Printf("failed to delete orphan category, %v\n", err)
			return err
		}
	}
	return nil
}
