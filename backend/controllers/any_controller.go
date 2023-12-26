package controllers

import "github.com/pocketbase/pocketbase/daos"

type CommonController interface {
	AppDao() *daos.Dao
}
