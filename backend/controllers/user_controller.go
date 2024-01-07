package controllers

import (
	users "be/model/users"
	"fmt"
	"log"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// controls users (not users pages)
type SimpleUserController struct {
	App            *pocketbase.PocketBase
	MeiliClient    *meilisearch.Client
	AuthController AuthControllerInterface
}

type UserControllerInterface interface {
	CommonController
	AuthorizationController() AuthControllerInterface
	AppDao() *daos.Dao
	SetApp(app *pocketbase.PocketBase)
	GetUserDetails(relatedUserId string) (*users.UserDetails, error)
	DropAccount(record *models.Record)
	SaveActivity(activity users.UserActivity) error
	StoreNewUserDetails(nickname string, relatedUserId string) error
	GetUserEmailFromId(userId string) (string, error)
	UserFromRequest(c echo.Context, mustBeVerified bool) (*users.Users, error)
	UserRecordFromRequest(c echo.Context, mustBeVerified bool) (*models.Record, error)
}

func NewUserController(pbApp *pocketbase.PocketBase, meilisearchClient *meilisearch.Client, authController AuthControllerInterface) UserControllerInterface {
	return &SimpleUserController{
		App:            pbApp,
		MeiliClient:    meilisearchClient,
		AuthController: authController,
	}
}

func (controller SimpleUserController) AuthorizationController() AuthControllerInterface {
	return controller.AuthController
}

func (controller SimpleUserController) AppDao() *daos.Dao {
	return controller.App.Dao()
}

func (controller *SimpleUserController) SetApp(app *pocketbase.PocketBase) {
	controller.App = app
}

func (controller SimpleUserController) UserFromRequest(c echo.Context, mustBeVerified bool) (*users.Users, error) {

	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || (mustBeVerified && !record.GetBool("verified")) {
		return nil, fmt.Errorf("unauthorized, user not verified")
	}

	user, err := controller.AuthController.FindUserById(record.Id)
	return user, err
}

func (controller SimpleUserController) UserRecordFromRequest(c echo.Context, mustBeVerified bool) (*models.Record, error) {

	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || (mustBeVerified && !record.GetBool("verified")) {
		return nil, fmt.Errorf("unauthorized, user not verified")
	}

	return record, nil
}

func (controller SimpleUserController) GetUserDetails(relatedUserId string) (*users.UserDetails, error) {
	details := &users.UserDetails{}
	log.Println("controller.AppDao()", controller.AppDao() == nil)
	err := controller.AppDao().DB().Select("*").From(details.TableName()).Where(dbx.In("related_user", relatedUserId)).One(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

// delete user in auth table
// shoud trigger
//   - delete user important data in db
//   - delete user details in db
//
// then
//   - delete user meili index
func (controller SimpleUserController) DropAccount(u *models.Record) {
	go func() {
		controller.MeiliClient.Index(u.GetId()).DeleteAllDocuments()
		controller.MeiliClient.DeleteIndex(u.GetId())
	}()

	go func() {
		// delete user data
		details, _ := controller.GetUserDetails(u.GetId())
		controller.AppDao().Delete(details)
		controller.AppDao().Delete(u)
	}()

}

func (controller SimpleUserController) SaveActivity(activity users.UserActivity) error {
	err := controller.AppDao().Save(&activity)
	return err
}

func (controller SimpleUserController) StoreNewUserDetails(relatedUserId string, nickname string) error {
	details := &users.UserDetails{
		Nickname:    nickname,
		RelatedUser: relatedUserId,
	}
	err := controller.AppDao().Save(details)
	return err
}

func (controller SimpleUserController) GetUserEmailFromId(userId string) (string, error) {
	userToFill := &users.Users{}
	err := controller.AppDao().FindById(userToFill, userId)
	if err != nil {
		return "", err
	}
	return userToFill.Email, nil
}
