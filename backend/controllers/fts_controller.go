package controllers

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"

	cats "be/pkg/categories"
	pagefts "be/pkg/page/fts"
	page "be/pkg/page/page"

	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase/daos"
)

type FTSController struct {
	PBDao       *daos.Dao
	MeiliClient *meilisearch.Client
}

type FTSControllerInterface interface {
	CommonController
	SetPBDAO(dao *daos.Dao)
	RemoveDocFTSIndex(pageId string) error
	SetDBCategoriesOnFTSDoc(owner string, FTSRef string, categories []cats.Category) error
	AlignCategoriesBetweenFTSAndDB(owner string, FTSRef string, pageId string) error
	CreateNewFTSIndex(indexName string, waitTimeRange float32) error
}

func NewFTSController(dao *daos.Dao, meiliClient *meilisearch.Client) FTSControllerInterface {
	return &FTSController{
		PBDao:       dao,
		MeiliClient: meiliClient,
	}
}

func (controller FTSController) AppDao() *daos.Dao {
	return controller.PBDao
}

func (controller *FTSController) SetPBDAO(dao *daos.Dao) {
	controller.PBDao = dao
}

func (controller FTSController) RemoveDocFTSIndex(pageId string) error {

	log.Debug().Msgf("deleting " + pageId)
	// convert docId to ftsRef
	page, err := page.FromId(controller.PBDao, pageId)
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
	err := pagefts.SetCategoriesForFTSDoc(controller.MeiliClient, owner, FTSRef, categoryNames)
	if err != nil {
		log.Debug().Msgf("error while setting categories for doc %s: %s , cannot align db e fts", FTSRef, err.Error())
	}
	return err
}

func (controller FTSController) AlignCategoriesBetweenFTSAndDB(owner string, FTSRef string, pageId string) error {
	cateories, err := cats.GetCategoriesByPageId(controller.PBDao, pageId)
	if err != nil {
		log.Debug().Msgf("error while getting categories for page %s: %s , cannot align db e fts", pageId, err.Error())
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
	_, err = controller.MeiliClient.Index(indexName).UpdateFilterableAttributes(&pagefts.FTSDOCATTRIBUTES)
	return err
}
