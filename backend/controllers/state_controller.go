package controllers

import (
	"be/pkg/config"

	"github.com/pocketbase/pocketbase/daos"
)

type AppStateController struct {
	PBDao *daos.Dao
}

type AppStateControllerInterface interface {
	CommonController
	SetPBDAO(dao *daos.Dao)
	MaxScrapePerMonth() int
	IsGreatWallEnabled() bool
	IsRequireMailVerification() bool
	UseProxy() bool
}

func NewConfigController(dao *daos.Dao) AppStateControllerInterface {
	return &AppStateController{
		PBDao: dao,
	}
}

// AppDao implements AppStateControllerInterface.
func (c AppStateController) AppDao() *daos.Dao {
	return c.PBDao
}

// SetPBDAO implements AppStateControllerInterface.
func (c *AppStateController) SetPBDAO(dao *daos.Dao) {
	c.PBDao = dao
}

func (c AppStateController) MaxScrapePerMonth() int {
	return config.CountMaxScrapePerMonth(c.AppDao())
}

func (c AppStateController) IsGreatWallEnabled() bool {
	return config.IsGreatWallEnabled(c.AppDao())
}

func (c AppStateController) UseProxy() bool {
	return config.UseProxy(c.AppDao())
}

func (c AppStateController) IsRequireMailVerification() bool {
	return config.IsMailVerificationRequired(c.AppDao())
}
