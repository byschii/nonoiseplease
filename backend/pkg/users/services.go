package users

import (
	"github.com/pocketbase/pocketbase/daos"
	"github.com/rs/zerolog/log"
)

// get a list of all users in the database
func List(dao *daos.Dao) ([]*User, error) {
	u := []*User{}
	err := dao.ModelQuery(&User{}).All(&u)
	if err != nil {
		log.Error().Err(err).Msg("error getting user list")
		return nil, err
	}

	return u, nil
}

func FromId(dao *daos.Dao, userId string) (*User, error) {
	userToFill := &User{}
	err := dao.FindById(userToFill, userId)
	if err != nil {
		return nil, err
	}
	return userToFill, nil
}

func EmailFromId(dao *daos.Dao, userId string) (string, error) {
	userToFill, err := FromId(dao, userId)
	if err != nil {
		return "", err
	}
	return userToFill.Email, nil
}
