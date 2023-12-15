package controllers

import (
	users "be/model/users"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type UserController interface {
	GetUserDetails(relatedUserId string) (*users.UserDetails, error)
	DeleteDropaccount(c echo.Context) error
	SaveActivity(activity users.UserActivity) error
	StoreNewUserDetails(nickname string, relatedUserId string) error
	GetUserEmailFromId(userId string) (string, error)
	GetDao() *daos.Dao
}

// controls users (not users pages)
type SimpleUserController struct {
	PBDao       *daos.Dao
	MeiliClient *meilisearch.Client
}

func (controller SimpleUserController) GetDao() *daos.Dao {
	return controller.PBDao
}

func (controller SimpleUserController) GetUserDetails(relatedUserId string) (*users.UserDetails, error) {
	details := &users.UserDetails{}
	err := controller.PBDao.DB().Select("*").From(details.TableName()).Where(dbx.In("related_user", relatedUserId)).One(details)
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
func (controller SimpleUserController) DeleteDropaccount(c echo.Context) error {
	// retrive user id from get params
	record, _ := c.Get("authRecord").(*models.Record)
	if record == nil || !record.GetBool("verified") {
		c.String(http.StatusUnauthorized, "unauthorized, user not verified")
		return nil
	}
	userID := record.Id

	go func() {
		controller.MeiliClient.Index(userID).DeleteAllDocuments()
		controller.MeiliClient.DeleteIndex(userID)
	}()

	go func() {
		// delete user data
		details, _ := controller.GetUserDetails(userID)
		controller.PBDao.Delete(details)
		controller.PBDao.Delete(record)
	}()

	return c.NoContent(http.StatusOK)
}

func (controller SimpleUserController) SaveActivity(activity users.UserActivity) error {
	err := controller.PBDao.Save(&activity)
	return err
}

func (controller SimpleUserController) StoreNewUserDetails(relatedUserId string, nickname string) error {
	details := &users.UserDetails{
		Nickname:    nickname,
		RelatedUser: relatedUserId,
	}
	err := controller.PBDao.Save(details)
	return err
}

func (controller SimpleUserController) GetUserEmailFromId(userId string) (string, error) {
	userToFill := &users.Users{}
	err := controller.PBDao.FindById(userToFill, userId)
	if err != nil {
		return "", err
	}
	return userToFill.Email, nil
}
