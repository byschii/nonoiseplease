package users

import (
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// get user from token
func GetUserFromJWT(jwt string, dao *daos.Dao, tokenSecret string) (*models.Record, error) {

	userRecord, err := dao.FindAuthRecordByToken(jwt, tokenSecret)
	if err != nil {
		return nil, err
	}

	return userRecord, nil
}

// get user from token in context
func GetUserFromJWTInContext(c echo.Context, dao *daos.Dao, tokenSecret string) (*models.Record, error) {

	token, err := c.Cookie("jwt")
	if err != nil {
		return nil, err
	}

	user, err := GetUserFromJWT(token.Value, dao, tokenSecret)
	if err != nil {
		return nil, err
	}

	return user, nil
}
