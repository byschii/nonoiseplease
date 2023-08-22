package rest

import (
	users "be/model/users"

	"github.com/pocketbase/dbx"

	"github.com/pocketbase/pocketbase/daos"
)

func StoreNewUserDetails(dao *daos.Dao, nickname string, relatedUserId string) error {
	details := &users.UserDetails{
		Nickname:    nickname,
		RelatedUser: relatedUserId,
	}
	err := dao.Save(details)
	return err
}

func SaveActivity(dao *daos.Dao, activity users.UserActivity) error {
	err := dao.Save(&activity)
	return err
}

func GetUserDetails(dao *daos.Dao, relatedUserId string) (*users.UserDetails, error) {
	details := &users.UserDetails{}
	err := dao.DB().Select("*").From(details.TableName()).Where(dbx.In("related_user", relatedUserId)).One(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

// returns UserImportantData and UserDetails or error
func GetUserPartFromId(dao *daos.Dao, userId string) (*users.UserDetails, error) {

	details, err := GetUserDetails(dao, userId)
	if err != nil {
		return nil, err
	}
	return details, nil
}

func GetUserEmailFromId(dao *daos.Dao, userId string) (string, error) {
	userToFill := &users.Users{}
	err := dao.FindById(userToFill, userId)
	if err != nil {
		return "", err
	}
	return userToFill.Email, nil
}
