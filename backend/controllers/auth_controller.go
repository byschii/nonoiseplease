package controllers

import (
	users "be/model/users"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type AuthController struct {
	PBDao       *daos.Dao
	TokenSecret string
}

// get user from token
func (authController AuthController) GetUserFromJWT(jwt string) (*models.Record, error) {

	userRecord, err := authController.PBDao.FindAuthRecordByToken(
		jwt,
		authController.TokenSecret)
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}

// get user from token in context
func (authController AuthController) GetUserFromJWTInContext(c echo.Context) (*models.Record, error) {

	token, err := c.Cookie("jwt")
	if err != nil {
		return nil, err
	}

	user, err := authController.GetUserFromJWT(token.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (authController AuthController) FindUserById(id string) (*users.Users, error) {
	u := &users.Users{}

	q := authController.PBDao.ModelQuery(&users.Users{})

	err := q.AndWhere(dbx.HashExp{"id": id}).
		Limit(1).
		One(u)

	if err != nil {
		return nil, err
	}

	return u, nil
}
