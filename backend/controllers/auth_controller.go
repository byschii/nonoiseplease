package controllers

import (
	users "be/pkg/users"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type AuthController struct {
	App *pocketbase.PocketBase
}

type AuthControllerInterface interface {
	CommonController
	TokenSecret() string
	SetApp(app *pocketbase.PocketBase)
	FindUserFromExtentionToken(userId string, extentionToken string) (*users.UserDetails, error)
	FindUserFromJWT(jwt string) (*models.Record, error)
	FindUserFromJWTInContext(c echo.Context) (*models.Record, error)
	FindUserById(id string) (*users.User, error)
	CheckAuthCredentials(email string, password string, endpoint string) error
}

func NewAuthController(pbApp *pocketbase.PocketBase) AuthControllerInterface {
	return &AuthController{
		App: pbApp,
	}
}

func (controller *AuthController) SetApp(app *pocketbase.PocketBase) {
	controller.App = app
}

func (controller AuthController) AppDao() *daos.Dao {
	return controller.App.Dao()
}

func (controller *AuthController) TokenSecret() string {
	return controller.App.Settings().RecordAuthToken.Secret
}

func (authController AuthController) FindUserFromExtentionToken(userId string, extentionToken string) (*users.UserDetails, error) {
	u := &users.UserDetails{}
	q := authController.AppDao().ModelQuery(&users.UserDetails{})

	err := q.AndWhere(dbx.HashExp{"related_user": userId}).
		AndWhere(dbx.HashExp{"extension_token": extentionToken}).
		Limit(1).
		One(u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// get user from token
func (authController AuthController) FindUserFromJWT(jwt string) (*models.Record, error) {
	userRecord, err := authController.AppDao().FindAuthRecordByToken(
		jwt,
		authController.TokenSecret())
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}

// get user from token in context
func (authController AuthController) FindUserFromJWTInContext(c echo.Context) (*models.Record, error) {

	token, err := c.Cookie("jwt")
	if err != nil {
		return nil, err
	}

	user, err := authController.FindUserFromJWT(token.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (authController AuthController) FindUserById(id string) (*users.User, error) {
	u := &users.User{}

	q := authController.AppDao().ModelQuery(&users.User{})

	err := q.AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(u)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (authController AuthController) CheckAuthCredentials(email string, password string, endpoint string) error {

	authCheckRequest, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewReader([]byte(
			`{"identity": "`+email+`", "password": "`+password+`"}`)),
	)
	if err != nil {
		return errors.New("cannot check auth")
	}
	authResp, err := http.DefaultClient.Do(authCheckRequest)
	if err != nil {
		return errors.New("cannot check auth")
	}
	// parse response to json
	authRespJson := map[string]interface{}{}
	err = json.NewDecoder(authResp.Body).Decode(&authRespJson)
	if err != nil {
		return errors.New("error parsing auth response")
	}
	// check if auth was successful
	if authRespJson["verified"] != true {
		return errors.New("auth failed")
	}
	return nil
}
